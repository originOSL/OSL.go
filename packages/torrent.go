// name: torrent
// description: BitTorrent file sharing utilities
// author: roturbot
// requires: encoding/binary, os, strings, path/filepath

type Torrent struct {
	info     map[string]any
	pieces   []string
	files    []map[string]any
	metadata map[string]string
}

func (t *Torrent) create(name any, files []any, pieceLength any) *Torrent {
	nameStr := OSLtoString(name)
	pieceLen := int(OSLcastNumber(pieceLength))
	if pieceLen <= 0 {
		pieceLen = 262144
	}

	t := &Torrent{
		info: map[string]any{
			"name":         nameStr,
			"piece_length": pieceLen,
			"files":        files,
		},
		metadata: map[string]string{
			"created_by": "OSL Torrent Client",
			"created_on": time.Now().Format(time.RFC3339),
			},
	}

	t.files = make([]map[string]any, len(files))
	for i, file := range files {
		if fileMap, ok := file.(map[string]any); ok {
			t.files[i] = fileMap
		} else {
			filePath := OSLtoString(file)
			fileInfo, _ := os.Stat(filePath)

			t.files[i] = map[string]any{
				"path":   filePath,
				"length": fileInfo.Size(),
				"md5sum": t.calculateMD5(filePath),
			}
		}
	}

	return t
}

func (Torrent) createFromDirectory(dirPath any) *Torrent {
	dirStr := OSLtoString(dirPath)
	
	var tt Torrent
	files := tt.scanDirectory(dirStr)
	totalSize := int64(0)
	
	for _, file := range files {
		tt.info["total_size"] = totalSize
	}

	return Torrent.create(filepath.Base(dirStr), files, 524288)
}

func (t *Torrent) scanDirectory(dirPath string) []any {
	var files []any

	filepath.WalkDir(dirPath, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		fileInfo, _ := entry.Info()

		files = append(files, map[string]any{
			"path":   path,
			"length": fileInfo.Size(),
			"md5sum": "",
		})

		return nil
	})

	return files
}

func (t *Torrent) calculateMD5(filePath string) string {
	data := fs.ReadFile(filePath)
	if data == "" {
		return ""
	}

	return crypto.md5Hash(data)
}

func (t *Torrent) save(path any) bool {
	pathStr := OSLtoString(path)

	torrentData := t.buildTorrent()
	return fs.WriteFile(path, torrentData)
}

func (t *Torrent) buildTorrent() string {
	var torrent strings.Builder

	torrent.WriteString("d8:announce7:http://localhost:6881/announce")
	torrent.WriteString("13:creation date20:" + time.Now().Unix() + "e")
	torrent.WriteString("4:info")

	torrent.WriteString("d")
	torrent.WriteString("5:name" + OSLtoString(t.info["name"]))
	torrent.WriteString("10:piece lengthi" + OSLtoString(t.info["piece_length"]) + "e")

	totalSize := int64(0)
	for _, file := range t.files {
		totalSize += int64(OSLcastNumber(file["length"]))
	}

	torrent.WriteString("6:lengthi" + OSLtoString(totalSize) + "e")

	torrent.WriteString("5:filesl")

	for _, file := range t.files {
		torrent.WriteString("d")
		torrent.WriteString("6:lengthi" + OSLtoString(file["length"]) + "e")

		if path, ok := file["path"]; ok {
			torrent.WriteString("4:pathl" + OSLtoString(path) + "e")
		}

		if md5sum, ok := file["md5sum"]; ok && md5sum != "" {
			torrent.WriteString("4:md5sum32:" + OSLtoString(md5sum) + "e")
		}

		torrent.WriteString("e")
	}

	torrent.WriteString("e")
	torrent.WriteString("e")

	return torrent.String()
}

func (Torrent) parse(torrentData any) *Torrent {
	dataStr := OSLtoString(torrentData)

	t := &Torrent{
		info:     map[string]any{},
		metadata: make(map[string]string),
	}

	if strings.Contains(dataStr, "d8:announce") {
		t.metadata["announce"] = t.extractString(dataStr, "d8:announce", ":")
		t.parseInfo(dataStr)
	}

	return t
}

