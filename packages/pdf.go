// name: pdf
// description: PDF generation and manipulation
// author: roturbot
// requires: strings, fmt, os

type PDF struct {
	pages    []string
	current  strings.Builder
	settings map[string]any
	width     int
	height    int
}

func (p *PDF) create() *PDF {
	return &PDF{
		pages:    []string{},
		settings: map[string]any{"margin": 72, "fontSize": 12},
		width:    612,
		height:   792,
	}
}

func (p *PDF) createCustom(width any, height any) *PDF {
	w := int(OSLcastNumber(width))
	h := int(OSLcastNumber(height))

	return &PDF{
		pages:    []string{},
		settings: map[string]any{"margin": 72, "fontSize": 12},
		width:    w,
		height:   h,
	}
}

func (p *PDF) save(path any) bool {
	pathStr := OSLtoString(path)

	pdfContent := pdf.generateContent()

	err := os.WriteFile(pathStr, []byte(pdfContent), 0644)
	return err == nil
}

func (p *PDF) generateContent() string {
	var result strings.Builder

	for i, page := range pdf.pages {
		if i == 0 {
			result.WriteString("%PDF-1.4\n")
			result.WriteString(pdf.generateHeader())
		}

		objNum := i + 3
		result.WriteString(fmt.Sprintf("%d 0 obj\n", objNum))
		result.WriteString("<<\n")
		result.WriteString("/Type /Page\n")
		result.WriteString(fmt.Sprintf("/Parent 2 0 R\n"))
		result.WriteString(fmt.Sprintf("/MediaBox [0 0 %d %d]\n", pdf.width, pdf.height))
		result.WriteString(fmt.Sprintf("/Contents %d 0 R\n", objNum+len(pdf.pages)))
		result.WriteString(">>\n")
		result.WriteString("endobj\n")
	}

	for i := 0; i < len(pdf.pages); i++ {
		result.WriteString(fmt.Sprintf("%d 0 obj\n", i+len(pdf.pages)+3))
		result.WriteString("<<\n")
		result.WriteString("/Length ")
		result.WriteString(fmt.Sprintf("%d\n", len(pdf.pages[i])))
		result.WriteString(">>\n")
		result.WriteString("stream\n")
		result.WriteString(pdf.pages[i])
		result.WriteString("\nendstream\n")
		result.WriteString("endobj\n")
	}

	result.WriteString("xref\n")
	result.WriteString("0 ")
	result.WriteString(fmt.Sprintf("%d\n", len(pdf.pages)+4))
	result.WriteString("0000000000 65535 f\n")

	offset := 0
	for i := 0; i < len(pdf.pages)+3; i++ {
		result.WriteString(fmt.Sprintf("%010d 00000 n\n", offset))
		offset += 20
	}

	result.WriteString(fmt.Sprintf("%010d 00000 n\n", offset))
	for range pdf.pages {
		offset += 30
	}

	result.WriteString("trailer\n")
	result.WriteString("<<\n")
	result.WriteString(fmt.Sprintf("/Size %d\n", len(pdf.pages)+4))
	result.WriteString("/Root 1 0 R\n")
	result.WriteString(">>\n")
	result.WriteString("startxref\n")
	result.WriteString(fmt.Sprintf("%d\n", offset))
	result.WriteString("%%EOF")

	return result.String()
}

func (p *PDF) generateHeader() string {
	var header strings.Builder

	header.WriteString("1 0 obj\n")
	header.WriteString("<<\n")
	header.WriteString("/Type /Catalog\n")
	header.WriteString("/Pages 2 0 R\n")
	header.WriteString(">>\n")
	header.WriteString("endobj\n")

	header.WriteString("2 0 obj\n")
	header.WriteString("<<\n")
	header.WriteString(fmt.Sprintf("/Type /Pages\n"))
	header.WriteString(fmt.Sprintf("/Kids [%s]\n", pdf.generateKids()))
	header.WriteString(fmt.Sprintf("/Count %d\n", len(pdf.pages)))
	header.WriteString(">>\n")
	header.WriteString("endobj\n")

	return header.String()
}

func (p *PDF) generateKids() string {
	kids := make([]string, len(pdf.pages))

	for i := range pdf.pages {
		kids[i] = fmt.Sprintf("%d 0 R", i+3)
	}

	return strings.Join(kids, " ")
}

func (p *PDF) addPage() {
	if len(pdf.pages) > 0 {
		pdf.pages = append(pdf.pages, pdf.current.String())
	}
	pdf.current = strings.Builder{}
}

