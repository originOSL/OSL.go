// name: tui
// description: Text User Interface utilities
// author: roturbot
// requires: fmt, os, strings, time

type TUI struct{}

const (
	ANSI_RESET       = "\033[0m"
	ANSI_CLEAR       = "\033[2J"
	ANSI_CLEAR_LINE  = "\033[2K"
	ANSI_HOME        = "\033[H"
	ANSI_BOLD        = "\033[1m"
	ANSI_DIM         = "\033[2m"
	ANSI_UNDERLINE   = "\033[4m"
	ANSI_BLINK       = "\033[5m"
	ANSI_REVERSE     = "\033[7m"
	ANSI_HIDDEN      = "\033[8m"
	ANSI_STRIKETHROUGH = "\033[9m"

	ANSI_BLACK   = "\033[30m"
	ANSI_RED     = "\033[31m"
	ANSI_GREEN   = "\033[32m"
	ANSI_YELLOW  = "\033[33m"
	ANSI_BLUE    = "\033[34m"
	ANSI_MAGENTA = "\033[35m"
	ANSI_CYAN    = "\033[36m"
	ANSI_WHITE   = "\033[37m"
	ANSI_DEFAULT = "\033[39m"

	ANSI_BG_BLACK   = "\033[40m"
	ANSI_BG_RED     = "\033[41m"
	ANSI_BG_GREEN   = "\033[42m"
	ANSI_BG_YELLOW  = "\033[43m"
	ANSI_BG_BLUE    = "\033[44m"
	ANSI_BG_MAGENTA = "\033[45m"
	ANSI_BG_CYAN    = "\033[46m"
	ANSI_BG_WHITE   = "\033[47m"
	ANSI_BG_DEFAULT = "\033[49m"

	ANSI_BRIGHT_BLACK   = "\033[90m"
	ANSI_BRIGHT_RED     = "\033[91m"
	ANSI_BRIGHT_GREEN   = "\033[92m"
	ANSI_BRIGHT_YELLOW  = "\033[93m"
	ANSI_BRIGHT_BLUE    = "\033[94m"
	ANSI_BRIGHT_MAGENTA = "\033[95m"
	ANSI_BRIGHT_CYAN    = "\033[96m"
	ANSI_BRIGHT_WHITE   = "\033[97m"

	ANSI_BRIGHT_BG_BLACK   = "\033[100m"
	ANSI_BRIGHT_BG_RED     = "\033[101m"
	ANSI_BRIGHT_BG_GREEN   = "\033[102m"
	ANSI_BRIGHT_BG_YELLOW  = "\033[103m"
	ANSI_BRIGHT_BG_BLUE    = "\033[104m"
	ANSI_BRIGHT_BG_MAGENTA = "\033[105m"
	ANSI_BRIGHT_BG_CYAN    = "\033[106m"
	ANSI_BRIGHT_BG_WHITE   = "\033[107m"

	ANSI_CURSOR_UP    = "\033["
	ANSI_CURSOR_DOWN  = "\033["
	ANSI_CURSOR_RIGHT = "\033["
	ANSI_CURSOR_LEFT  = "\033["
	ANSI_CURSOR_HOME  = "\033[H"
	ANSI_CURSOR_SAVE  = "\033[s"
	ANSI_CURSOR_RESTORE = "\033[u"
)

func (TUI) Clear() {
	fmt.Print(ANSI_CLEAR + ANSI_HOME)
}

func (TUI) ClearLine() {
	fmt.Print(ANSI_CLEAR_LINE)
}

func (TUI) ClearLines(count any) {
	n := OSLcastInt(count)
	for i := 0; i < n; i++ {
		fmt.Print("\033[2K\033[1A")
	}
}

func (TUI) MoveCursor(x any, y any) {
	xFmt := OSLtoString(x)
	yFmt := OSLtoString(y)
	fmt.Printf("\033[%s;%sH", yFmt, xFmt)
}

func (TUI) MoveUp(n any) {
	fmt.Printf("\033[%sA", OSLtoString(n))
}

func (TUI) MoveDown(n any) {
	fmt.Printf("\033[%sB", OSLtoString(n))
}

func (TUI) MoveRight(n any) {
	fmt.Printf("\033[%sC", OSLtoString(n))
}

