// name: ssh
// description: SSH and SFTP operations
// author: roturbot
// requires: golang.org/x/crypto/ssh, golang.org/x/crypto/ssh/knownhosts, io, os

type SSH struct{}

type SSHClient struct {
	client *ssh.Client
	session *ssh.Session
	connected bool
}

type SSHAuth struct {
	method string
	config []ssh.AuthMethod
}

func (SSH) connect(host any, port any, user any, password any, privateKey any) *SSHClient {
	hostStr := OSLtoString(host)
	portStr := OSLtoString(port)
	userStr := OSLtoString(user)
	passStr := OSLtoString(password)
	keyData := OSLtoString(privateKey)

	var authMethods []ssh.AuthMethod

	if passStr != "" {
		authMethods = append(authMethods, ssh.Password(passStr))
	}

	if keyData != "" {
		signer, err := ssh.ParsePrivateKey([]byte(keyData))
		if err == nil {
			authMethods = append(authMethods, ssh.PublicKeys(signer))
		}
	}

	if len(authMethods) == 0 {
		return &SSHClient{connected: false}
	}

	if portStr == "" {
		portStr = "22"
	}

	config := &ssh.ClientConfig{
		User: userStr,
		Auth: authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", hostStr+":"+portStr, config)
	if err != nil {
		return &SSHClient{connected: false}
	}

	return &SSHClient{
		client: client,
		connected: true,
	}
}

func (c *SSHClient) exec(command any) map[string]any {
	if !c.connected {
		return map[string]any{"success": false, "error": "not connected"}
	}

	session, err := c.client.NewSession()
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	defer session.Close()

	output, err := session.CombinedOutput(OSLtoString(command))
	success := err == nil
	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
	}

	return map[string]any{
		"success": success,
		"output":  string(output),
		"error":   errorMsg,
	}
}

func (c *SSHClient) startCommand(command any) bool {
	if !c.connected {
		return false
	}

	session, err := c.client.NewSession()
	if err != nil {
		return false
	}

	c.session = session
	return true
}

func (c *SSHClient) sendInput(input any) bool {
	if c.session == nil {
		return false
	}

	_, err := c.session.Write([]byte(OSLtoString(input)))
	return err == nil
}

func (c *SSHClient) readOutput(timeout any) string {
	if c.session == nil {
		return ""
	}

	timeoutDuration := time.Duration(OSLcastNumber(timeout)) * time.Second
	done := make(chan string)

	go func() {
		output := make([]byte, 4096)
		n, _ := c.session.Output(output)
		done <- string(output[:n])
	}()

	select {
	case result := <-done:
		return result
	case <-time.After(timeoutDuration):
		return ""
	}
}

func (c *SSHClient) close() bool {
	if !c.connected {
		return false
	}

	if c.session != nil {
		c.session.Close()
	}
	err := c.client.Close()
	c.connected = false

	return err == nil
}

func (c *SSHClient) isConnected() bool {
	return c.connected
}

func (SSH) execRemote(host any, port any, user any, password any, privateKey any, command any) map[string]any {
	client := ssh.connect(host, port, user, password, privateKey)
	result := client.exec(command)
	client.close()
	return result
}

func (SSH) scpUpload(client *SSHClient, localPath any, remotePath any) bool {
	if !client.connected {
		return false
	}

	localFile, err := os.Open(OSLtoString(localPath))
	if err != nil {
		return false
	}
	defer localFile.Close()

	stat, err := localFile.Stat()
	if err != nil {
		return false
	}

	session, err := client.client.NewSession()
	if err != nil {
		return false
	}
	defer session.Close()

	go func() {
		var wc io.WriteCloser
		wc, err = session.StdinPipe()
		if err != nil {
			return
		}
		defer wc.Close()

		fmt.Fprintln(wc, "C0755", stat.Size(), filepath.Base(OSLtoString(localPath)))
		io.Copy(wc, localFile)
		fmt.Fprint(wc, "\x00")
	}()

	err = session.Run("scp -t " + OSLtoString(remotePath))
	return err == nil
}

