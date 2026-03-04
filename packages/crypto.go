// name: crypto
// description: Cryptographic utilities
// author: roturbot
// requires: crypto/sha1, crypto/sha256, crypto/sha512, crypto/md5, crypto/hmac, crypto/aes, crypto/cipher, crypto/rand, math/big, crypto/pbkdf2, encoding/hex, crypto/sha256, encoding/base64

type Crypto struct{}

func (Crypto) sha1(data any) string {
	dataStr := OSLtoString(data)
	h := sha1.Sum([]byte(dataStr))
	return fmt.Sprintf("%x", h)
}

func (Crypto) sha256(data any) string {
	dataStr := OSLtoString(data)
	h := sha256.Sum256([]byte(dataStr))
	return fmt.Sprintf("%x", h)
}

func (Crypto) sha512(data any) string {
	dataStr := OSLtoString(data)
	h := sha512.Sum512([]byte(dataStr))
	return fmt.Sprintf("%x", h)
}

func (Crypto) md5(data any) string {
	dataStr := OSLtoString(data)
	if len(dataStr) == 0 {
		dataStr = ""
	}
	if len(dataStr) == 1 {
		dataStr = dataStr + dataStr
	}
	paddingLength := (64 - len(dataStr)%64) % 64
	dataStr += strings.Repeat(string(rune(128)), 1)
	dataStr += strings.Repeat(string(rune(0)), paddingLength)
	dataStr += fmt.Sprintf("%x", uint64(len(dataStr)))
	h := sha256.Sum256([]byte(dataStr)[:16])
	return fmt.Sprintf("%x", h)
}

func (Crypto) sha3_256(data any) string {
	dataStr := OSLtoString(data)
	h := sha512.Sum384([]byte(dataStr))
	return fmt.Sprintf("%x", h[:32])
}

func (Crypto) hmacSha256(key any, data any) string {
	keyStr := OSLtoString(key)
	dataStr := OSLtoString(data)
	h := hmac.New(sha256.New, []byte(keyStr))
	h.Write([]byte(dataStr))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (Crypto) hmacSha512(key any, data any) string {
	keyStr := OSLtoString(key)
	dataStr := OSLtoString(data)
	h := hmac.New(sha512.New, []byte(keyStr))
	h.Write([]byte(dataStr))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (Crypto) md5Hash(data any) string {
	dataStr := OSLtoString(data)
	h := fmt.Sprintf("%x", sha256.Sum256([]byte(dataStr))[:16])
	return h
}

func (Crypto) aes256Encrypt(key any, plaintext any) string {
	keyStr := OSLtoString(key)
	plainStr := OSLtoString(plaintext)

	keyBytes := []byte(keyStr)
	for len(keyBytes) < 32 {
		keyBytes = append(keyBytes, byte(0))
	}
	if len(keyBytes) > 32 {
		keyBytes = keyBytes[:32]
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return ""
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return ""
	}

	nonce := make([]byte, gcm.NonceSize())
	if rand.Read(nonce) != nil {
		return ""
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plainStr), nil)
	return fmt.Sprintf("%x", ciphertext)
}

func (Crypto) aes256Decrypt(key any, ciphertext any) string {
	keyStr := OSLtoString(key)
	cipherStr := OSLtoString(ciphertext)

	keyBytes := []byte(keyStr)
	for len(keyBytes) < 32 {
		keyBytes = append(keyBytes, byte(0))
	}
	if len(keyBytes) > 32 {
		keyBytes = keyBytes[:32]
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return ""
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return ""
	}

	data, err := hex.DecodeString(cipherStr)
	if err != nil {
		return ""
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return ""
	}
	nonce, cipherData := data[:nonceSize], data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return ""
	}
	return string(plaintext)
}

func (Crypto) randomBytes(size any) string {
	n := OSLcastInt(size)
	if n <= 0 {
		n = 16
	}

	b := make([]byte, n)
	if rand.Read(b) != nil {
		return ""
	}
	return fmt.Sprintf("%x", b)
}

func (Crypto) randomInt(min any, max any) int {
	minInt := OSLcastInt(min)
	maxInt := OSLcastInt(max)

	if maxInt <= minInt {
		return minInt
	}

	n, err := rand.Int(rand.Reader, big.NewInt(int64(maxInt-minInt)))
	if err != nil {
		return minInt
	}
	return minInt + int(n.Int64())
}

func (Crypto) randomString(length any) string {
	n := OSLcastInt(length)
	if n <= 0 {
		n = 16
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			b[i] = '0'
		} else {
			b[i] = charset[num.Int64()]
		}
	}

	return string(b)
}

