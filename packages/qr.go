// name: qr
// description: QR code generation and barcode utilities
// author: roturbot
// requires: strings, image, image/png, os

type QR struct{}

func (QR) generate(data any, size any, outputFile any) bool {
	dataStr := OSLtoString(data)
	sizeInt := int(OSLcastNumber(size))
	filePath := OSLtoString(outputFile)

	if sizeInt <= 0 {
		sizeInt = 256
	}

	if sizeInt%2 != 0 {
		sizeInt++
	}

	moduleCount := qr.calculateModuleCount(dataStr)
	moduleSize := sizeInt / moduleCount

	img := image.NewRGBA(image.Rect(0, 0, sizeInt, sizeInt))
	white := color.RGBA{255, 255, 255, 255}
	black := color.RGBA{0, 0, 0, 255}

	qrData := qr.generateQRMatrix(dataStr, moduleCount)

	for y := 0; y < moduleCount; y++ {
		for x := 0; x < moduleCount; x++ {
			for my := 0; my < moduleSize; my++ {
				for mx := 0; mx < moduleSize; mx++ {
					px := x*moduleSize + mx
					py := y*moduleSize + my

					if qrData[y][x] {
						img.Set(px, py, black)
					} else {
						img.Set(px, py, white)
					}
				}
			}
		}
	}

	file, err := os.Create(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	png.Encode(file, img)
	return true
}

func (QR) generateColored(data any, size any, color any, outputFile any) bool {
	dataStr := OSLtoString(data)
	sizeInt := int(OSLcastNumber(size))
	colorStr := OSLtoString(color)
	filePath := OSLtoString(outputFile)

	if sizeInt <= 0 {
		sizeInt = 256
	}

	moduleCount := qr.calculateModuleCount(dataStr)
	moduleSize := sizeInt / moduleCount

	img := image.NewRGBA(image.Rect(0, 0, sizeInt, sizeInt))
	white := color.RGBA{255, 255, 255, 255}
	blackColor := colors.Hex(colorStr)

	qrData := qr.generateQRMatrix(dataStr, moduleCount)

	for y := 0; y < moduleCount; y++ {
		for x := 0; x < moduleCount; x++ {
			for my := 0; my < moduleSize; my++ {
				for mx := 0; mx < moduleSize; mx++ {
					px := x*moduleSize + mx
					py := y*moduleSize + my

					if qrData[y][x] {
						img.Set(px, py, blackColor)
					} else {
						img.Set(px, py, white)
					}
				}
			}
		}
	}

	file, err := os.Create(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	png.Encode(file, img)
	return true
}

func (QR) generateToDataURL(data any, size any) string {
	dataStr := OSLtoString(data)
	sizeInt := int(OSLcastNumber(size))

	if sizeInt <= 0 {
		sizeInt = 256
	}

	moduleCount := qr.calculateModuleCount(dataStr)
	moduleSize := sizeInt / moduleCount

	img := image.NewRGBA(image.Rect(0, 0, sizeInt, sizeInt))
	white := color.RGBA{255, 255, 255, 255}
	black := color.RGBA{0, 0, 0, 255}

	qrData := qr.generateQRMatrix(dataStr, moduleCount)

	for y := 0; y < moduleCount; y++ {
		for x := 0; x < moduleCount; x++ {
			for my := 0; my < moduleSize; my++ {
				for mx := 0; mx < moduleSize; mx++ {
					px := x*moduleSize + mx
					py := y*moduleSize + my

					if qrData[y][x] {
						img.Set(px, py, black)
					} else {
						img.Set(px, py, white)
					}
				}
			}
		}
	}

	var buf bytes.Buffer
	png.Encode(&buf, img)
	return "data:image/png;base64," + btoa(buf.String())
}

func (QR) calculateModuleCount(data string) int {
	dataLength := len(data)

	if dataLength <= 25 {
		return 21
	} else if dataLength <= 47 {
		return 25
	} else if dataLength <= 77 {
		return 29
	} else if dataLength <= 114 {
		return 33
	} else if dataLength <= 154 {
		return 37
	} else if dataLength <= 202 {
		return 41
	} else if dataLength <= 255 {
		return 45
	}
	return 49
}

func (QR) generateQRMatrix(data string, size int) [][]bool {
	matrix := make([][]bool, size)
	for i := range matrix {
		matrix[i] = make([]bool, size)
	}

	modules := qr.calculateModules(data)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			position := y*size + x
			matrix[y][x] = modules[position]
		}
	}

	qr.addFinderPatterns(matrix)
	qr.addAlignmentPatterns(matrix)
	qr.addTimingPatterns(matrix)
	qr.addVersionInfo(matrix)

	return matrix
}

