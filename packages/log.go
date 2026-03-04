// name: log
// description: Logging utilities with levels and formatting
// author: roturbot
// requires: fmt, os, time

type Log struct{}

type LogLevel struct {
	name       string
	prefix     string
	color      string
	colorReset string
}

var (
	INFO    = LogLevel{name: "INFO", prefix: "[INFO] ", color: "\033[36m", colorReset: "\033[0m"}
	WARN    = LogLevel{name: "WARN", prefix: "[WARN] ", color: "\033[33m", colorReset: "\033[0m"}
	ERROR   = LogLevel{name: "ERROR", prefix: "[ERROR] ", color: "\033[31m", colorReset: "\033[0m"}
	DEBUG   = LogLevel{name: "DEBUG", prefix: "[DEBUG] ", color: "\033[90m", colorReset: "\033[0m"}
	SUCCESS = LogLevel{name: "SUCCESS", prefix: "[SUCCESS] ", color: "\033[32m", colorReset: "\033[0m"}
)

var logLevel = INFO

func (Log) setLevel(level any) {
	levelStr := strings.ToUpper(OSLtoString(level))

	switch levelStr {
	case "INFO", "INFORMATION":
		logLevel = INFO
	case "WARN", "WARNING":
		logLevel = WARN
	case "ERROR", "ERR":
		logLevel = ERROR
	case "DEBUG":
		logLevel = DEBUG
	case "SUCCESS", "OK":
		logLevel = SUCCESS
	}
}

func (Log) getLevel() string {
	return logLevel.name
}

func (Log) shouldLog(level LogLevel) bool {
	levels := map[string]int{
		"DEBUG":   0,
		"INFO":    1,
		"WARN":    2,
		"ERROR":   3,
		"SUCCESS": 1,
	}

	currentLevel := levels[logLevel.name]
	targetLevel := levels[level.name]

	return targetLevel >= currentLevel
}

func (Log) info(message any, args ...any) {
	if !log.shouldLog(INFO) {
		return
	}

	timestamp := time.Now().Format("15:04:05")
	msg := OSLtoString(message)

	if len(args) > 0 {
		argsStrs := make([]string, len(args))
		for i, arg := range args {
			argsStrs[i] = OSLtoString(arg)
		}
		msg = fmt.Sprintf(msg, argsStrs...)
	}

	entry := LogEntry{timestamp: timestamp + " " + INFO.prefix, level: INFO.name, message: msg}
	if len(logHistory) < maxHistorySize {
		logHistory = append(logHistory, entry)
	}
	fmt.Printf("%s%s%s%s\n", INFO.color, timestamp+" "+INFO.prefix, msg, INFO.colorReset)
}

func (Log) warn(message any, args ...any) {
	if !log.shouldLog(WARN) {
		return
	}

	timestamp := time.Now().Format("15:04:05")
	msg := OSLtoString(message)

	if len(args) > 0 {
		argsStrs := make([]string, len(args))
		for i, arg := range args {
			argsStrs[i] = OSLtoString(arg)
		}
		msg = fmt.Sprintf(msg, argsStrs...)
	}

	entry := LogEntry{timestamp: timestamp + " " + WARN.prefix, level: WARN.name, message: msg}
	if len(logHistory) < maxHistorySize {
		logHistory = append(logHistory, entry)
	}
	fmt.Printf("%s%s%s%s\n", WARN.color, timestamp+" "+WARN.prefix, msg, WARN.colorReset)
}

func (Log) error(message any, args ...any) {
	if !log.shouldLog(ERROR) {
		return
	}

	timestamp := time.Now().Format("15:04:05")
	msg := OSLtoString(message)

	if len(args) > 0 {
		argsStrs := make([]string, len(args))
		for i, arg := range args {
			argsStrs[i] = OSLtoString(arg)
		}
		msg = fmt.Sprintf(msg, argsStrs...)
	}

	entry := LogEntry{timestamp: timestamp + " " + ERROR.prefix, level: ERROR.name, message: msg}
	if len(logHistory) < maxHistorySize {
		logHistory = append(logHistory, entry)
	}
	fmt.Printf("%s%s%s%s\n", ERROR.color, timestamp+" "+ERROR.prefix, msg, ERROR.colorReset)
}

func (Log) debug(message any, args ...any) {
	if !log.shouldLog(DEBUG) {
		return
	}

	timestamp := time.Now().Format("15:04:05")
	msg := OSLtoString(message)

	if len(args) > 0 {
		argsStrs := make([]string, len(args))
		for i, arg := range args {
			argsStrs[i] = OSLtoString(arg)
		}
		msg = fmt.Sprintf(msg, argsStrs...)
	}

	entry := LogEntry{timestamp: timestamp + " " + DEBUG.prefix, level: DEBUG.name, message: msg}
	if len(logHistory) < maxHistorySize {
		logHistory = append(logHistory, entry)
	}
	fmt.Printf("%s%s%s%s\n", DEBUG.color, timestamp+" "+DEBUG.prefix, msg, DEBUG.colorReset)
}

func (Log) success(message any, args ...any) {
	if !log.shouldLog(SUCCESS) {
		return
	}

	timestamp := time.Now().Format("15:04:05")
	msg := OSLtoString(message)

	if len(args) > 0 {
		argsStrs := make([]string, len(args))
		for i, arg := range args {
			argsStrs[i] = OSLtoString(arg)
		}
		msg = fmt.Sprintf(msg, argsStrs...)
	}

	entry := LogEntry{timestamp: timestamp + " " + SUCCESS.prefix, level: SUCCESS.name, message: msg}
	if len(logHistory) < maxHistorySize {
		logHistory = append(logHistory, entry)
	}
	fmt.Printf("%s%s%s%s\n", SUCCESS.color, timestamp+" "+SUCCESS.prefix, msg, SUCCESS.colorReset)
}

