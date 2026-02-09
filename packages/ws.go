// name: ws
// description: websocket package for osl
// author: Mist
// requires: github.com/gorilla/websocket, crypto/tls, log, sync, net/http

type Connection struct {
	url       string
	conn      *websocket.Conn
	send      chan []byte
	closeOnce sync.Once

	dataMutex sync.RWMutex
	data      map[string]any

	onMessage func(string)
	onClose   func()
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
		data: map[string]any{},
		send: make(chan []byte, 10),
	}

	go c.readLoop()
	go c.writeLoop()

	return c
}

type Server struct {
	addr        string
	path        string
	upgrader    websocket.Upgrader
	connections map[*Connection]bool
	connMutex   sync.RWMutex
	server      *http.Server

	onConnect    func(*Connection)
	onMessage    func(*Connection, string)
	onDisconnect func(*Connection)
}

func (WS) NewServer(addr, path string) *Server {
	return &Server{
		addr:        addr,
		path:        path,
		connections: make(map[*Connection]bool),
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
		log.Println("upgrade error:", err)
		return
	}

	c := &Connection{
		url:  r.RemoteAddr,
		conn: conn,
		data: map[string]any{},
		send: make(chan []byte, 10),
	}

	c.onMessage = func(msg string) {
		if s.onMessage != nil {
			s.onMessage(c, msg)
		}
	}

	c.onClose = func() {
		s.removeConnection(c)
		if s.onDisconnect != nil {
			s.onDisconnect(c)
		}
	}

	s.addConnection(c)

	if s.onConnect != nil {
		s.onConnect(c)
	}

	go c.readLoop()
	go c.writeLoop()
}

func (s *Server) addConnection(c *Connection) {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()
	s.connections[c] = true
}

func (s *Server) removeConnection(c *Connection) {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()
	delete(s.connections, c)
}

func (s *Server) Broadcast(message string) {
	s.connMutex.RLock()
	defer s.connMutex.RUnlock()

	for conn := range s.connections {
		conn.Send(message)
	}
}

func (s *Server) GetConnections() []*Connection {
	s.connMutex.RLock()
	defer s.connMutex.RUnlock()

	conns := make([]*Connection, 0, len(s.connections))
	for conn := range s.connections {
		conns = append(conns, conn)
	}
	return conns
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc(s.path, s.handleWebSocket)

	s.server = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	log.Printf("WebSocket server starting on %s%s", s.addr, s.path)
	return s.server.ListenAndServe()
}

func (s *Server) StartTLS(certFile, keyFile string) error {
	mux := http.NewServeMux()
	mux.HandleFunc(s.path, s.handleWebSocket)

	s.server = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	log.Printf("WebSocket server (TLS) starting on %s%s", s.addr, s.path)
	return s.server.ListenAndServeTLS(certFile, keyFile)
}

func (s *Server) Stop() error {
	if s.server == nil {
		return nil
	}

	// Close all connections
	s.connMutex.Lock()
	for conn := range s.connections {
		conn.Close()
	}
	s.connMutex.Unlock()

	return s.server.Close()
}

// Connection methods (shared by client and server)

func (c *Connection) Send(message any) {
	switch v := message.(type) {
	case string:
		c.send <- []byte(v)
	case map[string]any:
		c.send <- []byte(JsonStringify(v))
	case []any:
		c.send <- []byte(JsonStringify(v))
	default:
		panic("Invalid message type: " + reflect.TypeOf(message).String())
	}
}

func (c *Connection) Close() {
	c.closeOnce.Do(func() {
		close(c.send)
		c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.conn.Close()
		if c.onClose != nil {
			c.onClose()
		}
	})
}

func (c *Connection) Set(key string, value any) {
	c.dataMutex.Lock()
	defer c.dataMutex.Unlock()
	c.data[key] = value
}

func (c *Connection) Get(key string) any {
	c.dataMutex.RLock()
	defer c.dataMutex.RUnlock()
	val, ok := c.data[key]
	if !ok {
		return nil
	}
	return val
}

func (c *Connection) Delete(key string) {
	c.dataMutex.Lock()
	defer c.dataMutex.Unlock()
	delete(c.data, key)
}

func (c *Connection) GetAll() map[string]any {
	c.dataMutex.RLock()
	defer c.dataMutex.RUnlock()
	copy := make(map[string]any, len(c.data))
	for k, v := range c.data {
		copy[k] = v
	}
	return copy
}

func (c *Connection) OnMessage(handler func(string)) {
	c.onMessage = handler
}

func (c *Connection) OnClose(handler func()) {
	c.onClose = handler
}

func (c *Connection) readLoop() {
	defer c.Close()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			return
		}
		if c.onMessage != nil {
			c.onMessage(string(message))
		}
	}
}

func (c *Connection) writeLoop() {
	for msg := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println("write error:", err)
			return
		}
	}
}

var ws = WS{}