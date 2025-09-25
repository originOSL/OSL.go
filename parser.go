package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/dlclark/regexp2"
)

// Token types
const (
	TKN_STR           = "str"
	TKN_NUM           = "num"
	TKN_RAW           = "raw"
	TKN_UNK           = "unk"
	TKN_OBJ           = "obj"
	TKN_ARR           = "arr"
	TKN_FNC           = "fnc"
	TKN_MTD           = "mtd"
	TKN_ASI           = "asi"
	TKN_OPR           = "opr"
	TKN_CMP           = "cmp"
	TKN_SPR           = "spr"
	TKN_LOG           = "log"
	TKN_QST           = "qst"
	TKN_BIT           = "bit"
	TKN_URY           = "ury"
	TKN_MTV           = "mtv"
	TKN_CMD           = "cmd"
	TKN_MOD_INDICATOR = "mod_indicator"
	TKN_INL           = "inl"
	TKN_BLK           = "blk"
	TKN_VAR           = "var"
	TKN_TSR           = "tsr"
	TKN_EVL           = "evl"
	TKN_RMT           = "rmt"
	TKN_MOD           = "mod"
	TKN_BSL           = "bsl"
)

// Token represents a parsed token
type Token struct {
	Type             string      `json:"type"`
	Data             interface{} `json:"data"`
	Source           string      `json:"source,omitempty"`
	Line             int         `json:"line,omitempty"`
	Left             *Token      `json:"left,omitempty"`
	Right            *Token      `json:"right,omitempty"`
	Right2           *Token      `json:"right2,omitempty"`
	Parameters       []*Token    `json:"parameters,omitempty"`
	IsStatic         bool        `json:"isStatic,omitempty"`
	Static           interface{} `json:"static,omitempty"`
	ParseError       string      `json:"parse_error,omitempty"`
	SetType          string      `json:"set_type,omitempty"`
	Returns          string      `json:"returns,omitempty"`
	Cases            interface{} `json:"cases,omitempty"`
	Final            *Token      `json:"final,omitempty"`
	Local            bool        `json:"local,omitempty"`
	StaticAssignment bool        `json:"staticAssignment,omitempty"`
}

// FunctionSignature represents a function's type signature
type FunctionSignature struct {
	Accepts    []string `json:"accepts"`
	Returns    string   `json:"returns"`
	ReturnType string   `json:"returnType"`
}

// OSLUtils is the main parser utility
type OSLUtils struct {
	operators           []string
	comparisons         []string
	logic               []string
	bitwise             []string
	unary               []string
	inlinableFunctions  map[string]any
	functionReturnTypes map[string]FunctionSignature

	// Regex patterns
	regex              *regexp.Regexp
	fullASTRegex       *regexp2.Regexp
	lineTokeniserRegex *regexp2.Regexp
	lineEndingRegex    *regexp.Regexp
	macLineEndingRegex *regexp.Regexp

	// Optimization settings
	optimizationSettings map[string]interface{}
	variableUsage        map[string]int
	definedVariables     map[string]bool

	// Caches and pools
	nodePool          []*Token
	tokenCache        map[string]*Token
	astCache          map[string][]*Token
	constFoldingCache map[string]any
	typeCache         map[string]string
	staticTypes       map[string]bool
	evaluableOps      map[string]bool
	inlinableOps      map[string]bool
	commonStrings     map[string]string
}

// GenerateError creates an error token
func (utils *OSLUtils) GenerateError(ast *Token, error string) []*Token {
	// Simplified error generation
	newToken := &Token{
		Type:   TKN_UNK,
		Data:   fmt.Sprintf("error: %s", error),
		Source: ast.Source,
		Line:   ast.Line,
	}
	return []*Token{newToken}
}

// NewOSLUtils creates a new OSL parser instance
func NewOSLUtils() *OSLUtils {
	utils := &OSLUtils{
		operators:   []string{"+", "++", "-", "*", "/", "//", "%", "??", "^", "b+", "b-", "b/", "b*", "b^"},
		comparisons: []string{"!=", "==", "!==", "===", ">", "<", "!>", "!<", ">=", "<=", "in", "notIn"},
		logic:       []string{"and", "or", "nor", "xor", "xnor", "nand"},
		bitwise:     []string{"|", "&", "<<", ">>", "^^"},
		unary:       []string{"typeof", "new"},

		inlinableFunctions:  make(map[string]interface{}),
		functionReturnTypes: make(map[string]FunctionSignature),

		optimizationSettings: map[string]interface{}{
			"maxLoopUnrollCount":  8,
			"maxLoopUnrollSize":   50,
			"deadCodeElimination": true,
			"constantFolding":     true,
			"loopUnrolling":       true,
		},

		variableUsage:     make(map[string]int),
		definedVariables:  make(map[string]bool),
		nodePool:          make([]*Token, 0),
		tokenCache:        make(map[string]*Token),
		astCache:          make(map[string][]*Token),
		constFoldingCache: make(map[string]any),
		typeCache:         make(map[string]string),

		staticTypes:  map[string]bool{"str": true, "num": true, "unk": true, "cmd": true, "raw": true},
		evaluableOps: map[string]bool{"+": true, "-": true, "*": true, "/": true, "%": true, "^": true, "==": true, "!=": true, ">": true, "<": true, ">=": true, "<=": true, "and": true, "or": true},
		inlinableOps: map[string]bool{"+": true, "-": true, "*": true, "/": true, "%": true, "^": true},

		commonStrings: map[string]string{
			"=": "=", "@=": "@=", "++": "++", "--": "--",
			"def": "def", "if": "if", "else": "else", "for": "for",
			"while": "while", "return": "return", "true": "true", "false": "false",
		},
	}

	// Initialize regex patterns
	utils.regex = regexp.MustCompile(`"[^"]+"|{[^}]+}|\[[^\]]+\]|[^."(]*\((?:(?:"[^"]+")*[^.]+)*|\d[\d.]+\d|[^." ]+`)

	var err error

	utils.fullASTRegex, err = regexp2.Compile(
		`("(?:[^"\\]|\\.)*"|`+
			"`"+`(?:[^`+"`"+`\\]|\\.)*`+"`"+`|'(?:[^'\\]|\\.)*')|\/\*[^*]+|[,{\[]\s*[\r\n]\s*[}\]]?|[\r\n]\s*[}\.\]]|;|(?<=[)"\]}a-zA-Z\d])\[(?=[^\]])|(?<=[)\]])\(|([\r\n]|^)\s*\/\/[^\r\n]+|[\r\n]`,
		0,
	)
	if err != nil {
		panic(fmt.Sprintf("bad fullASTRegex: %v", err))
	}

	utils.lineTokeniserRegex, err = regexp2.Compile(
		`("(?:[^"\\]|\\.)*"|`+
			"`"+`(?:[^`+"`"+`\\]|\\.)*`+"`"+`|'(?:[^'\\]|\\.)*')|(?<=[\]"}\w\)])(?:\+\+|\?\?|->|==|!=|<=|>=|[><?+*^%/\-|&])(?=\S)`,
		0,
	)
	if err != nil {
		panic(fmt.Sprintf("bad lineTokeniserRegex: %v", err))
	}
	utils.lineEndingRegex = regexp.MustCompile(`\r\n`)
	utils.macLineEndingRegex = regexp.MustCompile(`\r`)

	// Initialize function signatures
	utils.functionReturnTypes["random"] = FunctionSignature{
		Accepts: []string{"number", "number"},
		Returns: "number",
	}
	utils.functionReturnTypes["typeof"] = FunctionSignature{
		Accepts: []string{"any"},
		Returns: "string",
	}

	return utils
}

