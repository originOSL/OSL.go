// name: ws
// description: websocket package for osl
// author: Mist
// requires: github.com/gorilla/websocket, crypto/tls, log, sync, net/http, time

const (
	// connSendBuffer is the number of outbound messages buffered per connection
	// before drops occur. 256 handles bursts without consuming excessive memory.
	connSendBuffer = 256

	// connWorkBuffer is the number of inbound messages queued for the worker pool.
	connWorkBuffer = 256

	// connWorkers is the number of concurrent message handlers per connection.
	connWorkers = 4
)

// jsonPool reuses byte buffers for JSON serialisation to reduce allocations
// on hot send paths.
var jsonPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 512)
		return &b
	},
}

type Connection struct {
	url       string
	conn      *websocket.Conn
	send      chan []byte
	work      chan string
	closeOnce sync.Once
	closed    bool
	closeMu   sync.RWMutex

	reconnect bool // set by EnableReconnect()

	data sync.Map // replaces dataMutex + map — lock-free reads/writes

	onMessage func(*Connection, string)
	onClose   func(*Connection)
}

type WS struct{}

func (WS) Connect(url string, protocols ...string) *Connection {
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	headers := http.Header{}
	if len(protocols) > 0 {
		headers["Sec-WebSocket-Protocol"] = protocols
	}

	conn, _, err := dialer.Dial(url, headers)
	if err != nil {
		return nil
	}

	c := &Connection{
		url:  url,
		conn: conn,
		send: make(chan []byte, connSendBuffer),
		work: make(chan string, connWorkBuffer),
	}

	c.startWorkers()
	go c.readLoop()
	go c.writeLoop()

	return c
}

type Server struct {
	addr     string
	path     string
	upgrader websocket.Upgrader
	conns    sync.Map // map[*Connection]struct{} — lock-free connection tracking
	server   *http.Server

	onConnect    func(*Connection)
	onMessage    func(*Connection, string)
	onDisconnect func(*Connection)
}

func (WS) NewServer(addr, path string) *Server {
	return &Server{
		addr: addr,
		path: path,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (s *Server) OnConnect(handler func(*Connection)) {
	s.onConnect = handler
}

func (s *Server) OnMessage(handler func(*Connection, string)) {
	s.onMessage = handler
}

func (s *Server) OnDisconnect(handler func(*Connection)) {
	s.onDisconnect = handler
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ws: upgrade error:", err)
		return
	}

	c := &Connection{
		url:  r.RemoteAddr,
		conn: conn,
		send: make(chan []byte, connSendBuffer),
		work: make(chan string, connWorkBuffer),
	}

	c.onMessage = func(c *Connection, msg string) {
		if s.onMessage != nil {
			s.onMessage(c, msg)
		}
	}

	c.onClose = func(c *Connection) {
		s.conns.Delete(c)
		if s.onDisconnect != nil {
			s.onDisconnect(c)
		}
	}

	s.conns.Store(c, struct{}{})

	if s.onConnect != nil {
		s.onConnect(c)
	}

	c.startWorkers()
	go c.readLoop()
	go c.writeLoop()
}

// Broadcast sends a message to every connected client. The connection list is
// snapshotted without holding any lock during the actual sends, so a slow or
// closing connection cannot stall other recipients.
func (s *Server) Broadcast(message string) {
	data := []byte(message)
	s.conns.Range(func(key, _ any) bool {
		conn := key.(*Connection)
		conn.sendBytes(data)
		return true
	})
}

func (s *Server) GetConnections() []*Connection {
	var conns []*Connection
	s.conns.Range(func(key, _ any) bool {
		conns = append(conns, key.(*Connection))
		return true
	})
	return conns
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc(s.path, s.handleWebSocket)

	s.server = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	log.Printf("ws: server starting on %s%s", s.addr, s.path)
	return s.server.ListenAndServe()
}

func (s *Server) StartTLS(certFile, keyFile string) error {
	mux := http.NewServeMux()
	mux.HandleFunc(s.path, s.handleWebSocket)

	s.server = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	log.Printf("ws: server (TLS) starting on %s%s", s.addr, s.path)
	return s.server.ListenAndServeTLS(certFile, keyFile)
}

func (s *Server) Stop() error {
	if s.server == nil {
		return nil
	}

	// Close all connections. sync.Map.Range is safe to call concurrently with
	// Delete (which onClose will call), so no lock needed here.
	s.conns.Range(func(key, _ any) bool {
		key.(*Connection).Close()
		return true
	})

	return s.server.Close()
}

// Connection methods (shared by client and server)

// Send serialises and enqueues a message. It is safe to call from multiple
// goroutines concurrently.
func (c *Connection) Send(message any) {
	var data []byte
	switch v := message.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	case map[string]any:
		buf := jsonPool.Get().(*[]byte)
		*buf = []byte(JsonStringify(v))
		data = append([]byte(nil), *buf...) // copy before returning to pool
		*buf = (*buf)[:0]
		jsonPool.Put(buf)
	case []any:
		buf := jsonPool.Get().(*[]byte)
		*buf = []byte(JsonStringify(v))
		data = append([]byte(nil), *buf...)
		*buf = (*buf)[:0]
		jsonPool.Put(buf)
	default:
		log.Println("ws: invalid message type:", reflect.TypeOf(message).String())
		return
	}
	c.sendBytes(data)
}

