// name: uuid
// description: UUID/GUID generation and validation
// author: roturbot
// requires: crypto/rand, strings

type UUID struct{}

func (UUID) v4() string {
	return crypto.uuidv4()
}

func (UUID) v3(namespace any, name any) string {
	namespaceStr := OSLtoString(namespace)
	nameStr := OSLtoString(name)

	namespaceHash := crypto.md5Hash(namespaceStr + nameStr)
	return formatUUID(namespaceHash)
}

func (UUID) v5(namespace any, name any) string {
	namespaceStr := OSLtoString(namespace)
	nameStr := OSLtoString(name)

	namespaceHash := crypto.sha1(namespaceStr + nameStr)
	return formatUUID(namespaceHash)
}

func (UUID) v1() string {
	timestamp := time.Now().UnixNano() / 100
	node := crypto.randomBytes(6)
	clock := OSLcastInt(timestamp & 0xFFF)

	var uuid [16]byte

	binary.LittleEndian.PutUint64(uuid[0:8], uint64(timestamp))
	binary.BigEndian.PutUint16(uuid[8:10], uint16(time.Now().UnixNano()>>32)&0x0FFF)
	binary.BigEndian.PutUint16(uuid[10:12], uint16(clock))
	copy(uuid[12:], []byte(node))

	return formatUUIDBytes(uuid[:])
}