func (TUI) MoveLeft(n any) {
	fmt.Printf("\033[%sD", OSLtoString(n))
}

func (TUI) SaveCursor() {
	fmt.Print(ANSI_CURSOR_SAVE)
}

func (TUI) RestoreCursor() {
	fmt.Print(ANSI_CURSOR_RESTORE)
}

func (TUI) HideCursor() {
	fmt.Print("\033[?25l")
}

func (TUI) ShowCursor() {
	fmt.Print("\033[?25h")
}

func (TUI) Color(colorName string, text any) string {
	textStr := OSLtoString(text)
	var colorCode string

	switch strings.ToLower(colorName) {
	case "black":
		colorCode = ANSI_BLACK
	case "red":
		colorCode = ANSI_RED
	case "green":
		colorCode = ANSI_GREEN
	case "yellow":
		colorCode = ANSI_YELLOW
	case "blue":
		colorCode = ANSI_BLUE
	case "magenta", "purple":
		colorCode = ANSI_MAGENTA
	case "cyan":
		colorCode = ANSI_CYAN
	case "white":
		colorCode = ANSI_WHITE
	case "brightblack", "gray", "grey":
		colorCode = ANSI_BRIGHT_BLACK
	case "brightred":
		colorCode = ANSI_BRIGHT_RED
	case "brightgreen":
		colorCode = ANSI_BRIGHT_GREEN
	case "brightyellow":
		colorCode = ANSI_BRIGHT_YELLOW
	case "brightblue":
		colorCode = ANSI_BRIGHT_BLUE
	case "brightmagenta", "brightpurple":
		colorCode = ANSI_BRIGHT_MAGENTA
	case "brightcyan":
		colorCode = ANSI_BRIGHT_CYAN
	case "brightwhite":
		colorCode = ANSI_BRIGHT_WHITE
	default:
		return textStr
	}

	return colorCode + textStr + ANSI_RESET
}

func (TUI) BgColor(colorName string, text any) string {
	textStr := OSLtoString(text)
	var colorCode string

	switch strings.ToLower(colorName) {
	case "black":
		colorCode = ANSI_BG_BLACK
	case "red":
		colorCode = ANSI_BG_RED
	case "green":
		colorCode = ANSI_BG_GREEN
	case "yellow":
		colorCode = ANSI_BG_YELLOW
	case "blue":
		colorCode = ANSI_BG_BLUE
	case "magenta", "purple":
		colorCode = ANSI_BG_MAGENTA
	case "cyan":
		colorCode = ANSI_BG_CYAN
	case "white":
		colorCode = ANSI_BG_WHITE
	case "brightblack", "gray", "grey":
		colorCode = ANSI_BRIGHT_BG_BLACK
	case "brightred":
		colorCode = ANSI_BRIGHT_BG_RED
	case "brightgreen":
		colorCode = ANSI_BRIGHT_BG_GREEN
	case "brightyellow":
		colorCode = ANSI_BRIGHT_BG_YELLOW
	case "brightblue":
		colorCode = ANSI_BRIGHT_BG_BLUE
	case "brightmagenta", "brightpurple":
		colorCode = ANSI_BRIGHT_BG_MAGENTA
	case "brightcyan":
		colorCode = ANSI_BRIGHT_BG_CYAN
	case "brightwhite":
		colorCode = ANSI_BRIGHT_BG_WHITE
	default:
		return textStr
	}

	return colorCode + textStr + ANSI_RESET
}

func (TUI) Style(styleName string, text any) string {
	textStr := OSLtoString(text)
	var styleCode string

	switch strings.ToLower(styleName) {
	case "bold":
		styleCode = ANSI_BOLD
	case "dim":
		styleCode = ANSI_DIM
	case "underline":
		styleCode = ANSI_UNDERLINE
	case "blink":
		styleCode = ANSI_BLINK
	case "reverse", "invert":
		styleCode = ANSI_REVERSE
	case "hidden":
		styleCode = ANSI_HIDDEN
	case "strikethrough":
		styleCode = ANSI_STRIKETHROUGH
	default:
		return textStr
	}

	return styleCode + textStr + ANSI_RESET
}

