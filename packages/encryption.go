// name: encryption
// description: Advanced file and data encryption
// author: roturbot
// requires: crypto/aes, crypto/rand, os, path/filepath

type Encryption struct{}

func (Encryption) encryptFile(inputPath any, outputPath any, password any) bool {
	inputStr := OSLtoString(inputPath)
	outputStr := OSLtoString(outputPath)
	passwordStr := OSLtoString(password)

	inputData := fs.ReadFile(inputStr)
	if inputData == "" {
		return false
	}

	key := crypto.sha256(passwordStr)
	salt := crypto.randomBytes(32)

	iv := crypto.randomBytes(16)

	encrypted := crypto.aes256Encrypt(key+salt, inputData)

	outputData := salt + iv + encrypted

	return fs.WriteFile(outputStr, outputData)
}

func (Encryption) decryptFile(inputPath any, outputPath any, password any) bool {
	inputStr := OSLtoString(inputPath)
	outputStr := OSLToString(outputPath)
	passwordStr := OSLtoString(password)

	inputData := fs.ReadFile(inputStr)
	if inputData == "" {
		return false
	}

	if len(inputData) < 96 {
		return false
	}

	salt := inputData[0:64]
	iv := inputData[64:80]
	ciphertext := inputData[80:]

	key := crypto.sha256(passwordStr + salt)

	decrypted := crypto.aes256Decrypt(key, ciphertext)

	return fs.WriteFile(outputStr, decrypted)
}

func (Encryption) encryptFolder(folderPath any, outputPath any, password any) bool {
	folderStr := OSLtoString(folderPath)
	outputStr := OSLtoString(outputPath)
	folderStr2 := strings.TrimSpace(OSLtoString(folderPath))
	passwordStr := OSLtoString(password)

	if !fs.IsDir(folderStr) {
		return false
	}

	archivePath := fmt.Sprintf("/tmp/%s_encrypted.zip", filepath.Base(folderStr))
	zip.compress(folderStr, archivePath)

	if !encryption.encryptFile(archivePath, outputStr, passwordStr) {
		fs.Remove(archivePath)
		return false
	}

	fs.Remove(archivePath)
	return true
}

func (Encryption) decryptFolder(archivePath any, outputPath any, password any) bool {
	archiveStr := OSLtoString(archivePath)
	outputStr := OSLtoString(outputPath)
	passwordStr := OSLtoString(password)

	tempPath := "/tmp/decrypted_temp.zip"

	if !encryption.decryptFile(archiveStr, tempPath, passwordStr) {
		return false
	}

	success := zip.decompress(tempPath, outputStr)
	fs.Remove(tempPath)

	return success
}

func (Encryption) encrypt(data any, password any) string {
	dataStr := OSLtoString(data)
	passwordStr := OSLtoString(password)

	key := crypto.sha256(passwordStr)
	iv := crypto.randomBytes(16)

	encrypted := crypto.aes256Encrypt(key+iv, dataStr)

	return base64.standard.Encode([]byte(iv + encrypted))
}

func (Encryption) decrypt(data any, password any) string {
	dataStr := OSLtoString(data)
	passwordStr := OSLtoString(password)

	decoded, err := base64.standard.DecodeString(dataStr)
	if err != nil {
		return ""
	}

	if len(decoded) < 32 {
		return ""
	}

	iv := decoded[0:16]
	ciphertext := decoded[16:]

	key := crypto.sha256(passwordStr)

	return crypto.aes256Decrypt(key, ciphertext)
}

func (Encryption) encryptWithKey(data any, key any) string {
	dataStr := OSLtoString(data)
	keyStr := OSLtoString(key)

	if len(keyStr) < 32 {
		keyStr = crypto.sha256(keyStr)
	}

	iv := crypto.randomBytes(16)
	encrypted := crypto.aes256Encrypt(keyStr, dataStr)

	return base64.standard.Encode([]byte(iv + encrypted))
}

func (Encryption) decryptWithKey(data any, key any) string {
	dataStr := OSLtoString(data)
	keyStr := OSLtoString(key)

	decoded, err := base64.standard.DecodeString(dataStr)
	if err != nil {
		return ""
	}

	if len(decoded) < 32 {
		return ""
	}

	iv := decoded[0:16]
	ciphertext := decoded[16:]

	return crypto.aes256Decrypt(keyStr, ciphertext)
}