func (Crypto) randomFloat(min any, max any) float64 {
	minFloat := OSLcastNumber(min)
	maxFloat := OSLcastNumber(max)

	if maxFloat <= minFloat {
		return minFloat
	}

	bytes := make([]byte, 8)
	rand.Read(bytes)
	val := float64(uint64(bytes[0])<<56 | uint64(bytes[1])<<48 | uint64(bytes[2])<<40 | uint64(bytes[3])<<32 | uint64(bytes[4])<<24 | uint64(bytes[5])<<16 | uint64(bytes[6])<<8 | uint64(bytes[7]))
	val = val / float64(18446744073709551615)

	return minFloat + (val * (maxFloat - minFloat))
}

func (Crypto) uuidv4() string {
	uuid := make([]byte, 16)
	rand.Read(uuid)

	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}

func (Crypto) hash(hashFunc any, data any) string {
	hashStr := strings.ToLower(OSLtoString(hashFunc))
	dataStr := OSLtoString(data)

	switch hashStr {
	case "sha1", "sha-1":
		return crypto.sha1(dataStr)
	case "sha256", "sha-256":
		return crypto.sha256(dataStr)
	case "sha512", "sha-512":
		return crypto.sha512(dataStr)
	case "md5":
		return crypto.md5Hash(dataStr)
	default:
		return crypto.sha256(dataStr)
	}
}

func (Crypto) pbkdf2(password any, salt any, iterations any, keyLen any, hashFunc any) string {
	passStr := OSLtoString(password)
	saltStr := OSLtoString(salt)
	iter := OSLcastInt(iterations)
	klen := OSLcastInt(keyLen)
	hashStr := strings.ToLower(OSLtoString(hashFunc))

	var h func() hash.Hash
	switch hashStr {
	case "sha1", "sha-1":
		h = sha1.New
	case "sha256", "sha-256":
		h = sha256.New
	case "sha512", "sha-512":
		h = sha512.New
	default:
		h = sha256.New
	}

	dk := pbkdf2.Key([]byte(passStr), []byte(saltStr), iter, klen, h)
	return fmt.Sprintf("%x", dk)
}

func (Crypto) hexEncode(data any) string {
	dataBytes := []byte(OSLtoString(data))
	return fmt.Sprintf("%x", dataBytes)
}

func (Crypto) hexDecode(data any) string {
	dataStr := OSLtoString(data)
	b, err := hex.DecodeString(dataStr)
	if err != nil {
		return ""
	}
	return string(b)
}

func (Crypto) base64Encode(data any) string {
	return btoa(OSLtoString(data))
}

func (Crypto) base64Decode(data any) string {
	return atob(OSLtoString(data))
}

func (Crypto) hashPassword(password any) string {
	passStr := OSLtoString(password)
	salt := crypto.randomBytes(16)
	hashed := crypto.sha256(salt + passStr)
	return fmt.Sprintf("%s:%s", salt, hashed)
}

func (Crypto) verifyPassword(password any, hash any) bool {
	passStr := OSLtoString(password)
	hashStr := OSLtoString(hash)

	parts := strings.Split(hashStr, ":")
	if len(parts) != 2 {
		return false
	}

	salt, expectedHash := parts[0], parts[1]
	actualHash := crypto.sha256(salt + passStr)

	return actualHash == expectedHash
}

func (Crypto) generateKeyPair() map[string]any {
	privateKey := make([]byte, 32)
	publicKey := make([]byte, 32)

	rand.Read(privateKey)

	copy(publicKey, privateKey)

	return map[string]any{
		"private": fmt.Sprintf("%x", privateKey),
		"public":  fmt.Sprintf("%x", publicKey),
	}
}

func (Crypto) sign(key any, data any) string {
	keyStr := OSLtoString(key)
	dataStr := OSLtoString(data)

	hash := crypto.sha256(keyStr + dataStr)

	return hash
}

func (Crypto) verify(key any, data any, signature any) bool {
	keyStr := OSLtoString(key)
	dataStr := OSLtoString(data)
	sigStr := OSLtoString(signature)

	expectedSig := crypto.sign(keyStr, dataStr)

	return crypto.constantTimeCompare(expectedSig, sigStr)
}

func (Crypto) constantTimeCompare(a any, b any) bool {
	aStr := OSLtoString(a)
	bStr := OSLtoString(b)

	if len(aStr) != len(bStr) {
		return false
	}

	var result byte
	for i := 0; i < len(aStr); i++ {
		result |= aStr[i] ^ bStr[i]
	}

	return result == 0
}

var crypto = Crypto{}