func (TUI) RgbColor(r any, g any, b any, text any) string {
	rr := OSLcastInt(r)
	gg := OSLcastInt(g)
	bb := OSLcastInt(b)
	textStr := OSLtoString(text)
	return fmt.Sprintf("\033[38;2;%d;%d;%dm%s%s", rr, gg, bb, textStr, ANSI_RESET)
}

func (TUI) RgbBg(r any, g any, b any, text any) string {
	rr := OSLcastInt(r)
	gg := OSLcastInt(g)
	bb := OSLcastInt(b)
	textStr := OSLtoString(text)
	return fmt.Sprintf("\033[48;2;%d;%d;%dm%s%s", rr, gg, bb, textStr, ANSI_RESET)
}

func (TUI) Progress(current any, total any, width any) string {
	curr := OSLcastNumber(current)
	tot := OSLcastNumber(total)
	w := OSLcastInt(width)

	if tot <= 0 {
		return strings.Repeat(" ", w) + " 0%"
	}

	percent := (curr / tot) * 100
	if percent > 100 {
		percent = 100
	}

	fillWidth := int((float64(percent) / 100) * float64(w))
	if fillWidth > w {
		fillWidth = w
	}

	fill := strings.Repeat("█", fillWidth)
	empty := strings.Repeat("░", w-fillWidth)

	return fmt.Sprintf("[%s%s] %.0f%%", fill, empty, percent)
}

func (TUI) Spinner(finished bool) string {
	symbols := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	if finished {
		return "✓"
	}
	index := int(time.Now().UnixNano() / 100000000) % len(symbols)
	return symbols[index]
}

func (TUI) Horizontal(width any) string {
	w := OSLcastInt(width)
	if w <= 0 {
		w = 80
	}
	return strings.Repeat("─", w)
}

func (TUI) Vertical(height any) []any {
	h := OSLcastInt(height)
	if h <= 0 {
		h = 10
	}
	lines := make([]any, h)
	for i := 0; i < h; i++ {
		lines[i] = "│"
	}
	return lines
}

func (TUI) Box(title any, content any) string {
	titleStr := strings.TrimSpace(OSLtoString(title))
	contentStr := OSLtoString(content)
	contentLines := strings.Split(contentStr, "\n")

	maxWidth := 0
	if titleStr != "" {
		maxWidth = len(titleStr) + 4
	}
	for _, line := range contentLines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	top := "┌"
	if titleStr != "" {
		titlePadding := maxWidth - (len(titleStr) + 3)
		top += "─ " + titleStr + " "
		if titlePadding > 0 {
			top += strings.Repeat("─", titlePadding)
		}
	} else {
		top += strings.Repeat("─", maxWidth)
	}
	top += "┐"

	bottom := "└" + strings.Repeat("─", maxWidth) + "┘"

	var box strings.Builder
	box.WriteString(top + "\n")

	for _, line := range contentLines {
		padding := maxWidth - len(line) - 1
		box.WriteString("│ " + line + strings.Repeat(" ", padding) + " │\n")
	}

	box.WriteString(bottom)

	return box.String()
}
	for _, line := range contentLines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	top := "┌"
	if titleStr != "" {
		titlePadding := maxWidth - (len(titleStr) + 3)
		top += "─ " + titleStr + " "
		if titlePadding > 0 {
			top += strings.Repeat("─", titlePadding)
		}
	} else {
		top += strings.Repeat("─", maxWidth)
	}
	top += "┐"

	bottom := "└" + strings.Repeat("─", maxWidth) + "┘"

	var box strings.Builder
	box.WriteString(top + "\n")

	for _, line := range contentLines {
		padding := maxWidth - len(line)
		box.WriteString("│ " + line + strings.Repeat(" ", padding) + " │\n")
	}

	box.WriteString(bottom)

	return box.String()
}