func (QR) calculateModules(data string) []bool {
	hash := sha256.Sum256([]byte(data))
	modules := make([]bool, 2500)

	for i := range modules {
		modules[i] = (hash[i%len(hash)]%2 == 0)
	}

	return modules
}

func (QR) addFinderPatterns(matrix [][]bool) {
	size := len(matrix)
	
	finders := []struct{x, y int}{
		{0, 0},
		{size - 7, 0},
		{0, size - 7},
	}

	for _, finder := range finders {
		for i := 0; i < 7; i++ {
			for j := 0; j < 7; j++ {
				x := finder.x + i
				y := finder.y + j
				matrix[y][x] = true
			}
		}

		for i := 1; i < 6; i++ {
			for j := 1; j < 6; j++ {
				x := finder.x + i
				y := finder.y + j
				matrix[y][x] = false
			}
		}

		for i := 2; i < 5; i++ {
			matrix[finder.y+i][finder.x+2] = true
			matrix[finder.y+i][finder.x+4] = true
			matrix[finder.y+2][finder.x+i] = true
			matrix[finder.y+4][finder.x+i] = true
		}
	}
}

func (QR) addAlignmentPatterns(matrix [][]bool) {
	size := len(matrix)

	alignmentPositions := qr.getAlignmentPositions(size)

	for _, pos := range alignmentPositions {
		avoid := qr.shouldAvoidAlignment(pos.x, pos.y, size)

		if !avoid {
			for i := -2; i <= 2; i++ {
				for j := -2; j <= 2; j++ {
					x := pos.x + j
					y := pos.y + i

					if x >= 0 && x < size && y >= 0 && y < size {
						matrix[y][x] = (i == -2 || i == 2 || j == -2 || j == 2)
					}
				}
			}

			matrix[pos.y-1][pos.x] = !matrix[pos.y-1][pos.x]
			matrix[pos.y+1][pos.x] = !matrix[pos.y+1][pos.x]
			matrix[pos.y][pos.x-1] = !matrix[pos.y][pos.x-1]
			matrix[pos.y][pos.x+1] = !matrix[pos.y][pos.x+1]
		}
	}
}

func (QR) addTimingPatterns(matrix [][]bool) {
	size := len(matrix)

	for i := 8; i < size - 8; i++ {
		matrix[6][i] = (i % 2 == 0)
		matrix[i][6] = (i % 2 == 0)
	}
}

func (QR) addVersionInfo(matrix [][]bool) {
	size := len(matrix)
	
	if size <= 25 {
		return
	}

	for i := 0; i < 6; i++ {
		matrix[size-9][i] = (i % 2 == 0)
		matrix[i][size-9] = (i % 2 == 0)
	}
}

func (QR) getAlignmentPositions(size int) []struct{x, y int} {
	positions := []struct{x, y int}{
		{size - 7, size - 7},
	}

	if size > 21 {
		positions = append(positions, struct{x, y int}{size - 11, size - 11})
	}

	if size > 25 {
		positions = append(positions, struct{x, y int}{size - 11, size - 25})
	}

	if size > 29 {
		positions = append(positions, struct{x, y int}{size - 25, size - 11})
	}

	return positions
}