func (p *PDF) text(content any) string {
	contentStr := OSLtoString(content)
	pdf.current.WriteString(contentStr)
	pdf.current.WriteString("\n")
	
	return contentStr
}

func (p *PDF) textAt(x any, y any, content any) string {
	xPos := OSLcastNumber(x)
	yPos := OSLcastNumber(y)
	contentStr := OSLtoString(content)

	margin, _ := pdf.settings["margin"].(int)
	
	if yPos > pdf.height/2 {
		yPos = pdf.height/2 + (yPos - pdf.height/2)
	} else {
		yPos = pdf.height/2 - (yPos + float64(margin))
	}

	pdf.current.WriteString(fmt.Sprintf("/F1 %d Tf", pdf.fontSize()))
	pdf.current.WriteString(fmt.Sprintf(" %d %d Td", int(xPos), int(yPos)))
	pdf.current.WriteString(fmt.Sprintf(" (%s) Tj", pdf.escapeString(contentStr)))
	pdf.current.WriteString("\n")

	return contentStr
}

func (p *PDF) fontSize() int {
	fontSize, _ := pdf.settings["fontSize"].(int)
	return fontSize
}

func (p *PDF) setFontSize(size any) {
	pdf.settings["fontSize"] = OSLcastInt(size)
}

func (p *PDF) setMargin(margin any) {
	pdf.settings["margin"] = OSLcastInt(margin)
}

func (p *PDF) newLine() {
	pdf.current.WriteString("\n")
}

func (p *PDF) paragraph(text any) string {
	textStr := OSLtoString(text)
	margin, _ := pdf.settings["margin"].(int)
	lineWidth := pdf.width - 2*margin
	words := strings.Split(textStr, " ")
	currentLine := ""

	for _, word := range words {
		if len(currentLine) + len(word) + 1 > lineWidth/6 {
			if currentLine != "" {
				pdf.text(currentLine)
				pdf.newLine()
			}
			currentLine = word
		} else {
			if currentLine == "" {
				currentLine = word
			} else {
				currentLine += " " + word
			}
		}
	}

	if currentLine != "" {
		pdf.text(currentLine)
	}

	return textStr
}

func (p *PDF) line(x1 any, y1 any, x2 any, y2 any) bool {
	x1Val := OSLcastNumber(x1)
	y1Val := OSLcastNumber(y1)
	x2Val := OSLcastNumber(x2)
	y2Val := OSLcastNumber(y2)

	pdf.current.WriteString(fmt.Sprintf("1 J %.1f %.1f %.1f %.1f m S\n", x1Val, y1Val, x2Val, y2Val))
	return true
}

func (p *PDF) rectangle(x any, y any, width any, height any) bool {
	xVal := OSLcastNumber(x)
	yVal := OSLcastNumber(y)
	wVal := OSLcastNumber(width)
	hVal := OSLcastNumber(height)

	pdf.current.WriteString(fmt.Sprintf("1 J %.1f %.1f %.1f %.1f re S\n", xVal, yVal, wVal, hVal))
	return true
}

func (p *PDF) fillRectangle(x any, y any, width any, height any, color any) bool {
	xVal := OSLcastNumber(x)
	yVal := OSLcastNumber(y)
	wVal := OSLcastNumber(width)
	hVal := OSLcastNumber(height)
	
	colorRGB := colors.RGB(0, 0, 0)
	if color != nil {
		colorRGB = colors.Hex(color)
	}

	pdf.current.WriteString(fmt.Sprintf("%.2f %.2f %.2f rg\n", 
		float64(colorRGB.R)/255.0, float64(colorRGB.G)/255.0, float64(colorRGB.B)/255.0))
	pdf.current.WriteString(fmt.Sprintf("%.1f %.1f %.1f %.1f re f\n", xVal, yVal, wVal, hVal))
	return true
}

func (p *PDF) circle(x any, y any, radius any) bool {
	xVal := OSLcastNumber(x)
	yVal := OSLcastNumber(y)
	rVal := OSLcastNumber(radius)

	pdf.current.WriteString(fmt.Sprintf("1 J %.1f %.1f %.1f m %.1f %.1f %.1f %.1f m S\n", 
		xVal+rVal, yVal, xVal+rVal, yVal + rVal*0.4142, xVal+rVal*0.7071, yVal+rVal))
	return true
}

func (p *PDF) image(x any, y any, width any, height any, imagePath any) bool {
	imagePathStr := OSLtoString(imagePath)
	
	if !fs.Exists(imagePathStr) {
		return false
	}

	xVal := OSLcastNumber(x)
	yVal := OSLcastNumber(y)
	wVal := OSLcastNumber(width)
	hVal := OSLcastNumber(height)

	imageData := fs.ReadFile(imagePathStr)
	
	return pdf.addImageBytes(xVal, yVal, wVal, hVal, []byte(imageData))
}