func (TUI) DrawBox(x any, y any, width any, height any, title any) {
	px := OSLcastInt(x)
	py := OSLcastInt(y)
	w := OSLcastInt(width)
	h := OSLcastInt(height)
	titleStr := strings.TrimSpace(OSLtoString(title))

	if w < 3 {
		w = 3
	}
	if h < 2 {
		h = 2
	}

	fmt.Print("\033[s")
	fmt.Printf("\033[%d;%dH", py+1, px+1)

	if titleStr == "" {
		fmt.Print("┌" + strings.Repeat("─", w-2) + "┐")
	} else {
		titleLen := len(titleStr)
		if titleLen > w-6 {
			titleStr = titleStr[:w-6]
			titleLen = w - 6
		}
		leftSide := (w - titleLen - 4) / 2
		rightSide := w - titleLen - 4 - leftSide
		fmt.Print("┌" + strings.Repeat("─", leftSide) + "┤ " + titleStr + " ├" + strings.Repeat("─", rightSide) + "┐")
	}

	for i := 1; i < h-1; i++ {
		fmt.Printf("\033[%d;%dH", py+1+i, px+1)
		fmt.Print("│" + strings.Repeat(" ", w-2) + "│")
	}

	fmt.Printf("\033[%d;%dH", py+h, px+1)
	fmt.Print("└" + strings.Repeat("─", w-2) + "┘")
	fmt.Print("\033[u")
}

func (TUI) Table(headers []any, rows []any) string {
	var table strings.Builder

	headerStrs := make([]string, len(headers))
	maxWidths := make([]int, len(headers))

	for i, h := range headers {
		headerStrs[i] = OSLtoString(h)
		maxWidths[i] = len(headerStrs[i])
	}

	for _, row := range rows {
		rowArr := OSLcastArray(row)
		for i, cell := range rowArr {
			cellStr := OSLtoString(cell)
			if i < len(maxWidths) {
				if len(cellStr) > maxWidths[i] {
					maxWidths[i] = len(cellStr)
				}
			}
		}
	}

	border := "+"
	for _, w := range maxWidths {
		border += strings.Repeat("-", w+2) + "+"
	}
	table.WriteString(border + "\n")

	headerRow := "|"
	for i, h := range headerStrs {
		padding := maxWidths[i] - len(h)
		headerRow += fmt.Sprintf(" %s%s |", h, strings.Repeat(" ", padding))
	}
	table.WriteString(headerRow + "\n")

	table.WriteString(border + "\n")

	for _, row := range rows {
		rowArr := OSLcastArray(row)
		dataRow := "|"
		for i, cell := range rowArr {
			if i < len(maxWidths) {
				cellStr := OSLtoString(cell)
				padding := maxWidths[i] - len(cellStr)
				dataRow += fmt.Sprintf(" %s%s |", cellStr, strings.Repeat(" ", padding))
			}
		}
		table.WriteString(dataRow + "\n")
	}

	table.WriteString(border)

	return table.String()
}

