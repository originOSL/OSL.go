// name: ws
// description: websocket package for osl
// author: Mist
// requires: github.com/gorilla/websocket, crypto/tls, log, sync

type Connection struct {
	url       string
	conn      *websocket.Conn
	send      chan []byte
	closeOnce sync.Once

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
		send: make(chan []byte, 10),
	}

	go c.readLoop()
	go c.writeLoop()

	return c
}

func (c *Connection) Send(message string) {
	select {
	case c.send <- []byte(message):
	default:
		log.Println("Send channel full, dropping message")
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