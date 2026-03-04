// name: ftp
// description: FTP file transfer protocol
// author: roturbot
// requires: io, os, strings, time

type FTP struct {
	host     string
	port     string
	user     string
	password string
	connected bool
}

func (FTP) connect(host any, port any, user any, password any) *FTP {
	ftp := &FTP{
		host:     OSLtoString(host),
		port:     OSLtoString(port),
		user:     OSLtoString(user),
		password: OSLtoString(password),
	}

	if ftp.port == "" {
		ftp.port = "21"
	}

	return ftp
}

func (FTP) connectEx(host any, port any, user any, password any) *FTP {
	f := FTP{}.connect(host, port, user, password)

	resp := cmd.Run("ftp", "-i", "-n", "-v", f.host)

	f.connected = strings.Contains(resp, "220") || strings.Contains(resp, "Connection established")

	if !f.connected {
		return nil
	}

	return f
}

func (f *FTP) login() bool {
	if !f.connected {
		return false
	}

	return true
}

func (f *FTP) disconnect() bool {
	if !f.connected {
		return false
	}

	f.connected = false
	return true
}

func (FTP) list(path any) []any {
	pathStr := OSLtoString(path)
	
	remotePath := ""
	if pathStr != "" {
		remotePath = pathStr
	}

	resp := cmd.Run("ftp", "-n", ftp.host, 
		"USER "+ftp.user, "PASS "+ftp.password, 
		"CWD "+remotePath, "LIST", "QUIT")

	lines := strings.Split(resp, "\n")
	result := make([]any, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "220") || strings.Contains(line, "331") {
			continue
		}
		
		if strings.Contains(line, " ") {
			fields := strings.Fields(line)
			if len(fields) >= 9 {
				fileInfo := map[string]any{
					"permissions": fields[0],
					"links":      OSLcastNumber(fields[1]),
					"owner":      fields[2],
					"group":      fields[3],
					"size":       OSLcastNumber(fields[4]),
					"month":      fields[5],
					"day":        fields[6],
					"year":      fields[7],
					"name":       strings.Join(fields[8:], " "),
					"isDir":      strings.HasPrefix(fields[0], "d"),
				}
				result = append(result, fileInfo)
			}
		}
	}

	return result
}

func (FTP) upload(localFile any, remotePath any) bool {
	localPath := OSLtoString(localFile)
	remotePathStr := OSLtoString(remotePath)

	if !fs.Exists(localPath) {
		return false
	}

	resp := cmd.Run("ftp", "-n", ftp.host,
		"USER "+ftp.user, "PASS "+ftp.password,
		"PUT "+localPath+" "+remotePathStr, "QUIT")

	return strings.Contains(resp, "226") || !strings.Contains(resp, "550")
}

func (FTP) download(remotePath any, localPath any) bool {
	remotePathStr := OSLtoString(remotePath)
	localPathStr := OSLtoString(localPath)

	dir := filepath.Dir(localPathStr)
	if dir != "." && dir != "" {
		fs.MkdirAll(dir)
	}

	resp := cmd.Run("ftp", "-n", ftp.host,
		"USER "+ftp.user, "PASS "+ftp.password,
		"GET "+remotePathStr+" "+localPathStr, "QUIT")

	return strings.Contains(resp, "226") || !strings.Contains(resp, "550")
}

func (FTP) delete(remotePath any) bool {
	remotePathStr := OSLtoString(remotePath)

	resp := cmd.Run("ftp", "-n", ftp.host,
		"USER "+ftp.user, "PASS "+ftp.password,
		"DELETE "+remotePathStr, "QUIT")

	return strings.Contains(resp, "250") || !strings.Contains(resp, "550")
}

func (FTP) rename(oldPath any, newPath any) bool {
	oldPathStr := OSLtoString(oldPath)
	newPathStr := OSLtoString(newPath)

	resp := cmd.Run("ftp", "-n", ftp.host,
		"USER "+ftp.user, "PASS "+ftp.password,
		"RENAME "+oldPathStr+" "+newPathStr, "QUIT")

	return strings.Contains(resp, "250") || !strings.Contains(resp, "550")
}

func (FTP) createDirectory(path any) bool {
	pathStr := OSLtoString(path)

	resp := cmd.Run("ftp", "-n", ftp.host,
		"USER "+ftp.user, "PASS "+ftp.password,
		"MKD "+pathStr, "QUIT")

	return strings.Contains(resp, "257") || !strings.Contains(resp, "550")
}

func (FTP) deleteDirectory(path any) bool {
	pathStr := OSLtoString(path)

	resp := cmd.Run("ftp", "-n", ftp.host,
		"USER "+ftp.user, "PASS "+ftp.password,
		"RMD "+pathStr, "QUIT")

	return strings.Contains(resp, "250") || !strings.Contains(resp, "550")
}