func (TUI) TableColored(headers []any, rows []any, colorFn any) string {
	var table strings.Builder

	headerStrs := make([]string, len(headers))
	maxWidths := make([]int, len(headers))

	for i, h := range headers {
		headerStrs[i] = OSLtoString(h)
		maxWidths[i] = len(headerStrs[i])
	}

	for _, row := range rows {
		rowArr := OSLcastArray(row)
		for i, cell := range rowArr {
			cellStr := OSLtoString(cell)
			if i < len(maxWidths) {
				if len(cellStr) > maxWidths[i] {
					maxWidths[i] = len(cellStr)
				}
			}
		}
	}

	border := "+"
	for _, w := range maxWidths {
		border += strings.Repeat("-", w+2) + "+"
	}
	table.WriteString(ANSI_CYAN + border + ANSI_RESET + "\n")

	headerRow := "|"
	for i, h := range headerStrs {
		padding := maxWidths[i] - len(h)
		headerRow += fmt.Sprintf(" %s%s |", h, strings.Repeat(" ", padding))
	}
	table.WriteString(ANSI_BOLD + headerRow + ANSI_RESET + "\n")

	table.WriteString(ANSI_CYAN + border + ANSI_RESET + "\n")

	for _, row := range rows {
		rowArr := OSLcastArray(row)
		dataRow := "|"
		for i, cell := range rowArr {
			if i < len(maxWidths) {
				cellStr := OSLtoString(cell)
				padding := maxWidths[i] - len(cellStr)

				var displayCell string
				if OSLisFunc(colorFn) {
					color := OSLtoString(OSLcallFunc(colorFn, nil, []any{cell}))
					var colorCode string
					switch strings.ToLower(color) {
					case "black":
						colorCode = ANSI_BLACK
					case "red":
						colorCode = ANSI_RED
					case "green":
						colorCode = ANSI_GREEN
					case "yellow":
						colorCode = ANSI_YELLOW
					case "blue":
						colorCode = ANSI_BLUE
					case "magenta", "purple":
						colorCode = ANSI_MAGENTA
					case "cyan":
						colorCode = ANSI_CYAN
					case "white":
						colorCode = ANSI_WHITE
					case "brightblack", "gray", "grey":
						colorCode = ANSI_BRIGHT_BLACK
					case "brightred":
						colorCode = ANSI_BRIGHT_RED
					case "brightgreen":
						colorCode = ANSI_BRIGHT_GREEN
					case "brightyellow":
						colorCode = ANSI_BRIGHT_YELLOW
					case "brightblue":
						colorCode = ANSI_BRIGHT_BLUE
					case "brightmagenta", "brightpurple":
						colorCode = ANSI_BRIGHT_MAGENTA
					case "brightcyan":
						colorCode = ANSI_BRIGHT_CYAN
					case "brightwhite":
						colorCode = ANSI_BRIGHT_WHITE
					default:
						colorCode = ""
					}
					displayCell = colorCode + cellStr + ANSI_RESET
				} else {
					displayCell = cellStr
				}

				dataRow += fmt.Sprintf(" %s%s |", displayCell, strings.Repeat(" ", padding-len(displayCell)+len(cellStr)))
			}
		}
		table.WriteString(dataRow + "\n")
	}

	table.WriteString(ANSI_CYAN + border + ANSI_RESET)

	return table.String()
}

func (TUI) Select(prompt any, options []any) any {
	promptStr := OSLtoString(prompt)
	optsArr := OSLcastArray(options)

	if len(optsArr) == 0 {
		return ""
	}

tui.Clear()
	fmt.Println(promptStr)
	fmt.Println()

	for i, opt := range optsArr {
		fmt.Printf("  %d. %s\n", i+1, OSLtoString(opt))
	}
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter selection: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Print("\033[1A")
		tui.ClearLine()
			continue
		}

		input = strings.TrimSpace(input)
		selection := OSLcastInt(input)

		if selection >= 1 && selection <= len(optsArr) {
			return optsArr[selection-1]
		}

		fmt.Print("\033[1A")
		fmt.Print("\033[2K")
	}
}

func (TUI) Confirm(prompt any) bool {
	promptStr := OSLtoString(prompt)
	if !strings.HasSuffix(promptStr, "?") {
		promptStr += "?"
	}
	promptStr += " [y/N]: "

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(promptStr)
		input, err := reader.ReadString('\n')
		if err != nil {
			continue
		}

		input = strings.ToLower(strings.TrimSpace(input))
		if input == "y" || input == "yes" {
			return true
		}
		if input == "n" || input == "no" || input == "" {
			return false
		}
	}
}

func (TUI) Menu(title any, items []any) any {
	titleStr := OSLtoString(title)
	itemsArr := OSLcastArray(items)

	if len(itemsArr) == 0 {
		return ""
	}

	fmt.Print(ANSI_CLEAR + ANSI_HOME)
	if titleStr != "" {
		fmt.Println(ANSI_BOLD + ANSI_CYAN + "═══ "+titleStr+" ═══" + ANSI_RESET)
		fmt.Println()
	}

	for i, item := range itemsArr {
		fmt.Printf("  [%d] %s\n", i+1, OSLtoString(item))
	}
	fmt.Println()
	fmt.Println("  [0] Exit")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Select option: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		selection := OSLcastInt(input)

		if selection == 0 {
			return nil
		}
		if selection >= 1 && selection <= len(itemsArr) {
			return itemsArr[selection-1]
		}
	}
}