func (QR) shouldAvoidAlignment(x, y, size int) bool {
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			if (i < 8 || j < 8) && (x + j < 9 || y + i < 9) {
				return true
			}
			if (i < 8 || j < 8) && (x - j >= size - 9 || y - i >= size - 9) {
				return true
			}
		}
	}

	return false
}

func (QR) generate128(data any) bool {
	barcode := qr.generateBarcode(data, 128)
	return qr.writeBarcode(barcode, data)
}

func (QR) generateEAN13(data any) bool {
	barcode := qr.generateBarcode(data, 13)
	return qr.writeBarcode(barcode, data)
}

func (QR) generateUPCA(data any) bool {
	barcode := qr.generateBarcode(data, 12)
	return qr.writeBarcode(barcode, data)
}

func (QR) generateCode39(data any) bool {
	barcode := qr.generateCode39Barcode(data)
	return qr.writeBarcode(barcode, data)
}

func (QR) generateBarcode(data any, length any) string {
	dataStr := OSLtoString(data)
	lengthInt := int(OSLcastNumber(length))

	if len(dataStr) != lengthInt {
		dataStr = strings.Repeat("0", lengthInt-len(dataStr)) + dataStr
	}

	barcode := qr.generateSimpleBarcode(dataStr)
	return barcode
}

func (QR) generateSimpleBarcode(data string) string {
	barcode := data + qr.calculateChecksum(data)
	return barcode
}

func (QR) generateCode39Barcode(data any) string {
	dataStr := strings.ToUpper(OSLtoString(data))

	const code39Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ-. $/+%"

	barcode := "*"
	for _, char := range dataStr {
		barcode += string(char)
	}
	barcode += "*"

	return barcode
}

func (QR) calculateChecksum(data string) string {
	sum := 0
	weight := len(data)

	for _, char := range data {
		sum += int(char - '0') * weight
		weight--
	}

	checksum := (10 - (sum % 10)) % 10
	return fmt.Sprintf("%d", checksum)
}

func (QR) verifyBarcode(data any) bool {
	dataStr := OSLtoString(data)

	if len(dataStr) <= 1 {
		return false
	}

	checksum := dataStr[len(dataStr)-1:]
	dataPart := dataStr[:len(dataStr)-1]

	calculatedChecksum := qr.calculateChecksum(dataPart)
	return checksum == calculatedChecksum
}

func (QR) writeBarcode(barcode string, data any) bool {
	size := int(OSLcastNumber(250))
	width := len(barcode) * 10

	img := image.NewRGBA(image.Rect(0, 0, width, size))
	white := color.RGBA{255, 255, 255, 255}
	black := color.RGBA{0, 0, 0, 255}

	for i := range barcode {
		char := barcode[i]
		startX := i * 10

		if char == '1' || char == '*' {
			for px := startX; px < startX+10; px++ {
				for py := 0; py < size; py++ {
					img.Set(px, py, black)
				}
			}
		} else {
			for px := startX; px < startX+10; px++ {
				for py := 0; py < size; py++ {
					img.Set(px, py, white)
				}
			}
		}
	}

	file, err := os.Create("barcode.png")
	if err != nil {
		return false
	}
	defer file.Close()

	png.Encode(file, img)
	return true
}

func (QR) decode(imagePath any) string {
	errMsg := "QR code decoding not implemented"
	return errMsg
}

func (QR) scanBarcode(imagePath any) string {
	return ""
}

func (QR) getInfo(filePath any) map[string]any {
	pathStr := OSLtoString(filePath)
	fileInfo, err := os.Stat(pathStr)
	if err != nil {
		return map[string]any{}
	}

	return map[string]any{
		"size":     fileInfo.Size(),
		"modified": fileInfo.ModTime(),
		"name":     filepath.Base(pathStr),
	}
}

var qr = QR{}