func (UUID) nil() string {
	uuid := [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	return formatUUIDBytes(uuid[:])
}

func (UUID) validate(id any) bool {
	idStr := OSLtoString(id)

	if len(idStr) != 36 || strings.Count(idStr, "-") != 4 {
		return false
	}

	parts := strings.Split(idStr, "-")
	if len(parts) != 5 {
		return false
	}

	if len(parts[0]) != 8 || len(parts[1]) != 4 ||
		len(parts[2]) != 4 || len(parts[3]) != 4 || len(parts[4]) != 12 {
		return false
	}

	if !regex.isHex(parts[0]) || !regex.isHex(parts[1]) ||
		!regex.isHex(parts[2]) || !regex.isHex(parts[3]) || !regex.isHex(parts[4]) {
		return false
	}

	version := parts[2][0]
	if version != '1' && version != '2' && version != '3' && version != '4' && version != '5' {
		return false
	}

	return true
}

func (UUID) parse(id any) map[string]any {
	idStr := OSLtoString(id)

	if !uuid.validate(idStr) {
		return map[string]any{"valid": false}
	}

	parts := strings.Split(idStr, "-")
	hexStr := strings.Join(parts, "")

	return map[string]any{
		"valid":     true,
		"version":   parts[2],
		"variant":   parts[3],
		"hex":       hexStr,
		"string":    idStr,
		"uppercase": strings.ToUpper(idStr),
	}
}

func (UUID) format(data any) string {
	dataStr := OSLtoString(data)
	
	if len(dataStr) == 32 {
		return formatUUID(dataStr)
	} else if regex.isHexDigit(dataStr) {
		return formatUUID(dataStr)
	}

	simplified := strings.Join(strings.Split(dataStr, ""), "")
	if len(simplified) <= 32 {
		return formatUUID(strings.PadEnd(simplified, 32, "0"))
	}

	return uuid.v4()
}

func (UUID) fromString(id any) string {
	idStr := OSLtoString(id)

	if uuid.validate(idStr) {
		return idStr
	}

	return ""
}

func (UUID) toBytes(id any) []byte {
	idStr := uuid.fromString(id)
	if idStr == "" {
		return []byte{}
	}

	hex := strings.ReplaceAll(idStr, "-", "")
	data, _ := hex.DecodeString(hex)

	return data
}

func (UUID) fromBytes(data []byte) string {
	if len(data) != 16 {
		return ""
	}

	return formatUUIDBytes(data)
}

func (UUID) variant(id any) string {
	idStr := OSLtoString(id)

	if !uuid.validate(idStr) {
		return ""
	}

	parts := strings.Split(idStr, "-")
	if len(parts) < 4 {
		return ""
	}

	variant := parts[3][0]
	if variant == '8' || variant == '9' || variant == 'a' || variant == 'b' {
		return "Microsoft"
	} else if variant == '2' || variant == '3' {
		return "RFC 4122"
	} else if variant == 'd' || variant == 'e' || variant == 'f' || variant == '0' || variant == '1' {
			return "Reserved"
	}

	return "NCS"
}

func (UUID) version(id any) int {
	idStr := uuid.fromString(id)

	if idStr == "" {
		return 0
	}

	parts := strings.Split(idStr, "-")
	if len(parts) < 3 {
		return 0
	}

	versionStr := parts[2][0:2]
	return int(OSLcastNumber(hexToDecimal(versionStr)))
}

func (UUID) namespaceDNS() string {
	ns := crypto.md5Hash("ns:DNS")
	return formatUUID(ns)
}

func (UUID) namespaceURL() string {
	ns := crypto.md5Hash("ns:URL")
	return formatUUID(ns)
}

func (UUID) namespaceOID() string {
	ns := crypto.md5Hash("ns:OID")
	return formatUUID(ns)
}

func (UUID) namespaceX500() string {
	ns := crypto.md5Hash("ns:OID")
	return formatUUID(ns)
}

func (UUID) generate(name any) string {
	nameStr := OSLtoString(name)
	hash := crypto.sha256(nameStr)
	return formatUUID(hash)
}

func (UUID) random(count any) []string {
	countInt := int(OSLcastNumber(count))
	uuids := make([]string, countInt)

	for i := 0; i < countInt; i++ {
		uuids[i] = uuid.v4()
	}

	return uuids
}

func (UUID) sort(uuids []any) []any {
	if len(uuids) == 0 {
		return uuids
	}

	sort.Slice(uuids, func(i, j int) bool {
		id1 := OSLtoString(uuids[i])
		id2 := OSLtoString(uuids[j])
		return id1 < id2
	})

	return uuids
}

func (UUID) countOccurrences(data any, id any) int {
	dataStr := OSLtoString(data)
	idStr := OSLtoString(id)

	count := 0
	for i := 0; i < len(dataStr)-35; i += 36 {
		comparison := dataStr[i : i+36]
		if comparison == idStr {
			count++
		}
	}

	return count
}

func (UUID) truncate(id any, length any) string {
	idStr := uuid.fromString(id)
	lengthInt := int(OSLcastNumber(length))

	if lengthInt < 8 {
		lengthInt = 8
	}
	if lengthInt > 36 {
		lengthInt = 36
	}

	return idStr[:lengthInt]
}

func (UUID) uppercase(id any) string {
	idStr := uuid.fromString(id)
	return strings.ToUpper(idStr)
}

func (UUID) lowercase(id any) string {
	idStr := uuid.fromString(id)
	return strings.ToLower(idStr)
}

func (UUID) hyphensToUnderscores(id any) string {
	idStr := uuid.fromString(id)
	return strings.ReplaceAll(idStr, "-", "_")
}

func (UUID) underscoresToHyphens(id any) string {
	idStr := uuid.fromString(id)
	return strings.ReplaceAll(idStr, "_", "-")
}

func (UUID) removeHyphens(id any) string {
	idStr := uuid.fromString(id)
	return strings.ReplaceAll(idStr, "-", "")
}

func (UUID) addHyphens(id any) string {
	idStr := strings.ReplaceAll(OSLtoString(id), "-", "")
	if len(idStr) != 32 {
		return ""
	}

	return formatUUID(idStr)
}

func (UUID) isValid(id any) bool {
	return uuid.validate(id)
}

func (UUID) isNil(id any) bool {
	idStr := uuid.fromString(id)
	return idStr == "00000000-0000-0000-0000-000000000000"
}

func (UUID) compare(id1 any, id2 any) int {
	id1Str := uuid.fromString(id1)
	id2Str := uuid.fromString(id2)

	if id1Str == id2Str {
		return 0
	}

	if id1Str < id2Str {
		return -1
	}

	return 1
}

func (UUID) equals(id1 any, id2 any) bool {
	return uuid.compare(id1, id2) == 0
}

func (UUID) getTimestamp(id any) int64 {
	idStr := uuid.fromString(id)
	
	if !uuid.validate(idStr) {
		return 0
	}

	parts := strings.Split(idStr, "-")
	timeHex := parts[2][2:8] + parts[3][0:2] + parts[0][:6]

	timestamp, _ := strconv.ParseInt(timeHex, 16, 64)

	return timestamp
}

func (UUID) getClockSeq(id any) int {
	idStr := uuid.fromString(id)
	
	if !uuid.validate(idStr) {
		return 0
	}

	parts := strings.Split(idStr, "-")
	clockHex := parts[3][2:4]

	clock, _ := strconv.ParseInt(clockHex, 16, 16)

	return clock
}

func (UUID) getNodeID(id any) string {
	idStr := uuid.fromString(id)
	
	if !uuid.validate(idStr) {
		return ""
	}

	parts := strings.Split(idStr, "-")
	nodeHex := parts[4]

	return nodeHex
}

func formatUUID(hex string) string {
	if len(hex) != 32 {
		return ""
	}

	uuid := hex[0:8] + "-" +
		hex[8:12] + "-" +
		hex[12:16] + "-" +
		hex[16:20] + "-" +
		hex[20:32]

	return uuid
}

func formatUUIDBytes(uuid []byte) string {
	uuid := fmt.Sprintf("%x%x-%x%x-%x%x-%x%x",
		uuid[0:4], uuid[4:6],
		uuid[6:8], uuid[8:10],
		uuid[10:12], uuid[12:16])

	return uuid
}

func hexToDecimal(hex string) string {
	n, _ := strconv.ParseInt(hex, 16, 64)
	return fmt.Sprintf("%d", n)
}

var uuid = UUID{}