func (TUI) Input(prompt any) string {
	promptStr := OSLtoString(prompt)
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(promptStr)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func (TUI) Password(prompt any) string {
	promptStr := OSLtoString(prompt)
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(promptStr)

	var password strings.Builder
tui.HideCursor()

	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			continue
		}

		if r == '\r' || r == '\n' {
			fmt.Println()
		tui.ShowCursor()
			break
		}

		if r == 3 || r == 26 {
			fmt.Println()
		tui.ShowCursor()
			return ""
		}

		if r == 127 || r == 8 {
			if password.Len() > 0 {
				password.Reset()
			}
		} else {
			password.WriteRune(r)
		}

		fmt.Print("*")
		fmt.Print("\033[1D")
	}

	return password.String()
}

func (TUI) Center(text any) string {
	textStr := OSLtoString(text)
	termWidth := 80

	textLines := strings.Split(textStr, "\n")
	var centered []string

	for _, line := range textLines {
		padding := (termWidth - len(line)) / 2
		if padding > 0 {
			centered = append(centered, strings.Repeat(" ", padding)+line)
		} else {
			centered = append(centered, line)
		}
	}

	return strings.Join(centered, "\n")
}

func (TUI) Pad(text any, width any, align any) string {
	textStr := OSLtoString(text)
	w := OSLcastInt(width)
	alignStr := strings.ToLower(OSLtoString(align))

	if len(textStr) >= w {
		return textStr
	}

	padding := w - len(textStr)

	switch alignStr {
	case "right":
		return strings.Repeat(" ", padding) + textStr
	case "center":
		leftPad := padding / 2
		rightPad := padding - leftPad
		return strings.Repeat(" ", leftPad) + textStr + strings.Repeat(" ", rightPad)
	default:
		return textStr + strings.Repeat(" ", padding)
	}
}

func (TUI) Divider(char any, width any, title any) string {
	c := OSLtoString(char)
	if c == "" || len(c) > 1 {
		c = "─"
	}
	w := OSLcastInt(width)
	titleStr := OSLtoString(title)

	if w <= 0 {
		w = 80
	}

	if titleStr == "" {
		return strings.Repeat(c, w)
	}

	titleLen := len(titleStr)
	if titleLen > w-4 {
		titleStr = titleStr[:w-4]
		titleLen = w - 4
	}

	leftLen := (w - titleLen - 2) / 2
	rightLen := w - titleLen - 2 - leftLen

	return strings.Repeat(c, leftLen) + " " + titleStr + " " + strings.Repeat(c, rightLen)
}

func (TUI) Status(status any, message any) string {
	statusStr := strings.ToLower(OSLtoString(status))
	msgStr := OSLtoString(message)

	var icon, color string

	switch statusStr {
	case "success", "ok", "done":
		icon = "✓"
		color = "green"
	case "error", "fail", "failed":
		icon = "✗"
		color = "red"
	case "warning", "warn":
		icon = "⚠"
		color = "yellow"
	case "info":
		icon = "ℹ"
		color = "blue"
	case "loading", "pending":
		icon = "⏳"
		color = "cyan"
	default:
		icon = "•"
		color = "white"
 	}

	var colorCode string

	switch strings.ToLower(color) {
	case "black":
		colorCode = ANSI_BLACK
	case "red":
		colorCode = ANSI_RED
	case "green":
		colorCode = ANSI_GREEN
	case "yellow":
		colorCode = ANSI_YELLOW
	case "blue":
		colorCode = ANSI_BLUE
	case "magenta", "purple":
		colorCode = ANSI_MAGENTA
	case "cyan":
		colorCode = ANSI_CYAN
	case "white":
		colorCode = ANSI_WHITE
	case "brightblack", "gray", "grey":
		colorCode = ANSI_BRIGHT_BLACK
	case "brightred":
		colorCode = ANSI_BRIGHT_RED
	case "brightgreen":
		colorCode = ANSI_BRIGHT_GREEN
	case "brightyellow":
		colorCode = ANSI_BRIGHT_YELLOW
	case "brightblue":
		colorCode = ANSI_BRIGHT_BLUE
	case "brightmagenta", "brightpurple":
		colorCode = ANSI_BRIGHT_MAGENTA
	case "brightcyan":
		colorCode = ANSI_BRIGHT_CYAN
	case "brightwhite":
		colorCode = ANSI_BRIGHT_WHITE
	default:
		colorCode = ANSI_WHITE
	}

	return colorCode + icon + ANSI_RESET + " " + msgStr
}