// sendBytes is the internal hot path. Holding closeMu.RLock for the full
// duration closes the race window between the closed-check and channel write.
func (c *Connection) sendBytes(data []byte) {
	c.closeMu.RLock()
	defer c.closeMu.RUnlock()

	if c.closed {
		return
	}

	select {
	case c.send <- data:
	default:
		log.Println("ws: send buffer full, dropping message")
	}
}

func (c *Connection) Close() {
	c.closeOnce.Do(func() {
		c.closeMu.Lock()
		c.closed = true
		close(c.send)
		c.closeMu.Unlock()

		close(c.work)

		c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.conn.Close()

		if c.reconnect {
			go c.reconnectLoop()
			return
		}
		if c.onClose != nil {
			go c.onClose(c)
		}
	})
}

// EnableReconnect makes the connection automatically reconnect with exponential
// backoff (1s→2s→4s…30s cap) whenever the server drops it. The same
// *Connection handle stays valid — OnMessage/OnClose only need to be set once.
func (c *Connection) EnableReconnect() {
	c.reconnect = true
}

func (c *Connection) reconnectLoop() {
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	delay := time.Second
	for {
		time.Sleep(delay)
		log.Printf("ws: reconnecting to %s", c.url)
		conn, _, err := dialer.Dial(c.url, nil)
		if err != nil {
			log.Printf("ws: reconnect failed: %v (retry in %v)", err, delay)
			delay *= 2
			if delay > 30*time.Second {
				delay = 30 * time.Second
			}
			continue
		}
		// reset state for the new underlying connection
		c.closeMu.Lock()
		c.conn = conn
		c.send = make(chan []byte, connSendBuffer)
		c.work = make(chan string, connWorkBuffer)
		c.closed = false
		c.closeOnce = sync.Once{}
		c.closeMu.Unlock()

		delay = time.Second // reset backoff on success
		log.Printf("ws: reconnected to %s", c.url)

		c.startWorkers()
		go c.readLoop()
		go c.writeLoop()
		return
	}
}

// Set stores an arbitrary value on the connection. Uses sync.Map for
// lock-free concurrent access.
func (c *Connection) Set(key string, value any) {
	c.data.Store(key, value)
}

func (c *Connection) Get(key string) any {
	val, _ := c.data.Load(key)
	return val
}

func (c *Connection) Delete(key string) {
	c.data.Delete(key)
}

func (c *Connection) GetAll() map[string]any {
	result := make(map[string]any)
	c.data.Range(func(k, v any) bool {
		result[k.(string)] = v
		return true
	})
	return result
}

func (c *Connection) OnMessage(handler func(*Connection, string)) {
	c.onMessage = handler
	// Drain any messages that arrived before the handler was registered.
	for {
		select {
		case msg, ok := <-c.work:
			if !ok {
				return
			}
			go handler(c, msg)
		default:
			return
		}
	}
}

func (c *Connection) OnClose(handler func(*Connection)) {
	c.onClose = handler
}

// startWorkers launches a fixed pool of goroutines that drain the work channel.
// This bounds concurrency for message handling instead of spawning an unbounded
// number of goroutines (one per message).
func (c *Connection) startWorkers() {
	for i := 0; i < connWorkers; i++ {
		go func() {
			for msg := range c.work {
				if c.onMessage != nil {
					c.onMessage(c, msg)
				}
			}
		}()
	}
}

func (c *Connection) readLoop() {
	defer c.Close()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("ws: read error:", err)
			return
		}
		select {
		case c.work <- string(message):
		default:
			log.Println("ws: work queue full, dropping message")
		}
	}
}

func (c *Connection) writeLoop() {
	defer c.Close()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Println("ws: write error:", err)
			return
		}
	}
}

var ws = WS{}