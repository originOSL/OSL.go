// name: net
// description: Network utilities (TCP, UDP, DNS, HTTP client)
// author: roturbot
// requires: net, net/http, net/url, time, io

type Net struct{}

type TCPConn struct {
	conn net.Conn
}

type UDPConn struct {
	conn *net.UDPConn
	addr *net.UDPAddr
}

func (Net) dial(network any, address any) *TCPConn {
	networkStr := OSLtoString(network)
	addressStr := OSLtoString(address)

	conn, err := net.Dial(networkStr, addressStr)
	if err != nil {
		return nil
	}

	return &TCPConn{conn: conn}
}

func (Net) listen(protocol any, address any) *TCPConn {
	protocolStr := OSLtoString(protocol)
	addressStr := OSLtoString(address)

	listener, err := net.Listen(protocolStr, addressStr)
	if err != nil {
		return nil
	}

	return &TCPConn{conn: listener}
}

func (Net) listenUDP(network any, address any) *UDPConn {
	networkStr := OSLtoString(network)
	addressStr := OSLtoString(address)

	addr, err := net.ResolveUDPAddr(networkStr, addressStr)
	if err != nil {
		return nil
	}

	conn, err := net.ListenUDP(networkStr, addr)
	if err != nil {
		return nil
	}

	return &UDPConn{conn: conn, addr: addr}
}

func (c *TCPConn) write(data any) bool {
	if c == nil || c.conn == nil {
		return false
	}

	dataStr := OSLtoString(data)
	_, err := c.conn.Write([]byte(dataStr))
	return err == nil
}

func (c *TCPConn) writeBytes(data []byte) bool {
	if c == nil || c.conn == nil {
		return false
	}

	_, err := c.conn.Write(data)
	return err == nil
}

func (c *TCPConn) read(bufferSize any) string {
	if c == nil || c.conn == nil {
		return ""
	}

	size := OSLcastInt(bufferSize)
	if size <= 0 {
		size = 4096
	}

	buf := make([]byte, size)
	n, err := c.conn.Read(buf)
	if err != nil {
		return ""
	}

	return string(buf[:n])
}

func (c *TCPConn) readBytes(bufferSize any) []byte {
	if c == nil || c.conn == nil {
		return []byte{}
	}

	size := OSLcastInt(bufferSize)
	if size <= 0 {
		size = 4096
	}

	buf := make([]byte, size)
	n, err := c.conn.Read(buf)
	if err != nil {
		return []byte{}
	}

	return buf[:n]
}

func (c *TCPConn) close() bool {
	if c == nil || c.conn == nil {
		return false
	}

	err := c.conn.Close()
	return err == nil
}

func (c *TCPConn) remoteAddr() string {
	if c == nil || c.conn == nil {
		return ""
	}
	return c.conn.RemoteAddr().String()
}

func (c *TCPConn) localAddr() string {
	if c == nil || c.conn == nil {
		return ""
	}
	return c.conn.LocalAddr().String()
}

func (c *TCPConn) setTimeout(seconds any) bool {
	if c == nil || c.conn == nil {
		return false
	}

	duration := time.Duration(OSLcastNumber(seconds)) * time.Second
	err := c.conn.SetDeadline(time.Now().Add(duration))
	return err == nil
}

func (c *UDPConn) write(data any, targetAddress any) bool {
	if c == nil || c.conn == nil {
		return false
	}

	targetStr := OSLtoString(targetAddress)
	addr, err := net.ResolveUDPAddr("udp", targetStr)
	if err != nil {
		return false
	}

	_, err = c.conn.WriteTo([]byte(OSLtoString(data)), addr)
	return err == nil
}

func (c *UDPConn) read(bufferSize any) map[string]any {
	if c == nil || c.conn == nil {
		return map[string]any{"data": "", "addr": ""}
	}

	size := OSLcastInt(bufferSize)
	if size <= 0 {
		size = 4096
	}

	buf := make([]byte, size)
	n, addr, err := c.conn.ReadFromUDP(buf)
	if err != nil {
		return map[string]any{"data": "", "addr": ""}
	}

	return map[string]any{
		"data": string(buf[:n]),
		"addr": addr.String(),
	}
}

func (c *UDPConn) close() bool {
	if c == nil || c.conn == nil {
		return false
	}

	err := c.conn.Close()
	return err == nil
}

func (Net) httpGet(url any) string {
	urlStr := OSLtoString(url)
	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Get(urlStr)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	return string(body)
}

func (Net) httpPost(url any, data any) string {
	urlStr := OSLtoString(url)
	dataStr := OSLtoString(data)
	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Post(urlStr, "application/json", strings.NewReader(dataStr))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	return string(body)
}

func (Net) httpRequest(method any, url any, headers any, body any) map[string]any {
	methodStr := strings.ToUpper(OSLtoString(method))
	urlStr := OSLtoString(url)
	bodyStr := OSLtoString(body)
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest(methodStr, urlStr, strings.NewReader(bodyStr))
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}

	headersMap := OSLcastObject(headers)
	for k, v := range headersMap {
		req.Header.Set(k, OSLtoString(v))
	}

	resp, err := client.Do(req)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}

	responseHeaders := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			responseHeaders[k] = v[0]
		}
	}

	return map[string]any{
		"success": true,
		"status":  resp.StatusCode,
		"body":    string(responseBody),
		"headers": responseHeaders,
	}
}

func (Net) urlParse(url any) map[string]any {
	urlStr := OSLtoString(url)
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}

	return map[string]any{
		"scheme":   parsed.Scheme,
		"host":     parsed.Host,
		"hostname": parsed.Hostname(),
		"port":     parsed.Port(),
		"path":     parsed.Path,
		"query":    parsed.RawQuery,
		"fragment": parsed.Fragment,
	}
}

func (Net) urlEncode(data any) string {
	return url.QueryEscape(OSLtoString(data))
}

func (Net) urlDecode(data any) string {
	decoded, err := url.QueryUnescape(OSLtoString(data))
	if err != nil {
		return ""
	}
	return decoded
}

func (Net) lookupHost(hostname any) []any {
	hostnameStr := OSLtoString(hostname)
	addrs, err := net.LookupHost(hostnameStr)
	if err != nil {
		return []any{}
	}

	result := make([]any, len(addrs))
	for i, addr := range addrs {
		result[i] = addr
	}
	return result
}

func (Net) lookupIP(hostname any) []any {
	hostnameStr := OSLtoString(hostname)
	addrs, err := net.LookupIP(hostnameStr)
	if err != nil {
		return []any{}
	}

	result := make([]any, len(addrs))
	for i, addr := range addrs {
		result[i] = addr.String()
	}
	return result
}

func (Net) lookupPort(service any, network any) int {
	serviceStr := OSLtoString(service)
	networkStr := OSLtoString(network)
	port, err := net.LookupPort(networkStr, serviceStr)
	if err != nil {
		return 0
	}
	return port
}

func (n Net) getAddressInfo(hostname any) map[string]any {
	hostnameStr := OSLtoString(hostname)

	host := n.lookupHost(hostnameStr)
	ips := n.lookupIP(hostnameStr)

	return map[string]any{
		"hostname": hostnameStr,
		"hosts":    host,
		"ips":      ips,
	}
}

var net = Net{}