func (TUI) Frame(text any, width any) string {
	textStr := OSLtoString(text)
	w := OSLcastInt(width)

	if w <= 0 {
		w = len(textStr) + 4
	}

	if w < len(textStr)+4 {
		w = len(textStr) + 4
	}

	top := "╔" + strings.Repeat("═", w-2) + "╗"
	bottom := "╚" + strings.Repeat("═", w-2) + "╝"
	padding := w - len(textStr) - 2

	return top + "\n║ " + textStr + strings.Repeat(" ", padding) + " ║\n" + bottom
}

func (TUI) Grid(items []any, columns any) string {
	itemsArr := OSLcastArray(items)
	cols := OSLcastInt(columns)

	if cols <= 0 {
		cols = 2
	}
	if len(itemsArr) == 0 {
		return ""
	}

	maxWidth := 0
	for _, item := range itemsArr {
		itemStr := OSLtoString(item)
		if len(itemStr) > maxWidth {
			maxWidth = len(itemStr)
		}
	}

	var grid strings.Builder

	for i, item := range itemsArr {
		itemStr := OSLtoString(item)
		padding := maxWidth - len(itemStr)

		if i > 0 && i%cols == 0 {
			grid.WriteString("\n")
		}

		grid.WriteString(fmt.Sprintf("│ %s%s ", itemStr, strings.Repeat(" ", padding)))
	}

	return strings.Join(strings.Split(grid.String(), "\n")[:], "\n")
}

func (TUI) Tree(items []any, prefix any) string {
	itemsArr := OSLcastArray(items)
	prefixStr := OSLtoString(prefix)

	var output []string

	for i, item := range itemsArr {
		itemStr := OSLtoString(item)
		var connector, nextPrefix string

		if i == len(itemsArr)-1 {
			connector = "└── "
			nextPrefix = prefixStr + "    "
		} else {
			connector = "├── "
			nextPrefix = prefixStr + "│   "
		}

		if mapItem, ok := item.(map[string]any); ok {
			name := OSLtoString(mapItem["name"])
			children, hasChildren := mapItem["children"]

			output = append(output, prefixStr+connector+name)

			if hasChildren {
				childrenArr := OSLcastArray(children)
				if len(childrenArr) > 0 {
					output = append(output, TUI{}.Tree(childrenArr, nextPrefix))
				}
			}
		} else {
			output = append(output, prefixStr+connector+itemStr)
		}
	}

	return strings.Join(output, "\n")
}

func (TUI) BarChart(data []any, width any, showLabels any) string {
	dataArr := OSLcastArray(data)
	w := OSLcastInt(width)
	showLabelsBool := OSLcastBool(showLabels)

	if len(dataArr) == 0 {
		return ""
	}

	var bars []map[string]any
	maxValue := 0.0

	for _, item := range dataArr {
		var label string
		var value float64

		if mapItem, ok := item.(map[string]any); ok {
			label = OSLtoString(mapItem["label"])
			value = OSLcastNumber(mapItem["value"])
		} else if arrItem, ok := item.([]any); ok && len(arrItem) >= 2 {
			label = OSLtoString(arrItem[0])
			value = OSLcastNumber(arrItem[1])
		}

		if value > maxValue {
			maxValue = value
		}

		bars = append(bars, map[string]any{
			"label": label,
			"value": value,
		})
	}

	if maxValue == 0 {
		maxValue = 1
	}

	var chart strings.Builder

	if showLabelsBool {
		chart.WriteString(ANSI_GREEN + strings.Repeat("█", w) + ANSI_RESET + "\n")
	} else {
		chart.WriteString(strings.Repeat("█", w) + "\n")
	}

	for _, bar := range bars {
		label := OSLtoString(bar["label"])
		value := OSLcastNumber(bar["value"])
		barWidth := int((value / maxValue) * float64(w))

		line := fmt.Sprintf("%-15s", label)
		if showLabelsBool {
			line += " " + ANSI_GREEN + strings.Repeat("█", barWidth) + ANSI_RESET
		} else {
			line += " " + strings.Repeat("█", barWidth)
		}
		line += fmt.Sprintf(" %.2f", value)

		chart.WriteString(line + "\n")
	}

	return chart.String()
}

var tui = TUI{}