// Helper function to check if string is in slice
func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}

func JsonStringify(obj any) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.Encode(obj)
	return strings.TrimSpace(buf.String())
}

// NormalizeLineEndings normalizes line endings to Unix-style
func (utils *OSLUtils) NormalizeLineEndings(text string) string {
	text = utils.lineEndingRegex.ReplaceAllString(text, "\n")
	text = utils.macLineEndingRegex.ReplaceAllString(text, "\n")
	return text
}

// Tokenise splits code by delimiter respecting quotes and brackets
func Tokenise(code, delimiter string) []string {
	if len(code) == 0 || len(delimiter) == 0 {
		return []string{}
	}

	letter := 0
	brackets := 0
	bDepth := 0
	out := strings.Builder{}
	var split []string
	length := len(code)

	for letter < length {
		if letter >= length {
			break
		}

		depth := code[letter]

		if depth == '"' {
			brackets = 1 - brackets
			out.WriteByte('"')
		} else {
			out.WriteByte(depth)
		}

		if brackets == 0 {
			if depth == '[' || depth == '{' || depth == '(' {
				bDepth++
			}
			if depth == ']' || depth == '}' || depth == ')' {
				bDepth--
			}
			if bDepth < 0 {
				bDepth = 0
			}
		}
		letter++

		if brackets == 0 && letter < length && code[letter] == delimiter[0] && bDepth == 0 && len(delimiter) == 1 {
			split = append(split, out.String())
			out.Reset()
			letter++
		}
	}
	split = append(split, out.String())
	return split
}

// TokeniseEscaped handles escaped characters in tokenization
func TokeniseEscaped(code, delimiter string) []string {
	if len(code) == 0 || len(delimiter) == 0 {
		return []string{}
	}

	letter := 0
	brackets := 0
	bDepth := 0
	escaped := false
	out := strings.Builder{}
	var split []string
	length := len(code)

	for letter < length {
		if letter >= length {
			break
		}

		depth := code[letter]

		if brackets == 0 && !escaped {
			if depth == '[' || depth == '{' || depth == '(' {
				bDepth++
			}
			if depth == ']' || depth == '}' || depth == ')' {
				bDepth--
			}
			if bDepth < 0 {
				bDepth = 0
			}
		}

		if depth == '"' && !escaped {
			brackets = 1 - brackets
			out.WriteByte('"')
		} else if depth == '\\' && !escaped {
			escaped = !escaped
			out.WriteByte('\\')
		} else {
			out.WriteByte(depth)
			escaped = false
		}
		letter++

		if brackets == 0 && letter < length && code[letter] == delimiter[0] && bDepth == 0 && len(delimiter) == 1 {
			split = append(split, out.String())
			out.Reset()
			letter++
		}
	}
	split = append(split, out.String())
	return split
}

// ParseEscaped handles escape sequences in strings
func ParseEscaped(str string) string {
	result := strings.Builder{}
	for i := 0; i < len(str); i++ {
		if str[i] == '\\' && i+1 < len(str) {
			i++
			esc := str[i]
			switch esc {
			case 'n':
				result.WriteByte('\n')
			case 't':
				result.WriteByte('\t')
			case 'r':
				result.WriteByte('\r')
			case '"':
				result.WriteByte('"')
			case '\'':
				result.WriteByte('\'')
			case '\\':
				result.WriteByte('\\')
			default:
				result.WriteByte(esc)
			}
		} else {
			result.WriteByte(str[i])
		}
	}
	return result.String()
}

// Destr removes string delimiters and handles escaping
func Destr(t string, delim string) string {
	if delim == "" {
		delim = `"`
	}

	n := t
	r := delim

	if strings.HasPrefix(n, r) && strings.HasSuffix(n, r) && len(n) >= 2 {
		content := n[1 : len(n)-1]
		return ParseEscaped(content)
	}
	return t
}