func (t *Torrent) parseInfo(data string) string {
	infoStart := strings.Index(data, "6:infod")
	if infoStart == -1 {
		return ""
	}

	infoEnd := t.findMatchingBracket(data, infoStart+7)
	if infoEnd == -1 {
		return ""
	}

	infoData := data[infoStart+7 : infoEnd]

	if strings.Contains(infoData, "5:name") {
		t.info["name"] = t.extractString(infoData, "5:name", "")
	}

	if strings.Contains(infoData, "6:lengthi") {
		t.info["length"] = t.extractInt(infoData, "6:lengthi", "i")
	}

	if strings.Contains(infoData, "10:piece_lengthi") {
		t.info["piece_length"] = t.extractInt(infoData, "10:piece_lengthi", "i")
	}

	if strings.Contains(infoData, "6:pieces") {
		piecesLength := t.extractInt(infoData, "6:pieces", ":")
		piecesStart := strings.Index(infoData, "6:pieces") + 10
		if piecesStart > 0 && piecesLength > 0 {
			t.pieces = t.parsePieces(data[piecesStart+1:piecesStart+1+piecesLength])
		}
	}

	return ""
}

func (t *Torrent) parsePieces(data string) []string {
	var pieces []string
	pieceLength := 20

	for i := 0; i < len(data); i += pieceLength {
		if i+pieceLength <= len(data) {
			pieces = append(pieces, data[i:i+pieceLength])
		}
	}

	return pieces
}

func (Torrent) extractString(data string, marker string, suffix string) string {
	markerIndex := strings.Index(data, marker)
	if markerIndex == -1 {
		return ""
	}

	start := markerIndex + len(marker)
	end := strings.Index(data[start:], suffix)

	if end == -1 {
		return ""
	}

	return data[start : start+end]
}

func (t *Torrent) extractInt(data string, marker string, suffix string) string {
	return t.extractString(data, marker, suffix)
}

func (Torrent) findMatchingBracket(data string, start int) int {
	depth := 0

	for i := start; i < len(data); i++ {
		if data[i] == 'd' || data[i] == 'l' {
			depth++
		} else if data[i] == 'e' {
			depth--
			if depth == 0 {
				return i
			}
		}
	}

	return -1
}

func (t *Torrent) addTracker(tracker any) bool {
	trackers, ok := t.info["trackers"].([]string)
	
	if !ok {
		t.info["trackers"] = []string{}
		trackers = t.info["trackers"].([]string)
	}

	trackers = append(trackers, OSLtoString(tracker))
	t.info["trackers"] = trackers
	
	return true
}

func (t *Torrent) removeTracker(tracker any) bool {
	trackers, ok := t.info["trackers"].([]string)
	if !ok {
		return false
	}

	trackerStr := OSLtoString(tracker)
	newTrackers := make([]string, 0, len(trackers))

	found := false
	for _, t := range trackers {
		if t == trackerStr {
			found = true
		} else {
			newTrackers = append(newTrackers, t)
		}
	}

	t.info["trackers"] = newTrackers
	return found
}

func (t *Torrent) getTrackers() []any {
	trackers, _ := t.info["trackers"].([]string)
	
	result := make([]any, len(trackers))
	for i, t := range trackers {
		result[i] = t
	}

	return result
}

func (t *Torrent) getInfo() map[string]any {
	return t.info
}

func (t *Torrent) addMetadata(key any, value any) {
	keyStr := OSLtoString(key)
	valueStr := OSLtoString(value)
	t.metadata[keyStr] = valueStr
}

func (t *Torrent) getMetadata(key any) string {
	keyStr := OSLtoString(key)
	return t.metadata[keyStr]
}

func (t *Torrent) getAllMetadata() map[string]any {
	metadata := make(map[string]any)
	for k, v := range t.metadata {
		metadata[k] = v
	}
	return metadata
}

func (t *Torrent) getFileIndex(path any) int {
	pathStr := OSLtoString(path)

	for i, file := range t.files {
		if fp, ok := file["path"]; ok && fp == pathStr {
			return i
		}
	}

	return -1
}

func (t *Torrent) getPieceHashes() []string {
	return t.pieces
}

func (t *Torrent) setPieceHash(index any, hash any) bool {
	indexInt := int(OSLcastNumber(index))
	hashStr := OSLtoString(hash)

	if indexInt < 0 || indexInt >= len(t.pieces) {
		return false
	}

	t.pieces[indexInt] = hashStr
	return true
}

func (t *Torrent) validate() bool {
	if t.info["name"] == "" {
		return false
	}

	totalFiles := len(t.files)
	if totalFiles == 0 {
		return false
	}

	return true
}