func (SSH) scpDownload(client *SSHClient, remotePath any, localPath any) bool {
	if !client.connected {
		return false
	}

	session, err := client.client.NewSession()
	if err != nil {
		return false
	}
	defer session.Close()

	filename := filepath.Base(OSLtoString(localPath))
	remoteFile := filepath.Join(OSLtoString(remotePath), filename)

	localFile, err := os.Create(OSLtoString(localPath))
	if err != nil {
		return false
	}
	defer localFile.Close()

	go func() {
		wc, _ := session.StdoutPipe()
		io.Copy(localFile, wc)
	}()

	err = session.Run("scp -f " + remoteFile)
	return err == nil
}

func (SSH) tunnel(localPort any, remoteHost any, remotePort any, sshHost any, sshPort any) *SSHClient {
	sshHostStr := OSLtoString(sshHost)
	sshPortStr := OSLtoString(sshPort)
	remoteHostStr := OSLtoString(remoteHost)
	remotePortStr := OSLtoString(remotePort)
	localPortInt := OSLcastInt(localPort)

	authMethods := []ssh.AuthMethod{ssh.Password("")}
	config := &ssh.ClientConfig{
		User:            "root",
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if sshPortStr == "" {
		sshPortStr = "22"
	}

	sshClient, err := ssh.Dial("tcp", sshHostStr+":"+sshPortStr, config)
	if err != nil {
		return nil
	}

	listener, err := sshClient.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", localPortInt))
	if err != nil {
		sshClient.Close()
		return nil
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				break
			}

			remoteConn, err := sshClient.Dial("tcp", remoteHostStr+":"+remotePortStr)
			if err != nil {
				conn.Close()
				continue
			}

			go func() {
				defer conn.Close()
				defer remoteConn.Close()
				io.Copy(remoteConn, conn)
				io.Copy(conn, remoteConn)
			}()
		}
	}()

	return &SSHClient{
		client: sshClient,
		connected: true,
	}
}

func (SSH) generateKeyPair(keyType any) map[string]any {
	typeStr := strings.ToLower(OSLtoString(keyType))

	var privateKey string
	var publicKey string

	if typeStr == "rsa" || typeStr == "" {
		privateKey, publicKey = ssh.generateRSAKey()
	} else if typeStr == "ed25519" {
		privateKey, publicKey = ssh.generateEd25519Key()
	}

	if privateKey == "" {
		return map[string]any{}
	}

	return map[string]any{
		"private": privateKey,
		"public":  publicKey,
		"type":     typeStr,
	}
}

func (s SSH) generateRSAKey() (string, string) {
	privateKey, err := crypto.generateRSAKey()
	if err != nil {
		return "", ""
	}

	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return "", ""
	}

	publicKey, err := ssh.NewPublicKey(&signer.(ssh.Signer).PublicKey())
	if err != nil {
		return "", ""
	}

	return privateKey, string(ssh.MarshalAuthorizedKey(publicKey))
}

func (s SSH) generateEd25519Key() (string, string) {
	privateKey, err := crypto.generateEd25519Key()
	if err != nil {
		return "", ""
	}

	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return "", ""
	}

	publicKey, err := ssh.NewPublicKey(&signer.(ssh.Signer).PublicKey())
	if err != nil {
		return "", ""
	}

	return privateKey, string(ssh.MarshalAuthorizedKey(publicKey))
}

func (SSH) savePrivateKey(path any, key any) bool {
	privateKey := OSLtoString(key)
	return fs.WriteFile(path, privateKey)
}

func (SSH) savePublicKey(path any, key any) bool {
	publicKey := OSLtoString(key)
	return fs.WriteFile(path, publicKey)
}

func (SSH) loadPrivateKey(path any) string {
	return fs.ReadFile(path)
}

func (s SSH) fingerprint(publicKey any) string {
	pubKeyStr := OSLtoString(publicKey)
	pub, err := ssh.ParseAuthorizedKey([]byte(pubKeyStr))
	if err != nil {
		return ""
	}

	return ssh.FingerprintSHA256(pub)
}

var ssh = SSH{}