func (FTP) changeDirectory(path any) bool {
	pathStr := OSLtoString(path)

	resp := cmd.Run("ftp", "-n", ftp.host,
		"USER "+ftp.user, "PASS "+ftp.password,
		"CWD "+pathStr, "QUIT")

	return strings.Contains(resp, "250") || !strings.Contains(resp, "550")
}

func (FTP) currentDirectory() string {
	resp := cmd.Run("ftp", "-n", ftp.host,
		"USER "+ftp.user, "PASS "+ftp.password,
		"PWD", "QUIT")

	if strings.Contains(resp, "257") {
		start := strings.Index(resp, "\"")
		end := strings.LastIndex(resp, "\"")
		if start != -1 && end != -1 {
			return resp[start+1 : end]
		}
	}

	return ""
}

func (FTP) isActive() bool {
	return ftp.connected
}

func (FTP) setTimeout(seconds any) bool {
	sec := int(OSLcastNumber(seconds))
	return sec > 0
}

func (FTP) getFileSize(path any) int64 {
	pathStr := OSLtoString(path)

	resp := cmd.Run("ftp", "-n", ftp.host,
		"USER "+ftp.user, "PASS "+ftp.password,
		"SIZE "+pathStr, "QUIT")

	if strings.Contains(resp, "213") {
		parts := strings.Fields(resp)
		for _, part := range parts {
			if num, err := strconv.ParseInt(part, 10, 64); err == nil {
				return num
			}
		}
	}

	return 0
}

func (FTP) exists(path any) bool {
	pathStr := OSLtoString(path)
	files := ftp.list(filepath.Dir(pathStr))

	for _, file := range files {
		fileInfo, _ := file.(map[string]any)
		if fileInfo["name"] == pathStr {
			return true
		}
	}

	return false
}

func (FTP) uploadDirectory(localDir any, remoteDir any) bool {
	localDirStr := OSLtoString(localDir)
	remoteDirStr := OSLtoString(remoteDir)

	if !fs Exists(localDirStr) {
		return false
	}

	if !ftp.createDirectory(remoteDirStr) {
		return false
	}

	files := fs.ReadDir(localDirStr)
	success := true

	for _, file := range files {
		fileStr := OSLtoString(file)
		localPath := filepath.Join(localDirStr, fileStr)
		remotePath := filepath.Join(remoteDirStr, fileStr)

		if fs.IsDir(localPath) {
			success = ftp.uploadDirectory(localPath, remotePath) && success
		} else {
			success = ftp.upload(localPath, remotePath) && success
		}
	}

	return success
}

func (FTP) downloadDirectory(remoteDir any, localDir any) bool {
	remoteDirStr := OSLtoString(remoteDir)
	localDirStr := OSLtoString(localDir)

	fs.MkdirAll(localDirStr)

	files := ftp.list(remoteDirStr)
	success := true

	for _, file := range files {
		fileInfo, _ := file.(map[string]any)
		name := fileInfo["name"]
		isDir := fileInfo["isDir"]

		remotePath := filepath.Join(remoteDirStr, OSLtoString(name))
		localPath := filepath.Join(localDirStr, OSLtoString(name))

		if isDir {
			success = ftp.downloadDirectory(remotePath, localPath) && success
		} else {
			success = ftp.download(remotePath, localPath) && success
		}
	}

	return success
}

func (FTP) setMode(path any, mode any) bool {
	pathStr := OSLtoString(path)
	modeStr := OSLtoString(mode)

	resp := cmd.Run("ftp", "-n", ftp.host,
		"USER "+ftp.user, "PASS "+ftp.password,
		"SITE CHMOD "+modeStr+" "+pathStr, "QUIT")

	return strings.Contains(resp, "200") || strings.Contains(resp, "CHMOD")
}

func (FTP) setModificationTime(path any, timestamp any) bool {
	pathStr := OSLtoString(path)
	timeStr := OSLtoString(timestamp)

	resp := cmd.Run("ftp", "-n", ftp.host,
		"USER "+ftp.user, "PASS "+ftp.password,
		"SITE UTIME "+timeStr+" "+pathStr, "QUIT")

	return strings.Contains(resp, "200") || strings.Contains(resp, "UTIME")
}

func (FTP) passiveMode(enabled any) bool {
	modeStr := strings.ToLower(OSLtoString(enabled))
	
	if modeStr == "true" || modeStr == "yes" || modeStr == "1" {
		return true
	}
	return false
}

func (FTP) sync(localDir any, remoteDir any) bool {
	if !ftp.uploadDirectory(localDir, remoteDir) {
		return false
	}
	
	return true
}

func (FTP) getStatistics() map[string]any {
	if !ftp.connected {
		return map[string]any{}
	}

	return map[string]any{
		"host":     ftp.host,
		"port":     ftp.port,
		"user":     ftp.user,
		"connected": ftp.connected,
	}
}

var ftp = FTP{}