// AutoTokenise automatically chooses tokenization method
func AutoTokenise(code, delimiter string) []string {
	if delimiter == "" {
		delimiter = " "
	}

	if strings.Contains(code, "\\") {
		return TokeniseEscaped(code, delimiter)
	} else if strings.Contains(code, `"`) || strings.Contains(code, "[") || strings.Contains(code, "{") || strings.Contains(code, "(") {
		return Tokenise(code, delimiter)
	} else {
		return strings.Split(code, delimiter)
	}
}

// ParseTemplate parses template literals with ${} expressions
func ParseTemplate(str string) []any {
	depth := 0
	cur := strings.Builder{}
	var arr []any

	for i := 0; i < len(str); i++ {
		if i+1 < len(str) && str[i] == '$' && str[i+1] == '{' {
			if depth == 0 {
				if cur.Len() > 0 {
					arr = append(arr, cur.String())
				}
				cur.Reset()
				cur.WriteString("${")
			} else {
				cur.WriteString("${")
			}
			depth++
			i++ // Skip the next character
			continue
		}

		if str[i] == '}' && depth > 0 {
			depth--
			if depth == 0 {
				cur.WriteByte('}')
				arr = append(arr, cur.String())
				cur.Reset()
			} else {
				cur.WriteByte('}')
			}
			continue
		}

		cur.WriteByte(str[i])
	}

	if cur.Len() > 0 {
		arr = append(arr, cur.String())
	}
	return arr
}

// RandomString generates a random string of given length
func RandomString(length int) string {
	characters := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	result := strings.Builder{}

	for i := 0; i < length; i++ {
		result.WriteByte(characters[i%len(characters)]) // Simplified for demo
	}
	return result.String()
}

// IsStaticToken checks if a token represents a static value
func (utils *OSLUtils) IsStaticToken(token *Token) bool {
	if token == nil {
		return false
	}
	return utils.staticTypes[utils.getTokenTypeString(token.Type)]
}

// getTokenTypeString converts token type int to string
func (utils *OSLUtils) getTokenTypeString(tokenType string) string {
	switch tokenType {
	case TKN_STR:
		return "str"
	case TKN_NUM:
		return "num"
	case TKN_UNK:
		return "unk"
	case TKN_CMD:
		return "cmd"
	case TKN_RAW:
		return "raw"
	default:
		return "unknown"
	}
}

// TokeniseLineOSL tokenizes a single line of OSL code
func (utils *OSLUtils) TokeniseLineOSL(code string) []string {
	// Apply regex replacement for operators
	code, err := utils.lineTokeniserRegex.ReplaceFunc(code, func(m regexp2.Match) string {
		v := m.String()
		if strings.HasPrefix(v, `"`) || strings.HasPrefix(v, `'`) || strings.HasPrefix(v, "`") {
			return v
		}
		return " " + v + " "
	}, -1, -1)
	if err != nil {
		fmt.Println(err)
	}

	letter := 0
	quotes := 0
	squotes := 0
	backticks := 0
	mComm := 0
	bDepth := 0
	escaped := false
	out := strings.Builder{}
	var split []string
	length := len(code)

	for letter < length {
		if letter >= length {
			break
		}

		depth := code[letter]

		if quotes == 0 && squotes == 0 && backticks == 0 && !escaped {
			if depth == '[' || depth == '{' || depth == '(' {
				bDepth++
			}
			if depth == ']' || depth == '}' || depth == ')' {
				bDepth--
			}
			if bDepth < 0 {
				bDepth = 0
			}
		}

		if depth == '"' && !escaped && squotes == 0 && backticks == 0 {
			quotes = 1 - quotes
		} else if depth == '\'' && !escaped && quotes == 0 && backticks == 0 {
			squotes = 1 - squotes
		} else if depth == '`' && !escaped && quotes == 0 && squotes == 0 {
			backticks = 1 - backticks
		} else if depth == '/' && letter+1 < length && code[letter+1] == '*' && quotes == 0 && squotes == 0 && backticks == 0 {
			mComm = 1
		} else if depth == '*' && letter+1 < length && code[letter+1] == '/' && quotes == 0 && squotes == 0 && backticks == 0 && mComm == 1 {
			mComm = 0
		} else if depth == '\\' && !escaped {
			escaped = !escaped
		} else {
			escaped = false
		}

		if mComm == 0 {
			out.WriteByte(depth)
		}
		letter++

		if quotes == 0 && squotes == 0 && backticks == 0 && bDepth == 0 && mComm == 0 && letter < length && (code[letter] == ' ' || code[letter] == ')') {
			if code[letter] != ' ' && code[letter] != ')' {
				split = append(split, string(depth))
			} else {
				split = append(split, out.String())
			}
			out.Reset()
			letter++
		}
	}
	split = append(split, out.String())
	return split
}

// TokeniseLines tokenizes code into lines
func (utils *OSLUtils) TokeniseLines(code string) []string {
	// Normalize line endings first
	code = utils.NormalizeLineEndings(code)

	letter := 0
	quotes := 0
	backticks := 0
	bDepth := 0
	escaped := false
	out := strings.Builder{}
	var split []string
	length := len(code)

	for letter < length {
		if letter >= length {
			break
		}

		depth := code[letter]

		if quotes == 0 && backticks == 0 && !escaped {
			if depth == '[' || depth == '{' || depth == '(' {
				bDepth++
			}
			if depth == ']' || depth == '}' || depth == ')' {
				bDepth--
			}
			if bDepth < 0 {
				bDepth = 0
			}
		}

		if depth == '"' && !escaped && backticks == 0 {
			quotes = 1 - quotes
			out.WriteByte('"')
		} else if depth == '`' && !escaped && quotes == 0 {
			backticks = 1 - backticks
			out.WriteByte('`')
		} else if depth == '\\' && !escaped {
			escaped = !escaped
			out.WriteByte('\\')
		} else {
			out.WriteByte(depth)
			escaped = false
		}
		letter++

		if quotes == 0 && backticks == 0 && letter < length && (code[letter] == '\n' || code[letter] == ';') && bDepth == 0 {
			split = append(split, out.String())
			out.Reset()
			letter++
		}
	}
	split = append(split, out.String())
	return split
}