func (t *Torrent) download(torrentPath any, outputPath any) bool {
	torrentStr := OSLtoString(torrentPath)
	outputStr := OSLtoString(outputPath)

	Torrent.parse(fs.ReadFile(torrentStr))

	if !t.validate() {
		return false
	}

	return true
}

func (t *Torrent) seed(torrentPath any, port any) bool {
	torrentStr := OSLtoString(torrentPath)
	portStr := OSLtoString(port)

	if portStr == "" {
		portStr = "6881"
	}

	Torrent.parse(fs.ReadFile(torrentStr))

	return true
}

func (t *Torrent) getMagnetLink() string {
	name := OSLtoString(t.info["name"])
	b64Name := btoa(name)
	infoHash := t.calculateInfoHash()

	return fmt.Sprintf("magnet:?xt=urn:btih:%s&dn=%s", infoHash, b64Name)
}

func (t *Torrent) calculateInfoHash() string {
	infoData := t.buildTorrentInfo()
	hash := crypto.sha1(infoData)
	return hash
}

func (t *Torrent) buildTorrentInfo() string {
	var torrent strings.Builder

	torrent.WriteString("d")
	torrent.WriteString("5:name" + OSLtoString(t.info["name"]))
	torrent.WriteString("10:piece lengthi" + OSLtoString(t.info["piece_length"]) + "e")

	totalSize := int64(0)
	for _, file := range t.files {
		totalSize += int64(OSLcastNumber(file["length"]))
	}

	torrent.WriteString("6:lengthi" + OSLtoString(totalSize) + "e")

	torrent.WriteString("5:filesl")

	for _, file := range t.files {
		torrent.WriteString("d")
		torrent.WriteString("6:lengthi" + OSLtoString(file["length"]) + "e")

		if path, ok := file["path"]; ok {
			torrent.WriteString("4:pathl" + OSLtoString(path) + "e")
		}

		if md5sum, ok := file["md5sum"]; ok && md5sum != "" {
			torrent.WriteString("4:md5sum32:" + OSLtoString(md5sum) + "e")
		}

		torrent.WriteString("e")
	}

	torrent.WriteString("e")

	return torrent.String()
}

func (t *Torrent) getFiles() []map[string]any {
	return t.files
}

func (t *Torrent) setFiles(files []map[string]any) {
	t.files = files
}

func (t *Torrent) getFileCount() int {
	return len(t.files)
}

func (t *Torrent) getTotalSize() int64 {
	total := int64(0)

	for _, file := range t.files {
		total += int64(OSLcastNumber(file["length"]))
	}

	return total
}

func (t *Torrent) getPieceCount() int {
	return len(t.pieces)
}

func (t *Torrent) generatePeerID() string {
	randomData := crypto.randomBytes(20)
	return "OSL-" + randomData[:12]
}

func (t *Torrent) getMagnetURI() string {
	infoHash := t.calculateInfoHash()
	name := OSLtoString(t.info["name"])
	nameParam := ""

	if name != "" {
		encodedName := url.QueryEscape(name)
		nameParam = "&dn=" + encodedName
	}

	return fmt.Sprintf("magnet:?xt=urn:btih:%s%s", infoHash, nameParam)
}

func (t *Torrent) exportInfo(path any) bool {
	infoMap := t.getInfo()
	return fs.WriteFile(path, JsonFormat(infoMap))
}

func (t *Torrent) clone() *Torrent {
	var tt Torrent
	newTorrent := tt.create(t.info["name"], t.getFiles(), t.info["piece_length"])
	
	newTorrent.pieces = make([]string, len(t.pieces))
	copy(newTorrent.pieces, t.pieces)
	
	for k, v := range t.metadata {
		newTorrent.addMetadata(k, v)
	}

	return newTorrent
}

func (t *Torrent) merge(otherTorrent *Torrent) *Torrent {
	otherFiles := otherTorrent.getFiles()
	myFiles := t.getFiles()
	mergedFiles := append(myFiles, otherFiles...)
	
	var tt Torrent
	return tt.create(t.info["name"], mergedFiles, t.info["piece_length"])
}

func (t *Torrent) strip(metadata any) *Torrent {
	keyStr := strings.ToLower(OSLtoString(metadata))

	delete(t.metadata, keyStr)
	return t
}

var torrent = Torrent{}
var torrentTracker = Torrent{}