func (Log) log(level any, message any, args ...any) {
	levelStr := strings.ToUpper(OSLtoString(level))

	switch levelStr {
	case "INFO", "INFORMATION":
		log.info(message, args...)
	case "WARN", "WARNING":
		log.warn(message, args...)
	case "ERROR", "ERR":
		log.error(message, args...)
	case "DEBUG":
		log.debug(message, args...)
	case "SUCCESS", "OK":
		log.success(message, args...)
	default:
		log.info(message, args...)
	}
}

func (Log) plain(message any, args ...any) {
	msg := OSLtoString(message)

	if len(args) > 0 {
		argsStrs := make([]string, len(args))
		for i, arg := range args {
			argsStrs[i] = OSLtoString(arg)
		}
		fmt.Printf(msg+"\n", argsStrs...)
	} else {
		fmt.Println(msg)
	}
}

func (Log) json(data any) {
	jsonStr := JsonFormat(data)
	fmt.Println(jsonStr)
}

func (Log) table(headers []any, rows []any) {
	fmt.Println(tui.Table(headers, rows))
}

func (Log) separator() {
	fmt.Println(strings.Repeat("-", 80))
}

func (Log) clear() {
	fmt.Print("\033[2J\033[H")
}

func (Log) time(message any) {
	timestamp := time.Now().Format("15:04:05.000")
	fmt.Printf("[%s] %s\n", timestamp, OSLtoString(message))
}

func (Log) timestamp(message any) {
	unix := time.Now().UnixMicro()
	fmt.Printf("[%d] %s\n", unix, OSLtoString(message))
}

type LogEntry struct {
	timestamp string
	level     string
	message   string
}

var logHistory []LogEntry
var maxHistorySize = 1000

func (Log) enableHistory() {
	logHistory = []LogEntry{}
}

func (Log) disableHistory() {
	logHistory = []LogEntry{}
}

func (Log) getHistory() []map[string]any {
	result := make([]map[string]any, len(logHistory))
	for i, entry := range logHistory {
		result[i] = map[string]any{
			"timestamp": entry.timestamp,
			"level":     entry.level,
			"message":   entry.message,
		}
	}
	return result
}

func (Log) clearHistory() {
	logHistory = []LogEntry{}
}

func (Log) countByLevel() map[string]int {
	counts := make(map[string]int)

	for _, entry := range logHistory {
		counts[entry.level]++
	}

	return counts
}

func (Log) exportHistory(path any) bool {
	pathStr := OSLtoString(path)
	var history strings.Builder

	for _, entry := range logHistory {
		history.WriteString(fmt.Sprintf("[%s] [%s] %s\n", entry.timestamp, entry.level, entry.message))
	}

	err := os.WriteFile(pathStr, []byte(history.String()), 0644)
	if err != nil {
		return false
	}
	return true
}

func (Log) withTimestamp(level any, message any, args ...any) {
	log.log(level, time.Now().Format("2006-01-02 15:04:05")+" "+OSLtoString(message), args...)
}

func (Log) group(title any) {
	fmt.Printf("\n--- %s ---\n", OSLtoString(title))
}

func (Log) groupEnd() {
	fmt.Println("--- End Group ---\n")
}

func (Log) progressBar(current any, total any, width any, label any) {
	currentVal := OSLcastNumber(current)
	totalVal := OSLcastNumber(total)
	widthVal := OSLcastInt(width)
	labelStr := OSLtoString(label)

	if totalVal == 0 {
		return
	}

	percent := (currentVal / totalVal) * 100
	if percent > 100 {
		percent = 100
	}

	fillWidth := int((percent / 100) * float64(widthVal))
	fill := strings.Repeat("█", fillWidth)
	empty := strings.Repeat("░", widthVal-fillWidth)

	fmt.Printf("\r%s [%s%s] %.1f%% ", labelStr, fill, empty, percent)

	if currentVal >= totalVal {
		fmt.Println()
	}
}

func (Log) spinner(message any, done bool) {
	messageStr := OSLtoString(message)
	symbols := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

	if done {
		fmt.Printf("\r✓ %s\n", messageStr)
		return
	}

	index := int(time.Now().UnixNano()/100000000) % len(symbols)
	fmt.Printf("\r%s %s", symbols[index], messageStr)
}

func (Log) assert(condition any, message any) {
	conditionBool := OSLcastBool(condition)

	if !conditionBool {
		messageStr := OSLtoString(message)
		if messageStr == "" {
			messageStr = "Assertion failed"
		}
		log.error(messageStr)
	}
}

func (Log) trace(message any) {
	log.debug(fmt.Sprintf("[TRACE] %s", OSLtoString(message)))
}

func (Log) fatal(message any) {
	messageStr := OSLtoString(message)
	log.error(fmt.Sprintf("[FATAL] %s", messageStr))
	os.Exit(1)
}

func (Log) countdown(seconds any, message any) {
	sec := int(OSLcastNumber(seconds))
	messageStr := OSLtoString(message)

	for i := sec; i > 0; i-- {
		fmt.Printf("\r%s %d... ", messageStr, i)
		time.Sleep(1 * time.Second)
	}

	fmt.Printf("\r%s Done!      \n", messageStr)
}

var log = Log{}