// FindMatchingParentheses finds the matching closing parenthesis
func (utils *OSLUtils) FindMatchingParentheses(code string, startIndex int) int {
	depth := 1
	endIndex := startIndex + 1

	for endIndex < len(code) {
		switch code[endIndex] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return endIndex
			}
		}
		endIndex++
	}
	return -1
}

// EvalToken evaluates a token and sets its source
func (utils *OSLUtils) EvalToken(cur string, param bool) *Token {
	token := utils.StringToToken(cur, param)
	if token.Type == TKN_BLK {
		token.Source = "[ast BLK]"
	} else {
		token.Source = cur
	}
	return token
}

func (utils *OSLUtils) StringToToken(cur string, param bool) *Token {
	if len(cur) == 0 {
		return &Token{Type: TKN_UNK, Data: cur}
	}

	start := cur[0]

	if cur == "/@line" {
		return &Token{Type: TKN_UNK, Data: "/@line"}
	}

	numStr := strings.ReplaceAll(cur, "_", "")
	if num, err := strconv.ParseFloat(numStr, 64); err == nil {
		return &Token{Type: TKN_NUM, Data: num}
	}

	if cur == "true" || cur == "false" {
		return &Token{Type: TKN_RAW, Data: cur == "true"}
	}

	if contains(utils.operators, cur) {
		return &Token{Type: TKN_OPR, Data: cur}
	}

	if cur == "++" {
		return &Token{Type: TKN_OPR, Data: "++"}
	}
	if cur == "--" {
		return &Token{Type: TKN_UNK, Data: "--"}
	}

	if contains(utils.comparisons, cur) {
		return &Token{Type: TKN_CMP, Data: cur}
	}

	if strings.HasSuffix(cur, "=") {
		return &Token{Type: TKN_ASI, Data: cur}
	}

	if len(cur) >= 2 {
		if start == '"' && cur[len(cur)-1] == '"' {
			return &Token{Type: TKN_STR, Data: Destr(cur, "")}
		}
		if start == '\'' && cur[len(cur)-1] == '\'' {
			return &Token{Type: TKN_STR, Data: Destr(cur, "'")}
		}
		if start == '`' && cur[len(cur)-1] == '`' {
			templateData := ParseTemplate(Destr(cur, "`"))
			var filteredData []any
			for _, v := range templateData {
				if str, ok := v.(string); !ok || str != "" {
					if str, ok := v.(string); ok && strings.HasPrefix(str, "${") && strings.HasSuffix(str, "}") {
						ast := utils.GenerateAST(str, 0, false)
						if len(ast) == 0 {
							continue
						}
						filteredData = append(filteredData, ast[0])
					} else {
						filteredData = append(filteredData, &Token{Type: TKN_STR, Data: v})
					}
				}
			}
			return &Token{Type: TKN_TSR, Data: filteredData}
		}
	}

	if cur == "?" {
		return &Token{Type: TKN_QST, Data: cur}
	}

	if contains(utils.logic, cur) {
		return &Token{Type: TKN_LOG, Data: cur}
	}

	if contains(utils.bitwise, cur) {
		return &Token{Type: TKN_BIT, Data: cur}
	}

	if strings.HasPrefix(cur, "...") {
		return &Token{Type: TKN_SPR, Data: utils.StringToToken(cur[3:], false)}
	}

	if len(cur) > 1 && (start == '!' || start == '-' || start == '+') {
		return &Token{
			Type:  TKN_URY,
			Data:  string(start),
			Right: utils.StringToToken(cur[1:], false),
		}
	}

	if strings.Contains(cur, ".") {
		method := AutoTokenise(cur, ".")
		if len(method) >= 2 {
			var tokens []*Token
			for i, input := range method {
				tokens = append(tokens, utils.StringToToken(input, i > 0))
			}
			return &Token{Type: TKN_MTD, Data: tokens}
		}
	}

	if len(cur) >= 2 {
		if (start == '{' && cur[len(cur)-1] == '}') || (start == '[' && cur[len(cur)-1] == ']') {
			if start == '[' {
				if cur == "[]" {
					if param {
						return &Token{Type: TKN_MTV, Data: "item", Parameters: []*Token{}}
					} else {
						return &Token{Type: TKN_ARR, Data: []*Token{}}
					}
				}

				tokens := AutoTokenise(cur[1:len(cur)-1], ",")
				var filteredTokens []string
				for _, token := range tokens {
					if strings.TrimSpace(token) != "" {
						filteredTokens = append(filteredTokens, token)
					}
				}

				var parsedTokens []*Token
				for _, token := range filteredTokens {
					cur := strings.TrimSpace(token)
					if strings.HasPrefix(cur, "/@line ") {
						lines := strings.SplitN(cur, "\n", 2)
						cur = strings.TrimSpace(strings.ReplaceAll(cur, lines[1]+"\n", ""))
					}

					ast := utils.GenerateAST(cur, 0, false)
					parsedTokens = append(parsedTokens, ast[0])
				}

				if param {
					obj := &Token{Type: TKN_MTV, Data: "item", Parameters: parsedTokens}
					obj.IsStatic = true
					for _, token := range parsedTokens {
						if !utils.IsStaticToken(token) {
							obj.IsStatic = false
							break
						}
					}
					if obj.IsStatic {
						if len(parsedTokens) == 1 && parsedTokens[0].Type == TKN_STR {
							return &Token{Type: TKN_MTV, Data: parsedTokens[0].Data}
						}
						var static []any
						for _, token := range parsedTokens {
							static = append(static, token.Data)
						}
						obj.Static = static
					}
					return obj
				}

				arr := &Token{Type: TKN_ARR, Data: parsedTokens}
				arr.IsStatic = true
				for _, token := range parsedTokens {
					if !utils.IsStaticToken(token) {
						arr.IsStatic = false
						break
					}
				}
				if arr.IsStatic {
					var static []interface{}
					for _, token := range parsedTokens {
						static = append(static, token.Data)
					}
					arr.Static = static
				}
				return arr
			} else if cur[0] == '{' {
				if cur == "{}" {
					return &Token{Type: TKN_OBJ, Data: nil}
				}

				var output = [][]*Token{}
				tokens := AutoTokenise(cur[1:len(cur)-1], ",")
				for _, token := range tokens {
					if strings.TrimSpace(token) == "" {
						continue
					}
					keyValue := AutoTokenise(token, ":")
					key := strings.TrimSpace(keyValue[0])
					value := keyValue[1]
					if strings.HasPrefix(key, "/@line ") {
						first, _, _ := strings.Cut(key, "\n")
						key = strings.Replace(first, "\n", "", 1)
					}
					if value == "" {
						ast := utils.GenerateAST(key, 0, false)
						if len(ast) == 0 {
							continue
						}
						nkey := ast[0]
						if nkey.Type == TKN_VAR {
							ast := utils.GenerateAST(JsonStringify(key), 0, false)
							if len(ast) == 0 {
								continue
							}
							nkey = ast[0]
						}
						output = append(output, []*Token{nkey, nkey})
						continue
					}
					if strings.HasPrefix(key, "{") && strings.HasSuffix(key, "}") {
						key = key[1 : len(key)-1]
					} else {
						temp_key := utils.EvalToken(key, true)
						if temp_key.Type == TKN_STR {
							key = JsonStringify(key)
						}
					}
					key_ast := utils.GenerateAST(key, 0, false)
					if len(key_ast) == 0 {
						continue
					}
					if value == "" {
						output = append(output, []*Token{key_ast[0], nil})
					} else {
						ast := utils.GenerateAST(value, 0, false)
						if len(ast) == 0 {
							continue
						}
						output = append(output, []*Token{key_ast[0], ast[0]})
					}
				}
				return &Token{Type: TKN_OBJ, Data: output}
			}
		}
	}

	matched := regexp.MustCompile(`^(!+)?[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(cur)

	if matched {
		return &Token{Type: TKN_VAR, Data: cur}
	}

	if cur == "->" {
		return &Token{Type: TKN_INL, Data: "->"}
	}

	if strings.HasPrefix(cur, "(\n") && strings.HasSuffix(cur, ")") {
		return &Token{Type: TKN_BLK, Data: utils.GenerateFullAST(strings.TrimSpace(cur[1:len(cur)-1]), false)}
	}

	if strings.HasPrefix(cur, "(") && strings.HasSuffix(cur, ")") {
		end := utils.FindMatchingParentheses(cur, 0)
		if end == -1 {
			return &Token{Type: TKN_UNK, Data: cur, ParseError: "Unmatched parentheses"}
		}
		body := strings.TrimSpace(cur[1:end])
		ast := utils.GenerateAST(body, 0, false)
		if len(ast) == 0 {
			panic("Invalid OSL parentheses usage")
		}
		return ast[0]
	}

	if strings.HasSuffix(cur, ")") && len(cur) > 1 {
		parenIndex := strings.Index(cur, "(")
		if parenIndex > 0 {
			funcName := cur[:parenIndex]
			tokenType := TKN_FNC
			if param {
				tokenType = TKN_MTV
			}

			out := &Token{Type: tokenType, Data: funcName, Parameters: []*Token{}}

			if strings.HasSuffix(cur, "()") {
				return out
			}

			params := cur[parenIndex+1 : len(cur)-1]
			method := AutoTokenise(params, ",")
			var parsedParams []*Token

			for _, v := range method {
				trimmed := strings.TrimSpace(v)
				typePrefix := ""
				ast := utils.GenerateAST(trimmed, 0, false)
				if len(ast) == 0 {
					continue
				}
				if len(ast) > 1 {
					parts := AutoTokenise(trimmed, " ")
					typePrefix = parts[0]
					ast = ast[1:]
				}
				if typePrefix != "" {
					ast[0].SetType = typePrefix
				}
				parsedParams = append(parsedParams, ast[0])
			}
			out.Parameters = parsedParams
			return out
		}
	}

	if cur == ":" {
		return &Token{Type: TKN_MOD_INDICATOR, Data: ":"}
	}

	return &Token{Type: TKN_UNK, Data: cur}
}

func (utils *OSLUtils) evalASTNode(node *Token) *Token {
	if node == nil {
		return node
	}

	if node.Type == TKN_INL {
		var params string
		if node.Left != nil {
			if node.Left.Parameters != nil {
				var paramStrs []string
				for _, p := range node.Left.Parameters {
					if data, ok := p.Data.(string); ok {
						if p.SetType != "" {
							data = data + " " + p.SetType
						}
						paramStrs = append(paramStrs, data)
					}
				}
				params = strings.Join(paramStrs, ",")
			} else if node.Left.Type == TKN_VAR {
				if data, ok := node.Left.Data.(string); ok {
					params = data
				}
			}
		}

		right := node.Right
		if right != nil {
			if rightData, ok := right.Data.(string); ok {
				if !strings.HasPrefix(strings.TrimSpace(rightData), "(\n") && node.Left != nil {
					right.Data = fmt.Sprintf("(\nreturn %s\n)", right.Source)
				}
			}
		} else {
			panic("No body for inline function")
		}

		return &Token{
			Type: TKN_FNC,
			Data: "function",
			Parameters: []*Token{
				{Type: TKN_STR, Data: params, Source: params},
				right,
				{Type: TKN_UNK, Data: !strings.HasPrefix(node.Source, "def(")},
			},
		}
	}

	return node
}

func (utils *OSLUtils) GenerateAST(code string, start int, main bool) []*Token {
	code = utils.NormalizeLineEndings(code)
	startLine := strings.SplitN(code, "\n", 2)[0]
	handlingMods := false

	var ast []*Token
	tokens := utils.TokeniseLineOSL(code)
	for i := range tokens {
		cur := strings.TrimSpace(tokens[i])

		if cur == "->" {
			ast = append(ast, &Token{Type: TKN_INL, Data: "->"})
			continue
		}

		if handlingMods {
			token := &Token{Type: TKN_MOD, Data: cur, Source: cur}
			pivot := strings.Index(cur, "#")
			if pivot >= 0 {
				token.Data = []any{cur[:pivot], utils.EvalToken(cur[pivot+1:], false)}
			}
			ast = append(ast, token)
			continue
		}

		curT := utils.EvalToken(cur, false)
		if curT.Type == TKN_MOD_INDICATOR {
			handlingMods = true
			continue
		}
		ast = append(ast, curT)
	}

	types := []string{TKN_OPR, TKN_CMP, TKN_QST, TKN_BIT, TKN_LOG, TKN_INL}

	for _, nodeType := range types {
		startIdx := start
		if startIdx < 0 {
			if nodeType == TKN_ASI || nodeType == TKN_INL {
				startIdx = 1
			} else {
				startIdx = 2
			}
		}

		for i := startIdx; i < len(ast); i++ {
			if i >= len(ast) {
				break
			}
			cur := ast[i]

			if cur.Type == nodeType {
				var prev, next, next2 *Token
				if i > 0 {
					prev = ast[i-1]
				}
				if i+1 < len(ast) {
					next = ast[i+1]
				}
				if i+2 < len(ast) {
					next2 = ast[i+2]
				}

				if nodeType == TKN_QST {
					cur.Left = prev
					cur.Right = next
					cur.Right2 = next2
					if i > 0 {
						ast = append(ast[:i-1], ast[i:]...)
						i--
					}
					if i+1 < len(ast) {
						ast = append(ast[:i+1], ast[i+2:]...)
					}
					if i+1 < len(ast) {
						ast = append(ast[:i+1], ast[i+2:]...)
					}
					continue
				}

				if cur.Left == nil && prev != nil && next != nil {
					cur.Left = prev
					cur.Right = next
					source := ""
					if prev.Source != "" {
						source += prev.Source
					}
					if cur.Source != "" {
						source += " " + cur.Source
					}
					if next.Source != "" {
						source += " " + next.Source
					}
					cur.Source = source

					if i > 0 {
						ast = append(ast[:i-1], ast[i:]...)
						i--
					}
					if i+1 < len(ast) {
						ast = append(ast[:i+1], ast[i+2:]...)
					}
				}
			}
		}
	}

	for i := 0; i < len(ast); i++ {
		if ast[i].ParseError != "" {
			return utils.GenerateError(ast[i], ast[i].ParseError)
		}
	}

	for i := 0; i < len(ast); i++ {
		ast[i] = utils.evalASTNode(ast[i])
	}

	if len(ast) > 0 {
		first := ast[0]
		var second *Token
		if len(ast) > 1 {
			second = ast[1]
		}

		local := first.Data == "local" && first.Type == TKN_VAR

		if first.Type == TKN_VAR || first.Type == TKN_CMD {
			if firstData, ok := first.Data.(string); ok && firstData == "def" {
				if second != nil && second.Type == TKN_FNC {
					if local {
						ast = ast[1:]
						if len(ast) == 0 {
							return []*Token{}
						}
						if len(ast) > 1 {
							second = ast[1]
						}
					}

					funcName := second.Data.(string)
					var params []string
					for _, p := range second.Parameters {
						if data, ok := p.Data.(string); ok {
							if p.SetType != "" {
								data = data + " " + p.SetType
							}
							params = append(params, data)
						}
					}
					paramSpec := strings.Join(params, ",")

					var funcBody *Token
					returnType := ""

					if len(ast) > 2 && ast[2].Type == TKN_VAR {
						returnType = ast[2].Data.(string)
						if len(ast) > 3 {
							funcBody = ast[3]
						}
					} else if len(ast) > 2 {
						funcBody = ast[2]
					}

					funcNode := &Token{
						Type:    TKN_FNC,
						Data:    "function",
						Returns: returnType,
						Parameters: []*Token{
							{Type: TKN_STR, Data: paramSpec, Source: paramSpec},
							funcBody,
							{Type: TKN_UNK, Data: false},
						},
					}

					ast = []*Token{
						{
							Type:   TKN_ASI,
							Data:   "=",
							Source: startLine,
							Left:   &Token{Type: TKN_VAR, Data: funcName, Source: funcName},
							Right:  funcNode,
						},
					}

					if strings.TrimSpace(paramSpec) != "" {
						paramParts := strings.Split(paramSpec, ",")
						var paramTypes []string
						for _, paramPart := range paramParts {
							parts := strings.Fields(strings.TrimSpace(paramPart))
							if len(parts) >= 2 {
								paramTypes = append(paramTypes, parts[0])
							} else {
								paramTypes = append(paramTypes, "any")
							}
						}

						utils.functionReturnTypes[funcName] = FunctionSignature{
							ReturnType: returnType,
							Accepts:    paramTypes,
						}
					}
				}
			}
		}
	}

	if len(ast) == 0 {
		return []*Token{}
	}

	if len(ast) > 0 && ast[0].Type == TKN_MTD {
		if dataSlice, ok := ast[0].Data.([]*Token); ok {
			if len(dataSlice) > 0 {
				lastNode := dataSlice[len(dataSlice)-1]
				if lastNode.Type == TKN_MTV && len(ast) == 1 && main {
					firstMtvIndex := -1
					for i, node := range dataSlice {
						if node.Type == TKN_MTV {
							firstMtvIndex = i
							break
						}
					}

					var leftData []*Token
					if firstMtvIndex > 0 {
						leftData = dataSlice[:firstMtvIndex]
					} else {
						leftData = []*Token{dataSlice[0]}
					}

					var first *Token
					if len(leftData) == 1 {
						first = leftData[0]
					} else {
						first = &Token{Type: TKN_MTD, Data: leftData}
					}

					newAst := []*Token{
						first,
						{Type: TKN_ASI, Data: "=??", Source: startLine},
					}
					ast = append(newAst, ast...)
				}
			}
		}
	}

	startIdx := start
	if startIdx < 0 {
		startIdx = 1
	}

	for i := startIdx; i < len(ast); i++ {
		if i >= len(ast) {
			break
		}
		cur := ast[i]

		if cur.Type == TKN_ASI {
			if len(ast) > 0 && ast[0].Type == TKN_VAR {
				if data, ok := ast[0].Data.(string); ok && data == "local" {
					if i > 0 {
						prev := ast[i-1]
						newPrev := &Token{Type: TKN_STR, Data: "this." + prev.Source}
						ast[i-1] = newPrev
					}
					ast = append(ast[:0], ast[1:]...)
					i--
				}
			}

			if i > 1 && len(ast) > i-2 {
				if ast[i-2] != nil {
					if data, ok := ast[i-2].Data.(string); ok {
						cur.SetType = strings.ToLower(data)
						ast = append(ast[:i-2], ast[i-1:]...)
						i--
					}
				}
			}

			if cur.Left == nil {
				var prev, next *Token
				if i > 0 {
					prev = ast[i-1]
				}
				if i+1 < len(ast) {
					next = ast[i+1]
				}

				cur.Left = prev
				cur.Right = next

				if i > 0 {
					ast = append(ast[:i-1], ast[i:]...)
					i--
				}
				if i+1 < len(ast) {
					ast = append(ast[:i+1], ast[i+2:]...)
				}
			}

			if cur.Left != nil && cur.Left.Type == TKN_MTD {
				if pathData, ok := cur.Left.Data.([]*Token); ok && len(pathData) > 0 {
					path := make([]*Token, len(pathData)-1)
					copy(path, pathData[:len(pathData)-1])
					final := pathData[len(pathData)-1]

					cur.Left = &Token{
						Type:  TKN_RMT,
						Data:  path,
						Final: final,
					}
				}
			}

			cur.Source = startLine
		}
	}

	if len(ast) == 0 {
		return []*Token{}
	}

	if len(ast) == 2 {
		t1 := ast[1]
		if (t1.Data == "--" && t1.Type == TKN_UNK && t1.Right == nil) ||
			(t1.Data == "++" && t1.Type == TKN_OPR && t1.Right == nil) {
			ast[0] = &Token{
				Type:   TKN_ASI,
				Data:   t1.Data,
				Left:   ast[0],
				Source: code,
			}
			ast = ast[:1]
		}
	}

	if main && len(ast) > 0 {
		if ast[0].Type == TKN_VAR {
			ast[0].Type = TKN_CMD
		}
		ast[0].Source = strings.SplitN(code, "\n", 2)[0]
	}

	if len(ast) > 0 && ast[0].Type == TKN_CMD {
		if data, ok := ast[0].Data.(string); ok && data == "switch" {
			if len(ast) > 2 && ast[2].Type == TKN_BLK {
				cases := map[string]any{
					"type": "array",
					"all":  []any{},
				}
				ast[0].Cases = cases
			}
		}
	}

	if len(ast) > 0 && ast[0].Type == TKN_ASI {
		if ast[0].Right != nil && utils.IsStaticToken(ast[0].Right) {
			ast[0].Right.StaticAssignment = true
		}
	}

	var filtered []*Token
	for _, token := range ast {
		if token.Type != TKN_UNK || !strings.HasPrefix(fmt.Sprint(token.Data), "/*") {
			filtered = append(filtered, token)
		}
	}

	return filtered
}

func (utils *OSLUtils) GenerateFullAST(code string, main bool) [][]*Token {
	if main {
		utils.inlinableFunctions = make(map[string]any)
	}

	line := 0
	code = utils.NormalizeLineEndings(strings.TrimSpace(code))

	// Add line markers for main context
	if main {
		line++
		code = fmt.Sprintf("/@line %d\n", line) + code
	}

	// Apply regex transformations
	code, err := utils.fullASTRegex.ReplaceFunc(code, func(m regexp2.Match) string {
		match := m.String()
		if strings.HasPrefix(strings.TrimSpace(match), "//") {
			return match
		}
		if match == "\n" {
			if main {
				line++
				return fmt.Sprintf("\n/@line %d\n", line)
			}
			return "\n"
		}
		if match == ";" {
			return "\n"
		}
		if match == "(" {
			return ".call("
		}
		if match == "[" {
			return ".["
		}
		if strings.Contains(",{}[]", string(match[0])) {
			line++
			return match
		}
		if strings.HasPrefix(match, "\n") {
			line++
			return strings.ReplaceAll(match, "\n ", ".")
		}
		return match
	}, -1, -1)
	if err != nil {
		fmt.Println(err)
	}

	// Remove comment lines
	comment_lines := strings.Split(code, "\n")
	var comment_filteredLines []string
	for _, line := range comment_lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "//") {
			comment_filteredLines = append(comment_filteredLines, line)
		}
	}
	code = strings.Join(comment_filteredLines, "\n")

	// Handle def statements
	codeLines := AutoTokenise(code, "\n")
	for i, line := range codeLines {
		line = strings.TrimSpace(line)
		if line == "endef" {
			codeLines[i] = ")"
		} else if strings.HasPrefix(line, "def ") && !strings.HasSuffix(line, "(") && !strings.HasSuffix(line, ")") {
			codeLines[i] = line + " ("
		} else {
			codeLines[i] = line
		}
	}
	code = strings.Join(codeLines, "\n")

	// Generate AST for each line
	lineTokens := utils.TokeniseLines(code)
	var lines [][]*Token

	for _, line := range lineTokens {
		ast := utils.GenerateAST(strings.TrimSpace(line), -1, true)
		if len(ast) > 0 {
			lines = append(lines, ast)
		}
	}

	// Handle line markers
	for i := 0; i < len(lines); i++ {
		if len(lines[i]) > 1 && lines[i][0].Type == TKN_UNK {
			if data, ok := lines[i][0].Data.(string); ok && data == "/@line" {
				if len(lines[i]) > 1 {
					if lineNum, ok := lines[i][1].Data.(float64); ok {
						// Set line number on next statement
						if i+1 < len(lines) && len(lines[i+1]) > 0 {
							lines[i+1][0].Line = int(lineNum)
						}
					}
				}
				// Remove line marker
				lines = append(lines[:i], lines[i+1:]...)
				i--
			}
		}
	}

	// Filter out line markers that weren't processed
	var filteredLines [][]*Token
	for _, line := range lines {
		if len(line) > 0 {
			if line[0].Type != TKN_UNK || line[0].Data != "/@line" {
				filteredLines = append(filteredLines, line)
			}
		}
	}

	// Handle complex control structures
	for i := 0; i < len(filteredLines); i++ {
		if len(filteredLines[i]) == 0 {
			continue
		}

		cur := filteredLines[i]
		cmdType := ""
		if cur[0].Type == TKN_CMD {
			if data, ok := cur[0].Data.(string); ok {
				cmdType = data
			}
		}

		// Handle local class definitions
		if cmdType == "local" && len(cur) > 1 && cur[1].Type == TKN_CMD {
			if data, ok := cur[1].Data.(string); ok && data == "class" {
				cur[1].Source = cur[0].Source
				cur[1].Line = cur[0].Line
				cur = cur[1:]
				cur[0].Type = TKN_CMD
				cur[0].Local = true
				filteredLines[i] = cur
			}
		}

		if cmdType == "for" || cmdType == "each" || cmdType == "class" || cmdType == "while" || cmdType == "until" {
			if cmdType == "each" {
				lastIdx := len(cur) - 1
				if lastIdx >= 0 && cur[lastIdx].Type != TKN_BLK {
					filteredLines[i] = utils.GenerateError(cur[0], "'each' loop missing body block")
					continue
				}

				if len(cur) >= 4 && (cur[4] != nil && cur[4].Type == TKN_BLK || cur[3] != nil && cur[3].Type == TKN_BLK) {
					cur[0].Data = "loop"
				}
			}

			if cmdType == "while" || cmdType == "until" {
				if len(cur) > 1 {
					cur[1] = &Token{
						Type:   TKN_EVL,
						Data:   cur[1],
						Source: cur[1].Source,
					}
				}
			} else if cmdType != "each" {
				if len(cur) > 1 {
					cur[1].Type = TKN_STR
				}
			}
		}

		if cmdType == "def" {
			if len(cur) < 3 {
				filteredLines[i] = utils.GenerateError(cur[0], "Incomplete function definition")
				continue
			}
			lastIdx := len(cur) - 1
			if lastIdx >= 0 && cur[lastIdx].Type != TKN_BLK {
				filteredLines[i] = utils.GenerateError(cur[0], "Function body missing. Add a block: ( ... )")
				continue
			}
		}

		if contains([]string{"loop", "if", "while", "until", "for"}, cmdType) {
			if len(cur) == 2 || (cmdType == "for" && len(cur) == 3) {
				// Look for the next line as the block
				if i+1 < len(filteredLines) {
					blk := filteredLines[i+1]
					cur = append(cur, &Token{
						Type:   TKN_BLK,
						Data:   [][]*Token{blk},
						Source: "[ast BLK]",
					})
					filteredLines[i] = cur
					filteredLines = append(filteredLines[:i+1], filteredLines[i+2:]...)
				}
			}
		}

		// Validate operators and expressions
		for j := 0; j < len(cur); j++ {
			t := cur[j]
			if t == nil {
				continue
			}

			if t.Type == TKN_OPR || t.Type == TKN_CMP || t.Type == TKN_BIT || t.Type == TKN_LOG {
				if t.Left == nil || t.Right == nil {
					if j <= 1 {
						filteredLines[i] = utils.GenerateError(cur[0], fmt.Sprintf("Malformed line. Cannot use '%v' here", t.Data))
						break
					}
					missing := "operands"
					if t.Left != nil {
						missing = "right operand"
					} else if t.Right != nil {
						missing = "left operand"
					}

					opType := "operator"
					if t.Type != TKN_OPR {
						opType = utils.getTokenTypeString(t.Type)
					}

					filteredLines[i] = utils.GenerateError(t, fmt.Sprintf("Malformed %s '%v'. Missing %s.", opType, t.Data, missing))
					break
				}
			}

			if t.Type == TKN_QST {
				if t.Left == nil || t.Right == nil || t.Right2 == nil {
					filteredLines[i] = utils.GenerateError(t, "Incomplete ternary '?'. Expected pattern: condition ? valueIfTrue valueIfFalse")
					break
				}
			}
		}
	}

	return filteredLines
}