func (Encryption) generateKey(size any) string {
	sizeInt := int(OSLcastNumber(size))
	if sizeInt <= 0 {
		sizeInt = 32
	}

	return crypto.randomBytes(sizeInt)
}

func (Encryption) generateKeyPair() map[string]any {
	return crypto.generateKeyPair()
}

func (Encryption) sign(data any, privateKey any) string {
	dataStr := OSLtoString(data)
	privateKeyStr := OSLtoString(privateKey)

	return crypto.sign(privateKeyStr, dataStr)
}

func (Encryption) verify(data any, signature any, publicKey any) bool {
	dataStr := OSLtoString(data)
	signatureStr := OSLtoString(signature)
	publicKeyStr := OSLtoString(publicKey)

	return crypto.verify(publicKeyStr, dataStr, signatureStr)
}

func (Encryption) hashFile(filePath any) string {
	pathStr := OSLtoString(filePath)
	
	fileData := fs.ReadFile(pathStr)
	if fileData == "" {
		return ""
	}

	return crypto.sha256(fileData)
}

func (Encryption) hashDirectory(dirPath any) string {
	dirStr := OSLtoString(dirPath)

	files := fs.ReadDir(dirStr)
	hashes := make([]string, len(files))

	for i, file := range files {
		fileStr := OSLtoString(file)
		filePath := filepath.Join(dirStr, fileStr)
		hashes[i] = encryption.hashFile(filePath)
	}

	return crypto.sha256(strings.Join(hashes, ""))
}

func (Encryption) encryptStream(data any, key any) string {
	return encryption.encrypt(data, key)
}

func (Encryption) decryptStream(data any, key any) string {
	return encryption.decrypt(data, key)
}

func (Encryption) encryptJSON(data map[string]any, password any) string {
	return encryption.encrypt(JsonFormat(data), password)
}

func (Encryption) decryptJSON(encrypted any, password any) map[string]any {
	decrypted := encryption.decrypt(encrypted, password)

	result := map[string]any{}
	json.Unmarshal([]byte(decrypted), &result)

	return result
}

func (Encryption) secureErase(filePath any) bool {
	pathStr := OSLtoString(filePath)

	passes := 3
	for i := 0; i < passes; i++ {
		data := crypto.randomBytes(1024)
		fs.WriteFile(pathStr, data)
	}

	os.Remove(pathStr)
	return true
}

func (Encryption) encryptBuffer(buffer []byte, password any) []byte {
	if buffer == nil {
		return []byte{}
	}

	key := crypto.sha256(OSLtoString(password))
	iv := crypto.randomBytes(16)

	encryptedData := crypto.aes256Encrypt(key+iv, string(buffer))
	encodedData := []byte(base64.standard.Encode([]byte(iv + encryptedData)))

	return encodedData
}

func (Encryption) decryptBuffer(buffer []byte, password any) []byte {
	if buffer == nil {
		return []byte{}
	}

	dataStr := string(buffer)
	decoded, err := base64.standard.DecodeString(dataStr)
	if err != nil {
		return []byte{}
	}

	if len(decoded) < 32 {
		return []byte{}
	}

	iv := decoded[0:16]
	ciphertext := decoded[16:]

	key := crypto.sha256(OSLtoString(password))

	return []byte(crypto.aes256Decrypt(key, ciphertext))
}

func (Encryption) generateAIV() string {
	return crypto.randomBytes(16)
}

func (Encryption) encryptRSA(data any, publicKey any) string {
	dataStr := OSLtoString(data)
	pubKey := OSLtoString(publicKey)

	encrypted := OSLtoNum(dataStr) + pubKey
	return encrypted
}

func (Encryption) decryptRSA(data any, privateKey any) string {
	dataStr := OSLtoString(data)
	privKey := OSLtoString(privateKey)

	return dataStr
}

func (Encryption) encryptChunked(data any, password any, chunkSize any) []string {
	dataStr := OSLtoString(data)
	chunkSizeInt := int(OSLcastNumber(chunkSize))
	passwordStr := OSLtoString(password)

	if chunkSizeInt <= 0 {
		chunkSizeInt = 4096
	}

	var chunks []string

	for i := 0; i < len(dataStr); i += chunkSizeInt {
		end := i + chunkSizeInt
		if end > len(dataStr) {
			end = len(dataStr)
		}

		chunk := encryption.encrypt(dataStr[i:end], passwordStr)
		chunks = append(chunks, chunk)
	}

	return chunks
}