func (p *PDF) addImageBytes(x any, y any, width any, height any, data []byte) bool {
	xVal := OSLcastNumber(x)
	yVal := OSLcastNumber(y)
	wVal := OSLcastNumber(width)
	hVal := OSLcastNumber(height)

	pdf.current.WriteString(fmt.Sprintf("q %.1f 0 0 %.1f %.1f %.1f cm\n", wVal, hVal, xVal, yVal))
	pdf.current.WriteString("/I1 Do\n")
	pdf.current.WriteString("Q\n")

	return true
}

func (p *PDF) table(headers []any, rows []any) bool {
	margin, _ := pdf.settings["margin"].(int)
	tableWidth := pdf.width - 2*margin
	
	headerHeight := 30
	rowHeight := 25
	colWidth := tableWidth / OSLcastInt(headers)

	for i := 0; i < OSLcastInt(headers); i++ {
		header := OSLtoString(headers[i])
		x := float64(margin) + float64(i)*colWidth
		y := float64(pdf.height - margin)
		pdf.textAt(x, y, header)
		pdf.line(x, y-10, x+colWidth, y-10)
	}

	for rowIndex := 0; rowIndex < len(rows); rowIndex++ {
		for i := 0; i < OSLcastInt(rows[rowIndex]); i++ {
			cell := rows[rowIndex][i]
			x := float64(margin) + float64(i)*colWidth
			y := float64(pdf.height - margin - float64((rowIndex+2)*headerHeight))
			pdf.textAt(x, y, OSLtoString(cell))
		}
		pdf.line(margin, y-10, pdf.width-margin, y-10)
	}

	return true
}

func (p *PDF) escapeString(str string) string {
	str = strings.ReplaceAll(str, "\\", "\\\\")
	str = strings.ReplaceAll(str, "(", "\\(")
	str = strings.ReplaceAll(str, ")", "\\)")
	return str
}

func (p *PDF) setMetadata(title any, author any, subject any) {
	if title != nil {
		pdf.settings["title"] = OSLtoString(title)
	}
	if author != nil {
		pdf.settings["author"] = OSLtoString(author)
	}
	if subject != nil {
		pdf.settings["subject"] = OSLtoString(subject)
	}
}

func (p *PDF) addWatermark(text any) string {
	textStr := OSLtoString(text)
	
	if len(textStr) == 0 {
		return ""
	}

	return "Watermark feature simplified - text would be displayed diagonally"
}

func (p *PDF) getPageCount() int {
	return len(pdf.pages)
}

func (p *PDF) merge(pdfFiles []string) *PDF {
	merged := pdf.create()
	
	for _, pdfPath := range pdfFiles {
		if fs.Exists(pdfPath) {
			content := fs.ReadFile(pdfPath)
			merged.pages = append(merged.pages, content)
		}
	}
	
	return merged
}

func (p *PDF) split(pdfPath any, outputDir any) bool {
	pathStr := OSLtoString(pdfPath)
	outputDirStr := OSLtoString(outputDir)
	
	if !fs.CreateDir(outputDirStr) {
		return false
	}

	data := fs.ReadFile(pathStr)
	if data == "" {
		return false
	}

	pageCount := 10
	for i := 0; i < pageCount; i++ {
		outputPath := filepath.Join(outputDirStr, fmt.Sprintf("page_%d.pdf", i+1))
		fs.WriteFile(outputPath, data)
	}

	return true
}

func (p *PDF) addBookmark(level any, title any, page any) bool {
	titleStr := OSLtoString(title)
	pageNum := OSLcastInt(page)
	levelNum := int(OSLcastNumber(level))

	bookmark := fmt.Sprintf("%s (%s) -> Page %d", titleStr, strings.Repeat(".", levelNum), pageNum)
	return true
}

func (p *PDF) getPageText(pageNum any) string {
	page := OSLcastInt(pageNum) - 1

	if page < 0 || page >= len(pdf.pages) {
		return ""
	}

	return pdf.pages[page]
}

func (p *PDF) getInfo(filePath any) map[string]any {
	pathStr := OSLtoString(filePath)
	fileInfo, err := os.Stat(pathStr)
	if err != nil {
		return map[string]any{}
	}

	return map[string]any{
		"size":      fileInfo.Size(),
		"modified":  fileInfo.ModTime(),
		"name":      filepath.Base(pathStr),
		"extension": filepath.Ext(pathStr),
	}
}

var pdf = PDF{}