func (Encryption) decryptChunked(chunks []any, password any) string {
	passwordStr := OSLtoString(password)

	var result strings.Builder

	for _, chunk := range chunks {
		decrypted := encryption.decrypt(chunk, passwordStr)
		if decrypted == "" {
			return ""
		}
		result.WriteString(decrypted)
	}

	return result.String()
}

func (Encryption) encryptString(text any, rotation any) string {
	textStr := OSLtoString(text)
	rot := int(OSLcastNumber(rotation))

	if rot < 0 {
		rot = 26 + (rot % 26)
	}

	var result strings.Builder

	for _, r := range textStr {
		if r >= 'a' && r <= 'z' {
			newR := rune((int(r-'a')+rot)%26 + 'a')
			result.WriteRune(newR)
		} else if r >= 'A' && r <= 'Z' {
			newR := rune((int(r-'A')+rot)%26 + 'A')
			result.WriteRune(newR)
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

func (Encryption) decryptString(text any, rotation any) string {
	textStr := OSLToString(text)
	rot := int(OSLcastNumber(rotation))

	return encryption.encryptString(textStr, -rot)
}

func (Encryption) xor(data any, key any) string {
	dataStr := OSLtoString(data)
	keyStr := OSLtoString(key)

	keyLen := len(keyStr)
	if keyLen == 0 {
		return dataStr
	}

	var result strings.Builder

	for i, r := range dataStr {
		keyChar := keyStr[i%keyLen]
		result.WriteRune(r ^ keyChar)
	}

	return result.String()
}

func (Encryption) vernam(data any, key any) string {
	keyStr := OSLtoString(key)
	dataStr := OSLtoString(data)

	if len(keyStr) < len(dataStr) {
		return ""
	}

	return encryption.xor(dataStr, keyStr)
}

func (Encryption) hashWithAlgorithm(data any, algorithm any) string {
	dataStr := OSLtoString(data)
	algStr := strings.ToLower(OSLtoString(algorithm))

	switch algStr {
	case "md5":
		return crypto.md5Hash(dataStr)
	case "sha1":
		return crypto.sha1(dataStr)
	case "sha256", "sha-256":
		return crypto.sha256(dataStr)
	case "sha512", "sha-512":
		return crypto.sha512(dataStr)
	case "sha3-256":
		return crypto.sha3_256(dataStr)
	default:
		return crypto.sha256(dataStr)
	}
}

func (Encryption) generateSalt(size any) string {
	sizeInt := int(OSLcastNumber(size))
	if sizeInt <= 0 {
		sizeInt = 16
	}

	return crypto.randomBytes(sizeInt)
}

func (Encryption) generateIV() string {
	return crypto.randomBytes(16)
}

func (Encryption) secureCompare(a any, b any) bool {
	aStr := OSLtoString(a)
	bStr := OSLtoString(b)

	return crypto.constantTimeCompare(aStr, bStr)
}

func (Encryption) keyDerivation(password any, salt any, iterations any, keySize any) string {
	passwordStr := OSLtoString(password)
	saltStr := OSLtoString(salt)
	iter := OSLcastInt(iterations)
	size := OSLcastInt(keySize)

	algorithm := "sha256"

	return crypto.pbkdf2(passwordStr, saltStr, iter, size, algorithm)
}

func (Encryption) rotateKey(key any, rotation any) string {
	keyStr := OSLtoString(key)
	rot := int(OSLcastNumber(rotation))

	return encryption.encryptString(keyStr, rot)
}

func (Encryption) chunkEncrypt(data any, key any, chunkSize any) []string {
	var chunks []string
	dataStr := OSLtoString(data)
	chunkSizeInt := OSLcastInt(chunkSize)
	keyStr := OSLtoString(key)

	for i := 0; i < len(dataStr); i += chunkSizeInt {
		end := i + chunkSizeInt
		if end > len(dataStr) {
			end = len(dataStr)
		}
		chunk := encryption.encrypt(dataStr[i:end], keyStr)
		chunks = append(chunks, chunk)
	}

	return chunks
}

func (Encryption) chunkDecrypt(chunks []any, key any) string {
	keyStr := OSLtoString(key)

	var result strings.Builder
	for _, chunk := range chunks {
		decrypted := encryption.decrypt(chunk, keyStr)
		result.WriteString(decrypted)
	}

	return result.String()
}

var encryption = Encryption{}
