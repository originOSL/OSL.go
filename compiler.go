package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
)

//go:embed packages/*.go
var packagesFS embed.FS

type VariableContext struct {
	DeclaredVars        map[string]bool
	VariableTypes       map[string]string
	functionReturnTypes map[string]string
	Indent              int
	Globals             map[string]any
	Locals              map[string]any
	Prepend             map[string]string
	Imports             map[string]bool
	selfTypes           map[string]string
	ImportOrder         []string
	ImportAliases       map[string]string
	selfUsed            bool
	CustomCommands      map[string]bool
	ScopeLevel          int
	GlobalDeclaredVars  map[string]bool
	GlobalVariableTypes map[string]string
	builtinTypeMethods  map[string]map[string]MethodDefinition
	IsInit              bool
	GlobalVars          strings.Builder
	HoistedVars         []string
}

type MethodDefinition struct {
	TypeName   string
	MethodName string
	ReturnType string
	Token      *Token
}

func include(path string) string {
	data, err := packagesFS.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(data) + "\n"
}

var oslTypes = map[string]string{
	"string":  "string",
	"int":     "int",
	"number":  "float64",
	"boolean": "bool",
	"object":  "map[string]any",
	"array":   "[]any",
}

func mapOSLTypeToGo(oslType string) string {
	val, ok := oslTypes[oslType]
	if ok {
		return val
	}
	if before, ok := strings.CutSuffix(oslType, "[]"); ok {
		return "[]" + mapOSLTypeToGo(before)
	}
	return oslType
}

func processImports(ctx *VariableContext) (compiled string, goImports []string) {
	var orderedImports []string
	processed := make(map[string]bool)

	for _, importPath := range ctx.ImportOrder {
		if ctx.Imports[importPath] && !processed[importPath] {
			orderedImports = append(orderedImports, importPath)
			processed[importPath] = true
		}
	}

	var remaining []string
	for importPath, enabled := range ctx.Imports {
		if enabled && !processed[importPath] {
			remaining = append(remaining, importPath)
		}
	}
	sort.Strings(remaining)
	orderedImports = append(orderedImports, remaining...)

	for _, importPath := range orderedImports {
		switch {
		case strings.HasPrefix(importPath, "./"):
			data, err := os.ReadFile(strings.TrimPrefix(importPath, "./"))
			if err != nil {
				panic(err)
			}
			if strings.HasSuffix(importPath, ".osl") {
				compiledBlock := CompileBlock(scriptToAst(string(data)), ctx)
				compiled += "\n" + compiledBlock
			} else if strings.HasSuffix(importPath, ".go") {
				compiled += "\n" + string(data)
			}

		case strings.HasPrefix(importPath, "osl/"):
			packageName := strings.TrimPrefix(importPath, "osl/")
			data, err := packagesFS.ReadFile("packages/" + packageName + ".go")
			if err != nil {
				panic(err)
			}
			file := strings.TrimSpace(string(data))

			for _, line := range strings.Split(file, "\n") {
				if requires, ok := strings.CutPrefix(line, "// requires: "); ok {
					for _, part := range strings.Split(requires, ", ") {
						part = strings.TrimSpace(part)
						goImports = append(goImports, part)
					}
				}
			}
			if importPath == "osl/window" {
				fontUrl := "https://raw.githubusercontent.com/Mistium/Origin-OS/main/Fonts/origin.ojff"

				resp, err := http.Get(fontUrl)
				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()

				data, err := io.ReadAll(resp.Body)
				if err != nil {
					panic(err)
				}

				var fontMap map[string]any
				err = json.Unmarshal(data, &fontMap)
				if err != nil {
					panic(err)
				}
				delete(fontMap, "origin")

				compiled = "\n\nvar OSLfont = map[string]string" + JsonStringify(fontMap) + "\n\n" + compiled
			}
			compiled = "\n" + file + compiled

		default:
			goImports = append(goImports, importPath)
		}
	}

	return compiled, goImports
}

func Compile(ast [][]*Token) string {
	ctx := &VariableContext{
		Globals: make(map[string]any),
		Locals:  make(map[string]any),
		Indent:  0,
		Prepend: make(map[string]string),
		Imports: map[string]bool{
			"fmt":           true,
			"math/rand":     true,
			"strconv":       true,
			"strings":       true,
			"bytes":         true,
			"encoding/json": true,
			"bufio":         true,
			"os":            true,
			"reflect":       true,
			"io":            true,
			"time":          true,
			"math":          true,
			"runtime":       true,
			"sort":          true,
			"unsafe":        true,
			"sync":          true,
		},
		ImportOrder: []string{
			"fmt",
			"math/rand",
			"strconv",
			"strings",
			"bytes",
			"encoding/json",
			"bufio",
			"os",
			"reflect",
			"io",
			"time",
			"math",
			"runtime",
			"sort",
			"unsafe",
			"sync",
		},
		ImportAliases: map[string]string{
			"io":        "OSLio",
			"math/rand": "OSLrand",
		},
		DeclaredVars:        make(map[string]bool),
		VariableTypes:       make(map[string]string),
		GlobalDeclaredVars:  make(map[string]bool),
		GlobalVariableTypes: make(map[string]string),
		functionReturnTypes: make(map[string]string),
		selfTypes:           make(map[string]string),
		selfUsed:            false,
		CustomCommands:      make(map[string]bool),
		builtinTypeMethods:  make(map[string]map[string]MethodDefinition),
		IsInit:              false,
	}

	var init [][]*Token
	var main [][]*Token

	pivot := -1
	for i, line := range ast {
		if line[0].Type == TKN_CMD && line[0].Data == "mainloop:" {
			pivot = i
			break
		}
	}

	var topLevelFuncs [][]*Token
	var initNoFuncs [][]*Token
	initCompiled := ""
	mainCompiled := ""

	if pivot >= 0 {
		init = ast[:pivot]
		main = make([][]*Token, len(ast)-pivot-1)
		copy(main, ast[pivot+1:])

		hasDrawingCommands := HasDrawingCommands(append(init, main...))

		if hasDrawingCommands {
			ctx.Imports["osl/window"] = true
			ctx.ImportOrder = append([]string{"osl/window"}, ctx.ImportOrder...)
		}

		for _, line := range init {
			if len(line) > 0 {
				// Extract function assignments
				if line[0].Type == TKN_ASI {
					if line[0].Right != nil && line[0].Right.Type == TKN_FNC && line[0].Left.Type == TKN_VAR &&
						line[0].Right.Data == "function" && ctx.Indent == 0 {
						topLevelFuncs = append(topLevelFuncs, line)
						continue
					}
				}
				// Extract def commands
				if line[0].Type == TKN_CMD && line[0].Data == "def" {
					topLevelFuncs = append(topLevelFuncs, line)
					continue
				}
				initNoFuncs = append(initNoFuncs, line)
			}
		}

		// Compile the initialization first to populate global variables
		ctx.IsInit = true
		ctx.GlobalVars = strings.Builder{}
		initCompiled = CompileBlock(initNoFuncs, ctx)

		// Save the global state for functions
		ctx.GlobalDeclaredVars = make(map[string]bool)
		maps.Copy(ctx.GlobalDeclaredVars, ctx.DeclaredVars)
		ctx.GlobalVariableTypes = make(map[string]string)
		maps.Copy(ctx.GlobalVariableTypes, ctx.VariableTypes)

		ctx.IsInit = false
		funcsCompiled := ""
		if len(topLevelFuncs) > 0 {
			funcsCompiled = CompileBlock(topLevelFuncs, ctx)
		}

		funcsCompiled = ctx.GlobalVars.String() + funcsCompiled

		mainCompiled = CompileBlock(main, ctx)

		if hasDrawingCommands {
			mainCompiled = funcsCompiled + "\nfunc main() {\n\twindow.Create(OSLsetup)\n}\n\nfunc OSLsetup(window *OSLWindow) {\n" + initCompiled + "\twindow.loop = func(window *OSLWindow) {\n" + AddIndent(mainCompiled, 2) + "\n\t}\n}\n"
		} else {
			mainCompiled = funcsCompiled + initCompiled + "func main() {\n" + mainCompiled + "}\n\n"
		}
	} else {
		init = ast
		main = [][]*Token{}
	}

	// Process imports and prepends

	importsCompiled, goImports := processImports(ctx)

	var prepend strings.Builder
	for _, v := range ctx.Prepend {
		fmt.Fprintf(&prepend, "%v", v)
	}

	if len(goImports) > 0 {
		seen := make(map[string]bool)
		var uniqueImports []string

		for _, pkg := range goImports {
			if seen[pkg] {
				continue
			}
			seen[pkg] = true

			if strings.Contains(pkg, " as ") {
				parts := strings.Split(pkg, " as ")
				actualPkg := strings.TrimSpace(parts[0])
				alias := strings.TrimSpace(parts[1])
				uniqueImports = append(uniqueImports, fmt.Sprintf("%s %q", alias, actualPkg))
			} else {
				if alias, ok := ctx.ImportAliases[pkg]; ok {
					uniqueImports = append(uniqueImports, fmt.Sprintf("%s %q", alias, pkg))
				} else {
					uniqueImports = append(uniqueImports, fmt.Sprintf("%q", strings.TrimSpace(pkg)))
				}
			}
		}

		prepend.WriteString("import (\n")
		for _, imp := range uniqueImports {
			prepend.WriteString("\t" + imp + "\n")
		}
		prepend.WriteString(")\n\n")
		prepend.WriteString("var OSLwincreatetime float64 = OSLcastNumber(time.Now().UnixMilli())\n")
		prepend.WriteString("var OSLsystem_os = runtime.GOOS\n")
		prepend.WriteString("var OSLtimer func() float64 = func() float64 { return float64(time.Now().Unix()) - (OSLwincreatetime / 1000) }\n")
		prepend.WriteString("var OSLtimestamp func() int64 = func() int64 { return time.Now().UnixMilli() }\n")
		prepend.WriteString("var timer float64\n")
		prepend.WriteString("var timestamp int64\n")
		prepend.WriteString("func OSLupdateTimer() {\n\ttimer = OSLtimer()\n\ttimestamp = OSLtimestamp()\n}\n\n")
		prepend.WriteString(include("packages/std.go"))
	}

	var methodsCompiled strings.Builder
	for typeName, methods := range ctx.builtinTypeMethods {
		goType := oslTypes[typeName]
		capitalizedTypeName := strings.ToUpper(string(typeName[0])) + typeName[1:]
		for methodName, methodDef := range methods {
			capitalizedMethodName := strings.ToUpper(string(methodName[0])) + methodName[1:]
			funcName := fmt.Sprintf("OSLMET%s%s", capitalizedTypeName, capitalizedMethodName)

			funcNode := methodDef.Token
			var params_string string
			if len(funcNode.Parameters) > 0 {
				args := strings.Split(funcNode.Parameters[0].Data.(string), ",")
				for _, arg := range args {
					arg = strings.TrimSpace(arg)
					if arg == "" {
						continue
					}
					parts := strings.Split(arg, " ")
					typeNameValue := goType // Default to receiver's type
					varName := arg
					if len(parts) > 1 {
						typeNameValue = mapOSLTypeToGo(parts[1])
						varName = parts[0]
					}
					params_string += fmt.Sprintf("%v %v, ", varName, typeNameValue)
				}
			}
			params_string = strings.TrimSuffix(params_string, ", ")

			returnType := ""
			if methodDef.ReturnType != "" {
				returnType = mapOSLTypeToGo(methodDef.ReturnType) + " "
			}

			funcBody := ""
			var blockData [][]*Token
			if len(funcNode.Parameters) > 1 && funcNode.Parameters[1] != nil {
				if bd, ok := funcNode.Parameters[1].Data.([][]*Token); ok {
					blockData = bd
					savedDeclaredVars := ctx.DeclaredVars
					ctx.DeclaredVars = make(map[string]bool)
					funcBody = CompileBlock(blockData, ctx)
					if ctx.selfUsed {
						funcBody = AddIndent("OSLself := self\n", ctx.Indent*2) + funcBody
						ctx.selfUsed = false
					}
					ctx.DeclaredVars = savedDeclaredVars
				}
			}

			hasReturn := hasReturnStatement(blockData)
			// Methods without explicit return type always return any
			if returnType == "" {
				returnType = "any "
			}

			// If no return statement and return type is any, add return nil
			if returnType == "any " && !hasReturn {
				funcBody += AddIndent("return nil\n", ctx.Indent*2)
			}

			funcSignature := fmt.Sprintf("func %v(self %v", funcName, goType)
			if params_string != "" {
				funcSignature += ", " + params_string
			}
			fmt.Fprintf(&methodsCompiled, "%v) %v{\n%v}\n\n", funcSignature, returnType, funcBody)
		}
	}

	return prepend.String() + methodsCompiled.String() + importsCompiled + mainCompiled
}

func HasDrawingCommands(ast [][]*Token) bool {
	drawingCommands := map[string]bool{
		"c": true, "color": true, "colour": true,
		"goto": true, "change": true, "change_x": true, "change_y": true,
		"loc":       true,
		"square":    true,
		"icon":      true,
		"text":      true,
		"centext":   true,
		"turnright": true, "turnleft": true, "pointat": true,
		"direction": true,
	}

	for _, line := range ast {
		if len(line) > 0 && line[0].Type == TKN_CMD {
			cmdName := fmt.Sprintf("%v", line[0].Data)
			if drawingCommands[cmdName] {
				return true
			}
		}
		for _, token := range line {
			if token.Type == TKN_MOD {
				return true
			}
		}
	}
	return false
}

func hasReturnStatement(block [][]*Token) bool {
	for _, line := range block {
		if len(line) > 0 && line[0].Type == TKN_CMD && line[0].Data == "return" {
			return true
		}

		if len(line) > 0 && line[0].Type == TKN_BLK {
			if subBlock, ok := line[0].Data.([][]*Token); ok {
				if hasReturnStatement(subBlock) {
					return true
				}
			}
		}

		if len(line) > 0 && line[0].Type == TKN_FNC && line[0].Data == "function" {
			if len(line[0].Parameters) > 1 && line[0].Parameters[1] != nil {
				if subBlock, ok := line[0].Parameters[1].Data.([][]*Token); ok {
					if hasReturnStatement(subBlock) {
						return true
					}
				}
			}
		}
	}
	return false
}

func OSLcastToString(expr string, typeHint string) string {
	if typeHint == TYPE_STR {
		return expr
	}
	return fmt.Sprintf("OSLtoString(%v)", expr)
}

func collectVariableDeclarations(block [][]*Token, ctx *VariableContext) map[string]string {
	varDecls := make(map[string]string)
	savedDeclaredVars := make(map[string]bool)
	for k, v := range ctx.DeclaredVars {
		savedDeclaredVars[k] = v
	}

	for _, line := range block {
		if len(line) == 0 {
			continue
		}

		if line[0].Type == TKN_ASI {
			if line[0].Left != nil && line[0].Left.Type == TKN_VAR {
				varName := line[0].Left.Data.(string)
				if !savedDeclaredVars[varName] && !ctx.GlobalDeclaredVars[varName] {
					var varType string = "any"
					if line[0].SetType != "" {
						varType = mapOSLTypeToGo(line[0].SetType)
					} else if line[0].Right != nil && line[0].Right.ReturnedType != "" {
						switch line[0].Right.ReturnedType {
						case TYPE_STR:
							varType = "string"
						case TYPE_INT:
							varType = "int"
						case TYPE_NUM:
							varType = "float64"
						case TYPE_BOOL:
							varType = "bool"
						case TYPE_ARR:
							varType = "[]any"
						case TYPE_OBJ:
							varType = "map[string]any"
						}
					}
					varDecls[varName] = varType
				}
			}
		}

		if line[0].Type == TKN_BLK {
			blk := line[0].Data.([][]*Token)
			for _, subLine := range blk {
				if len(subLine) > 0 && subLine[0].Type == TKN_CMD {
					if data, ok := subLine[0].Data.(string); ok && data == "def" {
						if len(subLine) > 2 {
							argsStr := ""
							if subLine[1].Type == TKN_STR {
								argsStr = subLine[1].Data.(string)
							}
							args := strings.Split(argsStr, ",")
							for _, arg := range args {
								arg = strings.TrimSpace(arg)
								if arg == "" {
									continue
								}
								savedDeclaredVars[arg] = true
							}
						}
					}
				}
			}
		}

		if line[0].Type == TKN_BLK {
			blk := line[0].Data.([][]*Token)
			saveIndent := ctx.Indent
			ctx.Indent++
			subDecls := collectVariableDeclarations(blk, ctx)
			ctx.Indent = saveIndent
			for k, v := range subDecls {
				varDecls[k] = v
			}
		}
	}

	return varDecls
}

func CompileBlock(block [][]*Token, ctx *VariableContext) string {
	var out strings.Builder

	varDecls := collectVariableDeclarations(block, ctx)

	for varName, varType := range varDecls {
		if !ctx.DeclaredVars[varName] {
			ctx.DeclaredVars[varName] = true
			if varType != "" {
				ctx.VariableTypes[varName] = varType
				if ctx.IsInit && ctx.Indent == 0 {
					fmt.Fprintf(&ctx.GlobalVars, "var %v %v\n", varName, varType)
					ctx.GlobalDeclaredVars[varName] = true
				} else {
					out.WriteString(AddIndent(fmt.Sprintf("var %v %v\n", varName, varType), ctx.Indent*2))
				}
			} else {
				if ctx.IsInit && ctx.Indent == 0 {
					fmt.Fprintf(&ctx.GlobalVars, "var %v\n", varName)
					ctx.GlobalDeclaredVars[varName] = true
				} else {
					out.WriteString(AddIndent(fmt.Sprintf("var %v\n", varName), ctx.Indent*2))
				}
			}
		}
	}

	for _, line := range block {
		out.WriteString(AddIndent(CompileLine(line, ctx), ctx.Indent*2))
	}

	return out.String()
}

func CompileModifier(mod *Token, ctx *VariableContext) string {
	var cmdData []any
	if data, ok := mod.Data.([]any); ok {
		cmdData = data
	} else {
		cmdData = []any{mod.Data}
	}

	cmdName := ""
	var argToken *Token

	if len(cmdData) >= 2 {
		if firstPart, ok := cmdData[0].(string); ok {
			cmdName = firstPart
		}
		if tokenArg, ok := cmdData[1].(*Token); ok {
			if cmdName == "c" {
				newToken := &Token{
					Type:   TKN_STR,
					Data:   tokenArg.Source,
					Source: tokenArg.Source,
				}
				argToken = newToken
			} else {
				argToken = tokenArg
			}
		} else {
			argToken = &Token{
				Type:   TKN_NUM,
				Data:   cmdData[1],
				Source: fmt.Sprintf("%v", cmdData[1]),
			}
		}
	} else if len(cmdData) == 1 {
		if s, ok := cmdData[0].(string); ok {
			before, after, ok0 := strings.Cut(s, "#")
			if ok0 {
				cmdName = before
				argData := after
				argToken = &Token{
					Type:   TKN_STR,
					Data:   argData,
					Source: argData,
				}
			} else {
				cmdName = s
			}
		} else {
			cmdName = fmt.Sprintf("%v", cmdData[0])
		}
	}

	var fullCmdName string
	switch cmdName {
	case "c":
		fullCmdName = "color"
	case "chx":
		fullCmdName = "change_x"
	case "chy":
		fullCmdName = "change_y"
	default:
		fullCmdName = cmdName
	}

	cmdTokens := []*Token{
		{Type: TKN_CMD, Data: fullCmdName},
	}

	if argToken != nil {
		cmdTokens = append(cmdTokens, argToken)
	}

	return CompileCmd(cmdTokens, ctx)
}

func CompileLine(line []*Token, ctx *VariableContext) string {
	var out string

	defer func() {
		if r := recover(); r != nil {
			lineInfo := ""
			if len(line) > 0 && line[0].Line > 0 {
				lineInfo = fmt.Sprintf("Line %d: ", line[0].Line)
			}
			panic(fmt.Sprintf("%s%v", lineInfo, r))
		}
	}()

	var modifiers []*Token
	var mainLine []*Token

	for _, token := range line {
		if token.Type == TKN_MOD {
			modifiers = append(modifiers, token)
		} else {
			mainLine = append(mainLine, token)
		}
	}

	for _, mod := range modifiers {
		out += CompileModifier(mod, ctx)
	}

	if len(mainLine) == 0 {
		return out + "\n"
	}

	if mainLine[0].Type == TKN_CMD {
		out += CompileCmd(mainLine, ctx)
		return out
	}
	for _, token := range mainLine {
		out += CompileToken(token, ctx)
	}
	return out + "\n"
}

func CompileToken(token *Token, ctx *VariableContext) string {
	if token == nil {
		return ""
	}

	defer func() {
		if r := recover(); r != nil {
			lineInfo := ""
			if token.Line > 0 {
				lineInfo = fmt.Sprintf("Line %d: ", token.Line)
			}
			tokenInfo := ""
			if token.Source != "" {
				tokenInfo = fmt.Sprintf(" in '%s'", token.Source)
			}
			panic(fmt.Sprintf("%s%s: %v", lineInfo, tokenInfo, r))
		}
	}()

	switch token.Type {
	case TKN_ASI:
		if token.Right != nil && token.Left != nil &&
			token.Right.Type == TKN_FNC && token.Left.Type == TKN_VAR &&
			token.Right.Data == "function" && ctx.Indent == 0 {
			ctx.Indent++
			ctx.ScopeLevel++

			savedDeclaredVars := ctx.DeclaredVars
			ctx.DeclaredVars = make(map[string]bool)

			params_string := ""
			if len(token.Right.Parameters) > 0 {
				params := token.Right.Parameters
				args := strings.Split(params[0].Data.(string), ",")
				for i, arg := range args {
					args[i] = strings.TrimSpace(arg)
					if args[i] == "" {
						continue
					}
					parts := strings.Split(args[i], " ")
					typeName := "any"
					varName := args[i]
					if len(parts) > 1 {
						typeName = mapOSLTypeToGo(parts[1])
						varName = parts[0]
					}

					ctx.DeclaredVars[varName] = true
					ctx.VariableTypes[varName] = typeName
					params_string += fmt.Sprintf("%v %v, ", varName, typeName)
				}
			}
			params_string = strings.TrimSuffix(params_string, ", ")
			returns := ""
			if token.Right.Returns != "" {
				returns = mapOSLTypeToGo(token.Right.Returns) + " "
			}

			funcBody := ""
			var blockData [][]*Token
			if len(token.Right.Parameters) > 1 && token.Right.Parameters[1] != nil {
				if bd, ok := token.Right.Parameters[1].Data.([][]*Token); ok {
					blockData = bd
					funcBody = CompileBlock(blockData, ctx)
					if ctx.selfUsed {
						params_string = "OSLself any, " + params_string
						ctx.selfUsed = false
					}
				}
			}

			hasReturn := hasReturnStatement(blockData)

			if returns == "" && hasReturn {
				returns = "any "
			}

			out := "func " + token.Left.Data.(string) + "(" +
				params_string + ") " + returns + "{\n" +
				funcBody

			if returns == "any " && hasReturn && !strings.Contains(funcBody, "return ") {
				out += "return nil\n"
			}

			out += "}"

			ctx.DeclaredVars = savedDeclaredVars
			ctx.Indent--
			ctx.ScopeLevel--
			return out
		}

		compiledLeft := CompileToken(token.Left, ctx)
		compiledRight := CompileToken(token.Right, ctx)
		op := ""
		switch token.Data {
		case "@=":
			op = "="
		case "=":
			op = "="
		case ":=":
			op = ":="
		case "++":
			op = "+="
			compiledRight = "1"
		case "--":
			op = "-="
			compiledRight = "1"
		case "++=":
			op = "+="
		case "+=":
			op = "+="
		case "-=":
			op = "-="
		case "*=":
			op = "*="
		case "/=":
			op = "/="
		case "%=":
			op = "%="
		case "=??":
			return compiledRight
		}

		varName := ""
		if token.Left.Type == TKN_VAR {
			if leftData, ok := token.Left.Data.(string); ok {
				varName = leftData
			}
		}
		if token.Left.Type == TKN_RMT {
			var objPath string
			path := token.Left.ObjPath
			if len(path) > 0 {
				tkn := &Token{
					Type: TKN_MTD,
					Data: path[0 : len(path)-1],
				}
				objPath = CompileToken(tkn, ctx)
			} else {
				objPath = "nil"
			}
			var keyExpr *Token
			isRaw := false
			if token.Left.Final != nil && token.Left.Final.Type == TKN_MTV {
				if methodName, ok := token.Left.Final.Data.(string); ok && methodName == "item" {
					if len(token.Left.Final.Parameters) > 0 {
						keyExpr = token.Left.Final.Parameters[0]
					} else {
						keyExpr = token.Left.Final
					}
				} else {
					keyExpr = token.Left.Final
				}
			} else {
				isRaw = true
				keyExpr = token.Left.Final
			}

			keyStr := CompileToken(keyExpr, ctx)

			typeStr, hasType := oslTypes[objPath]
			if ctx.Indent == 0 && hasType && keyExpr.Type == TKN_VAR && token.Right.Type == TKN_FNC && token.Right.Data == "function" {
				if token.Data == "@=" {
					if methodName, ok := keyExpr.Data.(string); ok {
						if ctx.builtinTypeMethods[objPath] == nil {
							ctx.builtinTypeMethods[objPath] = make(map[string]MethodDefinition)
						}
						ctx.builtinTypeMethods[objPath][methodName] = MethodDefinition{
							TypeName:   objPath,
							MethodName: methodName,
							ReturnType: token.Right.Returns,
							Token:      token.Right,
						}
						return ""
					}
				}
				funcPart := strings.TrimPrefix(strings.TrimSuffix(compiledRight, ")"), "(func")
				return fmt.Sprintf("func (OSLself %v) %v%v", typeStr, keyExpr.Data, funcPart)
			} else if isRaw {
				keyStr = JsonStringify(keyStr)
			}

			return fmt.Sprintf("OSLsetItem(%v, %v, %v)", objPath, keyStr, compiledRight)
		}

		if varName != "" {
			declared := ctx.DeclaredVars[varName]
			if !declared {
				ctx.DeclaredVars[varName] = true
			}
			if token.SetType != "" {
				tokenType := token.SetType
				goType := mapOSLTypeToGo(tokenType)

				ctx.VariableTypes[varName] = goType

				switch tokenType {
				case "string":
					if token.Right.ReturnedType != TYPE_STR {
						compiledRight = fmt.Sprintf("OSLtoString(%v)", compiledRight)
					}
				case "int":
					if token.Right.ReturnedType == TYPE_NUM {
						if after, ok := strings.CutPrefix(compiledRight, "OSLcastNumber("); ok {
							compiledRight = fmt.Sprintf("OSLcastInt(%v", after)
						} else {
							compiledRight = fmt.Sprintf("int(%v)", compiledRight)
						}
					} else if token.Right.ReturnedType != TYPE_INT {
						compiledRight = fmt.Sprintf("OSLcastInt(%v)", compiledRight)
					}
				case "number":
					if token.Right.ReturnedType != TYPE_NUM {
						compiledRight = fmt.Sprintf("OSLcastNumber(%v)", compiledRight)
					}
				case "boolean":
					if token.Right.ReturnedType != TYPE_BOOL {
						compiledRight = fmt.Sprintf("bool(%v)", compiledRight)
					}
				case "array":
					ctx.VariableTypes[varName] = "[]any"
					if token.Right.ReturnedType != TYPE_ARR {
						compiledRight = fmt.Sprintf("OSLcastArray(%v)", compiledRight)
					}
				case "object":
					ctx.VariableTypes[varName] = "map[string]any"
					if token.Right.ReturnedType != TYPE_OBJ {
						compiledRight = fmt.Sprintf("OSLcastObject(%v)", compiledRight)
					}
				}
				if ctx.IsInit && ctx.Indent == 0 && op == "=" {
					fmt.Fprintf(&ctx.GlobalVars, "var %v %v = %v\n", varName, goType, compiledRight)
					ctx.GlobalDeclaredVars[varName] = true
					return ""
				}
				if goType == "auto" && op == "=" {
					return fmt.Sprintf("%v := %v", varName, compiledRight)
				}
				return fmt.Sprintf("%v := %v", varName, compiledRight)
			}
			if !declared {
				inferredType := ""
				if strings.HasPrefix(compiledRight, "OSLtoString(") {
					inferredType = "string"
				} else if strings.HasPrefix(compiledRight, "OSLcastNumber(") {
					inferredType = "float64"
				} else if strings.HasPrefix(compiledRight, "OSLcastInt(") {
					inferredType = "int"
				} else if strings.HasPrefix(compiledRight, "OSLcastBool(") {
					inferredType = "bool"
				} else if strings.HasPrefix(compiledRight, "OSLcastObject(") {
					inferredType = "map[string]any"
				}
				if inferredType != "" {
					ctx.VariableTypes[varName] = inferredType
				}
			} else {
				leftVar := token.Left.Data.(string)
				expectedType := ctx.VariableTypes[leftVar]

				if expectedType != "" && strings.HasPrefix(compiledRight, "OSLgetItem(") {
					switch expectedType {
					case "map[string]any":
						compiledRight = fmt.Sprintf("%v.(map[string]any)", compiledRight)
					case "[]any":
						compiledRight = fmt.Sprintf("%v.([]any)", compiledRight)
					case "string":
						compiledRight = fmt.Sprintf("OSLtoString(%v)", compiledRight)
					case "int":
						compiledRight = fmt.Sprintf("OSLcastInt(%v)", compiledRight)
					case "float64":
						compiledRight = fmt.Sprintf("OSLcastNumber(%v)", compiledRight)
					case "bool":
						compiledRight = fmt.Sprintf("OSLcastBool(%v)", compiledRight)
					}
				}
			}

			var varOut string
			if op == ":=" {
				varOut = fmt.Sprintf("var %v = %v", varName, compiledRight)
			} else if op == "=" && !declared && !ctx.GlobalDeclaredVars[varName] && token.SetType == "" {
				varOut = fmt.Sprintf("var %v = %v", varName, compiledRight)
				if ctx.IsInit && ctx.Indent == 0 {
					ctx.GlobalVars.WriteString(varOut + "\n")
					ctx.GlobalDeclaredVars[varName] = true
					return ""
				}
			} else {
				varOut = fmt.Sprintf("%v %v %v", varName, op, compiledRight)
			}
			return varOut
		}
		return fmt.Sprintf("%v %v %v", compiledLeft, op, compiledRight)
	case TKN_OPR:
		if token.Data == "//" {
			return ""
		}

		compiledLeft := CompileToken(token.Left, ctx)
		compiledRight := CompileToken(token.Right, ctx)

		LT := token.Left.ReturnedType
		RT := token.Right.ReturnedType

		switch token.Data {
		case "??":
			return fmt.Sprintf("OSLnullishCoaless(%v, %v)", compiledLeft, compiledRight)
		case "+":
			if isNumberCompatible(LT) && isNumberCompatible(RT) {
				if LT == TYPE_INT {
					compiledLeft = fmt.Sprintf("float64(%v)", compiledLeft)
				}
				if RT == TYPE_INT {
					compiledRight = fmt.Sprintf("float64(%v)", compiledRight)
				}

				token.ReturnedType = TYPE_NUM
				return fmt.Sprintf("(%v + %v)", compiledLeft, compiledRight)
			}
			if LT == TYPE_STR || RT == TYPE_STR {
				token.ReturnedType = TYPE_STR
				return fmt.Sprintf("(%v + \" \" + %v)", OSLcastToString(compiledLeft, LT), OSLcastToString(compiledRight, RT))
			}
			return fmt.Sprintf("OSLadd(%v, %v)", compiledLeft, compiledRight)
		case "-":
			if LT != "" || RT != "" {
				token.ReturnedType = TYPE_NUM
				return fmt.Sprintf("(%v - %v)", compiledLeft, compiledRight)
			}
			return fmt.Sprintf("OSLsub(%v, %v)", compiledLeft, compiledRight)
		case "*":
			return fmt.Sprintf("OSLmultiply(%v, %v)", compiledLeft, compiledRight)
		case "/":
			return fmt.Sprintf("OSLdivide(%v, %v)", compiledLeft, compiledRight)
		case "%":
			return fmt.Sprintf("OSLmod(%v, %v)", compiledLeft, compiledRight)
		case "++":
			if isAbsolutelyNot(LT, TYPE_ARR) || isAbsolutelyNot(RT, TYPE_ARR) {
				token.ReturnedType = TYPE_STR
				return fmt.Sprintf("(%v + %v)", OSLcastToString(compiledLeft, LT), OSLcastToString(compiledRight, RT))
			}
			return fmt.Sprintf("OSLjoin(%v, %v)", compiledLeft, compiledRight)
		}
		return fmt.Sprintf("%v %v %v", compiledLeft, token.Data, compiledRight)
	case TKN_EVL:
		return CompileToken(token.Data.(*Token), ctx)
	case TKN_CMP:
		compiledLeft := CompileToken(token.Left, ctx)
		compiledRight := CompileToken(token.Right, ctx)
		switch token.Data {
		case "!=":
			return fmt.Sprintf("OSLnotEqual(%v, %v)", compiledLeft, compiledRight)
		case "==":
			return fmt.Sprintf("OSLequal(%v, %v)", compiledLeft, compiledRight)
		case "===":
			return fmt.Sprintf("%v == %v", compiledLeft, compiledRight)
		case "!==":
			return fmt.Sprintf("%v != %v", compiledLeft, compiledRight)
		case ">", "<", "<=", ">=":
			return fmt.Sprintf("OSLcastNumber(%v) %v OSLcastNumber(%v)", compiledLeft, token.Data, compiledRight)
		}
		return fmt.Sprintf("%v %v %v", compiledLeft, token.Data, compiledRight)
	case TKN_LOG:
		left := CompileToken(token.Left, ctx)
		right := CompileToken(token.Right, ctx)
		op := ""
		token.ReturnedType = TYPE_BOOL
		switch token.Data {
		case "and":
			if left == "true" {
				return right
			}
			if right == "true" {
				return left
			}
			if left == "false" || right == "false" {
				return "false"
			}
			op = "&&"
		case "or":
			if left == "false" {
				return right
			}
			if right == "false" {
				return left
			}
			if left == "true" || right == "true" {
				return "true"
			}
			op = "||"
		}
		return fmt.Sprintf("%v %v %v", left, op, right)
	case TKN_BIT:
		token.ReturnedType = TYPE_BOOL
		return fmt.Sprintf("%v %v %v", CompileToken(token.Left, ctx), token.Data, CompileToken(token.Right, ctx))
	case TKN_STR:
		token.ReturnedType = TYPE_STR
		return JsonStringify(token.Data)
	case TKN_NUM:
		token.ReturnedType = TYPE_NUM
		return fmt.Sprintf("%v", token.Data)
	case TKN_VAR:
		varName := token.Data.(string)
		if strings.HasPrefix(varName, "OSL") {
			panic("Cannot use reserved variable name: " + varName)
		}
		switch varName {
		case "self":
			ctx.selfUsed = true
			return "OSLself"
		case "null":
			return "nil"
		case "timestamp":
			token.ReturnedType = TYPE_NUM
			return "float64(time.Now().UnixMilli())"
		case "performance":
			token.ReturnedType = TYPE_NUM
			return "float64(time.Now().UnixMicro())"
		}
		return varName
	case TKN_RAW:
		switch v := token.Data.(type) {
		case bool:
			token.ReturnedType = TYPE_BOOL
			return fmt.Sprintf("%t", v)
		case string:
			token.ReturnedType = TYPE_STR
			return v
		default:
			return fmt.Sprintf("%v", v)
		}
	case TKN_BLK:
		ctx.Indent++
		blk := CompileBlock(token.Data.([][]*Token), ctx)
		ctx.Indent--
		return fmt.Sprintf("(\n%v\n)", blk)
	case TKN_ARR:
		ctx.Indent++
		arr := CompileArray(token.Data.([]*Token), ctx)
		ctx.Indent--
		token.ReturnedType = TYPE_ARR
		return arr
	case TKN_OBJ:
		if token.Data == nil {
			return "map[string]any{}"
		}
		ctx.Indent++
		obj := CompileObject(token.Data.([][]*Token), ctx)
		ctx.Indent--
		token.ReturnedType = TYPE_OBJ
		return obj
	case TKN_TSR:
		var compiled strings.Builder
		for _, t := range token.Data.([]*Token) {
			if t.Type == TKN_STR {
				compiled.WriteString("+\"" + t.Data.(string) + "\"")
			} else {
				compiled.WriteString("+" + CompileToken(t, ctx))
			}
		}
		token.ReturnedType = TYPE_STR
		return strings.TrimSpace(strings.TrimPrefix(compiled.String(), "+"))
	case TKN_RMT:
		var path []*Token
		if token.Data != nil {
			if p, ok := token.Data.([]*Token); ok && p != nil {
				path = p
			} else {
				path = []*Token{}
			}
		} else {
			path = []*Token{}
		}
		out := ""
		if len(path) > 0 {
			out += CompileToken(path[0], ctx)
		}

		var keyExpr string
		if token.Final != nil && token.Final.Type == TKN_MTV {
			if methodName, ok := token.Final.Data.(string); ok && methodName == "item" {
				if len(token.Final.Parameters) > 0 {
					keyExpr = CompileToken(token.Final.Parameters[0], ctx)
				} else {
					keyExpr = CompileToken(token.Final, ctx)
				}
			} else {
				keyExpr = CompileToken(token.Final, ctx)
			}
		} else {
			keyExpr = CompileToken(token.Final, ctx)
		}

		out = fmt.Sprintf("OSLgetItem(%v, %v)", out, keyExpr)

		return out
	case TKN_FNC:
		params := token.Parameters
		if params == nil {
			params = []*Token{}
		}
		switch token.Data {
		case "function":
			var paramString strings.Builder
			if len(params) > 0 {
				args := strings.Split(params[0].Data.(string), ",")
				for i, arg := range args {
					args[i] = strings.TrimSpace(arg)
					if args[i] == "" {
						continue
					}
					parts := strings.Split(args[i], " ")
					typeName := "any"
					varName := args[i]
					if len(parts) > 1 {
						typeName = mapOSLTypeToGo(parts[1])
						varName = parts[0]
					}
					fmt.Fprintf(&paramString, "%v %v, ", varName, typeName)
				}
			}
			returns := " "
			// Inline functions without explicit return type always return any
			if len(params) > 1 && token.Returns == "" {
				returns = "any "
			} else if len(params) > 1 {
				returns = mapOSLTypeToGo(token.Returns) + " "
			} else if len(params) <= 1 {
				returns = "any "
			}
			out := fmt.Sprintf("(func(%v) %v{\n", strings.TrimSuffix(paramString.String(), ", "), returns)

			ctx.Indent++
			ctx.ScopeLevel++
			savedDeclaredVars := ctx.DeclaredVars
			ctx.DeclaredVars = make(map[string]bool)

			if len(params) > 1 {
				blk := params[1]
				if blk != nil {
					inner := CompileBlock(blk.Data.([][]*Token), ctx)
					if ctx.selfUsed {
						inner = AddIndent("OSLself := OSLself\n", ctx.Indent*2) + inner
						ctx.selfUsed = false
					}

					// If no return statement and return type is any, add return nil
					hasReturn := hasReturnStatement(blk.Data.([][]*Token))
					if returns == "any " && !hasReturn {
						inner += AddIndent("return nil\n", ctx.Indent*2)
					}
					out += inner
				}
			}
			out += AddIndent("})", ctx.Indent*2-2)

			ctx.DeclaredVars = savedDeclaredVars
			ctx.Indent--
			ctx.ScopeLevel--
			return out
		case "worker":
			if len(token.Parameters) > 0 {
				return fmt.Sprintf("OSLworker(%v)", CompileToken(token.Parameters[0], ctx))
			}
			panic("worker osl function needs 1 parameter")
		case "typeof":
			if len(token.Parameters) > 0 {
				token.ReturnedType = TYPE_STR
				return fmt.Sprintf("OSLtypeof(%v)", CompileToken(token.Parameters[0], ctx))
			}
			panic("typeof osl function needs 1 parameter")
		case "delete":
			if len(token.Parameters) > 0 {
				return fmt.Sprintf("OSLdelete(%v, %v)", CompileToken(token.Parameters[0], ctx), CompileToken(token.Parameters[1], ctx))
			}
		case "round":
			if len(token.Parameters) > 0 {
				token.ReturnedType = TYPE_INT
				return fmt.Sprintf("OSLround(%v)", CompileToken(token.Parameters[0], ctx))
			}
		case "ceil":
			if len(token.Parameters) > 0 {
				token.ReturnedType = TYPE_INT
				return fmt.Sprintf("OSLceil(%v)", CompileToken(token.Parameters[0], ctx))
			}
		case "floor":
			if len(token.Parameters) > 0 {
				token.ReturnedType = TYPE_INT
				return fmt.Sprintf("OSLfloor(%v)", CompileToken(token.Parameters[0], ctx))
			}
		case "min":
			if len(token.Parameters) > 0 {
				token.ReturnedType = TYPE_NUM
				return fmt.Sprintf("OSLmin(%v, %v)", CompileToken(token.Parameters[0], ctx), CompileToken(token.Parameters[1], ctx))
			}
		case "max":
			if len(token.Parameters) > 0 {
				token.ReturnedType = TYPE_NUM
				return fmt.Sprintf("OSLmax(%v, %v)", CompileToken(token.Parameters[0], ctx), CompileToken(token.Parameters[1], ctx))
			}
		case "random":
			if len(token.Parameters) >= 2 {
				token.ReturnedType = TYPE_NUM
				return fmt.Sprintf("OSLrandom(%v, %v)", CompileToken(token.Parameters[0], ctx), CompileToken(token.Parameters[1], ctx))
			}
		case "raw":
			if len(token.Parameters) > 0 {
				return fmt.Sprintf("%v", token.Parameters[0].Data)
			}
		default:
			nameStr := token.Data.(string)
			_, ok := oslTypes[nameStr]
			if ok {
				return "OSL_new_" + nameStr + "()"
			}
			var paramString strings.Builder
			if len(token.Parameters) > 0 {
				for i, p := range params {
					paramString.WriteString(CompileToken(p, ctx))
					if i < len(token.Parameters)-1 {
						paramString.WriteString(", ")
					}
				}
			}
			functionReturnType, ok := allFunctionTypes[token.Data.(string)]
			if ok {
				token.ReturnedType = functionReturnType.Returns
			}
			return fmt.Sprintf("%v(%v)", token.Data, paramString.String())
		}
	case TKN_URY:
		op := token.Data
		value := CompileToken(token.Right, ctx)
		if op == "@" {
			op = "&"
		}
		if op == "!" {
			token.ReturnedType = TYPE_BOOL
			return fmt.Sprintf("(%v != true)", value)
		}
		return fmt.Sprintf("%v%v", op, value)
	case TKN_MTV:
		params := make([]string, len(token.Parameters))
		for i, p := range token.Parameters {
			params[i] = CompileToken(p, ctx)
		}
		return fmt.Sprintf("%v(%v)", token.Data, strings.Join(params, ", "))
	case TKN_MTD:
		out := ""
		var parts []*Token
		if token.Data != nil {
			if p, ok := token.Data.([]*Token); ok && p != nil {
				parts = p
			}
		}
		if len(parts) == 0 {
			return out
		}
		first := parts[0]
		if first.Type == TKN_VAR {
			if strings.HasPrefix(first.Data.(string), "OSL") {
				panic("Cannot use reserved variable name: " + first.Data.(string))
			}
		}
		out = CompileToken(first, ctx)
		previous := first
		parts = parts[1:]
		for _, part := range parts {
			name := part.Data.(string)
			switch part.Type {
			case TKN_VAR:
				switch name {
				case "len":
					part.ReturnedType = TYPE_INT
					out = fmt.Sprintf("OSLlen(%v)", out)
				default:
					if previous.Type == TKN_VAR && previous.Data.(string) == "self" {
						typeStr, hasType := ctx.selfTypes[name]
						if hasType {
							out = fmt.Sprintf("%v.%v", out, name)
							part.ReturnedType = typeStr
							break
						}
					}
					if previous.ReturnedType == TYPE_OBJ {
						out = fmt.Sprintf("%v[OSLtoString(%v)]", out, name)
					}
					typeName, isType := previous.Data.(string)
					if isType {
						if _, isBuiltin := oslTypes[typeName]; isBuiltin {
							if methods, hasMethods := ctx.builtinTypeMethods[typeName]; hasMethods {
								if methodDef, hasCustomMethod := methods[name]; hasCustomMethod {
									capitalizedTypeName := strings.ToUpper(string(typeName[0])) + typeName[1:]
									capitalizedMethodName := strings.ToUpper(string(name[0])) + name[1:]
									funcName := fmt.Sprintf("OSLMET%s%s", capitalizedTypeName, capitalizedMethodName)
									out = fmt.Sprintf("%v(%v", funcName, out)
									for _, p := range part.Parameters {
										out += ", " + CompileToken(p, ctx)
									}
									out += ")"
									part.ReturnedType = methodDef.ReturnType
									if part.ReturnedType != "" {
										token.ReturnedType = part.ReturnedType
									}
									break
								}
							}
						}
					}
					out = fmt.Sprintf("OSLgetItem(%v, \"%v\")", out, name)
				}
			case TKN_MTV:
				params := make([]string, len(part.Parameters))
				for i, p := range part.Parameters {
					params[i] = CompileToken(p, ctx)
				}

				if previous.ReturnedType != "" {
					builtinName := ""
					switch previous.ReturnedType {
					case TYPE_STR:
						builtinName = "string"
					case TYPE_INT:
						builtinName = "int"
					case TYPE_NUM:
						builtinName = "number"
					case TYPE_BOOL:
						builtinName = "boolean"
					case TYPE_OBJ:
						builtinName = "object"
					case TYPE_ARR:
						builtinName = "array"
					}
					if builtinName != "" {
						if methods, hasMethods := ctx.builtinTypeMethods[builtinName]; hasMethods {
							if methodDef, hasCustomMethod := methods[name]; hasCustomMethod {
								capitalizedTypeName := strings.ToUpper(string(builtinName[0])) + builtinName[1:]
								capitalizedMethodName := strings.ToUpper(string(name[0])) + name[1:]
								funcName := fmt.Sprintf("OSLMET%s%s", capitalizedTypeName, capitalizedMethodName)
								out = fmt.Sprintf("%v(%v", funcName, out)
								for _, p := range part.Parameters {
									out += ", " + CompileToken(p, ctx)
								}
								out += ")"
								part.ReturnedType = methodDef.ReturnType
								if part.ReturnedType != "" {
									token.ReturnedType = part.ReturnedType
								}
								previous = part
								continue
							}
						}
					}
				}

				switch name {
				case "call":
					out = fmt.Sprintf("OSLcallFunc(%v.%v, %v)", out, part.Parameters[0].Data, JsonStringify(params[1:]))
				case "toStr":
					part.ReturnedType = TYPE_STR
					out = fmt.Sprintf("OSLtoString(%v)", out)
				case "toInt":
					part.ReturnedType = TYPE_INT
					out = fmt.Sprintf("OSLcastInt(%v)", out)
				case "toNum":
					part.ReturnedType = TYPE_NUM
					out = fmt.Sprintf("OSLcastNumber(%v)", out)
				case "toBool":
					part.ReturnedType = TYPE_BOOL
					out = fmt.Sprintf("OSLcastBool(%v)", out)
				case "toArray":
					part.ReturnedType = TYPE_ARR
					out = fmt.Sprintf("OSLcastArray(%v)", out)
				case "toObject":
					part.ReturnedType = TYPE_OBJ
					out = fmt.Sprintf("OSLcastObject(%v)", out)
				case "pop":
					part.ReturnedType = TYPE_UNK
					out = fmt.Sprintf("OSLpop(&(%v))", out)
				case "shift":
					part.ReturnedType = TYPE_UNK
					out = fmt.Sprintf("OSLshift(&(%v))", out)
				case "to":
					if len(params) > 1 {
						part.ReturnedType = TYPE_ARR
						out = fmt.Sprintf("OSLto(%v, %v)", out, params[0])
					}
				case "append":
					if len(params) > 0 {
						part.ReturnedType = TYPE_ARR
						out = fmt.Sprintf("OSLappend(&(%v), %v)", out, params[0])
					}
				case "prepend":
					if len(params) > 0 {
						part.ReturnedType = TYPE_ARR
						out = fmt.Sprintf("OSLprepend(&(%v), %v)", out, params[0])
					}
				case "in":
					if len(params) > 0 {
						part.ReturnedType = TYPE_BOOL
						out = fmt.Sprintf("OSLKeyIn(%v, %v)", params[0], out)
					}
				case "ask":
					part.ReturnedType = TYPE_STR
					out = fmt.Sprintf("input(%v)", out)
				case "chr":
					part.ReturnedType = TYPE_STR
					out = fmt.Sprintf("string(rune(OSLcastInt(%v)))", out)
				case "ord":
					part.ReturnedType = TYPE_INT
					out = fmt.Sprintf("int(OSLtoString(%v)[0])", out)
				case "toLower":
					part.ReturnedType = TYPE_STR
					out = fmt.Sprintf("strings.ToLower(%v)", out)
				case "toUpper":
					part.ReturnedType = TYPE_STR
					out = fmt.Sprintf("strings.ToUpper(%v)", out)
				case "onKeyDown":
					part.ReturnedType = TYPE_BOOL
					out = "false" // Stub implementation
				case "isKeyDown":
					part.ReturnedType = TYPE_BOOL
					out = fmt.Sprintf("window.KeyPressed(%v)", out)
				case "getKeys":
					part.ReturnedType = TYPE_ARR
					out = fmt.Sprintf("OSLgetKeys(%v)", out)
				case "getValues":
					part.ReturnedType = TYPE_ARR
					out = fmt.Sprintf("OSLgetValues(%v)", out)
				case "floor":
					part.ReturnedType = TYPE_INT
					out = fmt.Sprintf("OSLfloor(%v)", out)
				case "ceil":
					part.ReturnedType = TYPE_INT
					out = fmt.Sprintf("OSLceil(%v)", out)
				case "round":
					part.ReturnedType = TYPE_INT
					out = fmt.Sprintf("OSLround(%v)", out)
				case "startsWith":
					if len(params) > 0 {
						part.ReturnedType = TYPE_BOOL
						out = fmt.Sprintf("strings.HasPrefix(%v, %v)", out, params[0])
					}
				case "endsWith":
					if len(params) > 0 {
						part.ReturnedType = TYPE_BOOL
						out = fmt.Sprintf("strings.HasSuffix(%v, %v)", out, params[0])
					}
				case "contains":
					if len(params) > 0 {
						part.ReturnedType = TYPE_BOOL
						out = fmt.Sprintf("OSLcontains(%v, %v)", out, params[0])
					}
				case "sort":
					part.ReturnedType = TYPE_ARR
					out = fmt.Sprintf("OSLsort(%v)", out)
				case "sortBy":
					if len(params) > 0 {
						part.ReturnedType = TYPE_ARR
						out = fmt.Sprintf("OSLsortBy(%v, %v)", out, params[0])
					}
				case "index":
					if len(params) > 0 {
						part.ReturnedType = TYPE_NUM
						out = fmt.Sprintf("(strings.Index(%v, %v) + 1)", out, params[0])
					}
				case "strip":
					part.ReturnedType = TYPE_STR
					out = fmt.Sprintf("strings.TrimSpace(%v)", out)
				case "clone":
					out = fmt.Sprintf("OSLclone(%v)", out)
				case "join":
					if len(params) > 0 {
						part.ReturnedType = TYPE_STR
						out = fmt.Sprintf("OSLarrayJoin(%v, %v)", out, params[0])
					}
				case "split":
					if len(params) > 0 {
						part.ReturnedType = TYPE_ARR
						out = fmt.Sprintf("OSLSplit(%v, %v)", out, params[0])
					}
				case "delete":
					if len(params) > 0 {
						out = fmt.Sprintf("OSLdelete(%v, %v)", out, params[0])
					}
				case "slice":
					if len(params) > 1 {
						out = fmt.Sprintf("OSLslice(%v, %v, %v)", out, params[0], params[1])
					} else if len(params) > 0 {
						out = fmt.Sprintf("OSLslice(%v, %v, -1)", out, params[0])
					}
				case "trim":
					if len(params) > 1 {
						out = fmt.Sprintf("OSLtrim(%v, %v, %v)", out, params[0], params[1])
					} else if len(params) > 0 {
						out = fmt.Sprintf("OSLtrim(%v, %v, -1)", out, params[0])
					} else {
						part.ReturnedType = TYPE_STR
						out = fmt.Sprintf("strings.TrimSpace(%v)", out)
					}
				case "JsonStringify":
					part.ReturnedType = TYPE_STR
					out = fmt.Sprintf("JsonStringify(%v)", out)
				case "JsonParse":
					out = fmt.Sprintf("JsonParse(%v)", out)
				case "JsonFormat":
					part.ReturnedType = TYPE_STR
					out = fmt.Sprintf("JsonFormat(%v)", out)
				case "stripStart":
					if len(params) > 0 {
						part.ReturnedType = TYPE_STR
						out = fmt.Sprintf("strings.TrimPrefix(%v, %v)", out, params[0])
					}
				case "stripEnd":
					if len(params) > 0 {
						part.ReturnedType = TYPE_STR
						out = fmt.Sprintf("strings.TrimSuffix(%v, %v)", out, params[0])
					}
				case "padStart":
					if len(params) > 1 {
						part.ReturnedType = TYPE_STR
						out = fmt.Sprintf("OSLpadStart(%v, int(OSLcastNumber(%v)), %v)", out, params[1], params[0])
					}
				case "padEnd":
					if len(params) > 1 {
						part.ReturnedType = TYPE_STR
						out = fmt.Sprintf("OSLpadEnd(%v, int(OSLcastNumber(%v)), %v)", out, params[1], params[0])
					}
				case "assert":
					if len(params) > 0 {
						paramStr := strings.Trim(params[0], "\"")

						goType := mapOSLTypeToGo(paramStr)
						if goType != paramStr {
							part.ReturnedType = paramStr
						}
						out = fmt.Sprintf("%v.(%v)", out, goType)
					}
				case "item":
					if len(params) > 0 {
						if previous.ReturnedType == TYPE_OBJ {
							out = fmt.Sprintf("%v[OSLtoString(%v)]", out, params[0])
						}
						out = fmt.Sprintf("OSLgetItem(%v, %v)", out, params[0])
					}
				case "sin":
					part.ReturnedType = TYPE_NUM
					out = fmt.Sprintf("math.Sin((OSLcastNumber(%v) * math.Pi) / 180)", out)
				case "cos":
					part.ReturnedType = TYPE_NUM
					out = fmt.Sprintf("math.Cos((OSLcastNumber(%v) * math.Pi) / 180)", out)
				case "tan":
					part.ReturnedType = TYPE_NUM
					out = fmt.Sprintf("math.Tan((OSLcastNumber(%v) * math.Pi) / 180)", out)
				case "clamp":
					if len(params) >= 2 {
						part.ReturnedType = TYPE_NUM
						out = fmt.Sprintf("OSLmin(OSLmax(%v, %v), %v)", out, params[0], params[1])
					}
				case "abs":
					part.ReturnedType = TYPE_NUM
					out = fmt.Sprintf("math.Abs(OSLcastNumber(%v))", out)
				case "sqrt":
					part.ReturnedType = TYPE_NUM
					out = fmt.Sprintf("math.Sqrt(OSLcastNumber(%v))", out)
				default:
					if len(part.Parameters) == 0 {
						out = fmt.Sprintf("%v.%v()", out, name)
						break
					}
					out = fmt.Sprintf("%v.%v(%v)", out, name, strings.Join(params, ", "))
				}
			}
			previous = part
		}
		token.ReturnedType = previous.ReturnedType
		return out
	case TKN_QST:
		if token.Left == nil || token.Right == nil || token.Right2 == nil {
			return "nil"
		}
		left := CompileToken(token.Left, ctx)
		right := CompileToken(token.Right, ctx)
		right2 := CompileToken(token.Right2, ctx)
		return fmt.Sprintf("func() any { if %v { return %v } else { return %v } }()", left, right, right2)
	case TKN_UNK:
		if data, ok := token.Data.(string); ok {
			if matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, data); matched {
				return data
			}
			return JsonStringify(data)
		}
		return "nil"
	}
	return fmt.Sprintf("<%v>", token.Type)
}

func CompileCmd(cmd []*Token, ctx *VariableContext) string {
	var out string
	switch cmd[0].Data {
	case "//":
		return ""
	case "if":
		if len(cmd) < 3 {
			panic("If command requires at least 2 parameters")
		}
		blk := cmd[2]
		if blk.Type != TKN_BLK {
			panic("If command requires a block")
		}
		var condition string
		if len(cmd) > 1 {
			condition = CompileToken(cmd[1], ctx)
		}
		if cmd[1].ReturnedType != TYPE_BOOL {
			condition = fmt.Sprintf("OSLcastBool(%v)", condition)
		}

		out += fmt.Sprintf("if %v {\n", condition)
		ctx.Indent++
		out += CompileBlock(blk.Data.([][]*Token), ctx)
		ctx.Indent--
		out += AddIndent("}", ctx.Indent*2)

		i := 3
		for i < len(cmd) {
			if i+1 < len(cmd) && cmd[i].Data == "else" && cmd[i+1].Data == "if" && i+3 < len(cmd) {
				blk := cmd[i+3]
				if blk.Type != TKN_BLK {
					panic("Else if command requires a block")
				}
				var condition string
				if i+2 < len(cmd) {
					condition = CompileToken(cmd[i+2], ctx)
				}
				out += fmt.Sprintf(" else if %v {\n", condition)
				ctx.Indent++
				out += CompileBlock(blk.Data.([][]*Token), ctx)
				ctx.Indent--
				out += AddIndent("}", ctx.Indent*2)
				i += 4
			} else if i+1 < len(cmd) && cmd[i].Data == "else" {
				blk := cmd[i+1]
				if blk.Type != TKN_BLK {
					panic("Else command requires a block")
				}
				out += " else {\n"
				ctx.Indent++
				if i+1 < len(cmd) {
					out += CompileBlock(cmd[i+1].Data.([][]*Token), ctx)
				}
				ctx.Indent--
				out += AddIndent("}", ctx.Indent*2)
				i += 2
				break
			} else {
				break
			}
		}
	case "loop":
		if len(cmd) < 3 {
			panic("Loop command requires at least 1 parameter")
		}
		var iteratorVar string = "i_" + RandomString(5)
		loopNumber := CompileToken(cmd[1], ctx)
		blk := cmd[2]
		var blockData [][]*Token
		if blkData, ok := blk.Data.([][]*Token); ok {
			blockData = blkData
		} else if dataStr, ok := blk.Data.(string); ok {
			blockData = scriptToAst(dataStr)
		} else if singleToken, ok := blk.Data.([]*Token); ok {
			blockData = append(blockData, singleToken)
		}
		ctx.Indent++

		var loopCondition string
		returnType := cmd[1].ReturnedType

		// Try to get variable type from context if it's a variable reference
		if cmd[1].Type == TKN_VAR {
			if varName, ok := cmd[1].Data.(string); ok {
				if varType, exists := ctx.VariableTypes[varName]; exists {
					if varType == "float64" {
						returnType = TYPE_NUM
					}
				}
			}
		}

		switch returnType {
		case TYPE_NUM:
			loopCondition = fmt.Sprintf("float64(%v) <= %v", iteratorVar, loopNumber)
			loopNumber = fmt.Sprintf("int(%v)", loopNumber)
		case TYPE_STR, TYPE_ARR, TYPE_OBJ:
			loopCondition = fmt.Sprintf("%v <= OSLround(%v)", iteratorVar, loopNumber)
			loopNumber = fmt.Sprintf("OSLround(%v)", loopNumber)
		default:
			loopCondition = fmt.Sprintf("%v <= %v", iteratorVar, loopNumber)
		}

		out += fmt.Sprintf("for %v := 1; %v; %v++ {\n", iteratorVar, loopCondition, iteratorVar)
		out += CompileBlock(blockData, ctx)
		ctx.Indent--
		out += AddIndent("}", ctx.Indent*2)
	case "for":
		if len(cmd) < 3 {
			panic("For command requires at least 2 parameters")
		}
		var iteratorVar string
		if len(cmd) > 1 {
			iteratorVar = cmd[1].Data.(string)
		}
		loopNumber := CompileToken(cmd[2], ctx)
		blk := cmd[3]
		var blockData [][]*Token
		if blkData, ok := blk.Data.([][]*Token); ok {
			blockData = blkData
		} else if dataStr, ok := blk.Data.(string); ok {
			blockData = scriptToAst(dataStr)
		} else if singleToken, ok := blk.Data.([]*Token); ok {
			blockData = append(blockData, singleToken)
		}
		ctx.Indent++
		returnType := cmd[2].ReturnedType
		if returnType == TYPE_NUM {
			loopNumber = fmt.Sprintf("int(%v)", loopNumber)
		} else if returnType != TYPE_INT {
			loopNumber = fmt.Sprintf("OSLround(%v)", loopNumber)
		}
		out += fmt.Sprintf("for %v := 1; %v <= %v; %v++ {\n", iteratorVar, iteratorVar, loopNumber, iteratorVar)
		out += CompileBlock(blockData, ctx)
		ctx.Indent--
		out += AddIndent("}", ctx.Indent*2)
	case "while":
		if len(cmd) < 3 {
			panic("While command requires at least 2 parameters")
		}
		blk := cmd[2]
		if blk.Type != TKN_BLK {
			panic("While command requires a block")
		}
		var condition string
		if len(cmd) > 1 {
			condition = CompileToken(cmd[1], ctx)
		}
		out += fmt.Sprintf("for %v {\n", condition)
		ctx.Indent++
		out += CompileBlock(blk.Data.([][]*Token), ctx)
		ctx.Indent--
		out += AddIndent("}", ctx.Indent*2)
	case "log", "say":
		if len(cmd) < 2 {
			panic("Log command requires at least 1 parameter")
		}
		out = "OSLlogValues("
		for i, param := range cmd {
			if i == 0 {
				continue
			}
			out += CompileToken(param, ctx)
			if i < len(cmd)-1 {
				out += ", "
			}
		}
		out += ")"
	case "window":
		params := []string{}
		for i := 1; i < len(cmd); i++ {
			params = append(params, CompileToken(cmd[i], ctx))
		}
		switch cmd[1].Data {
		case "dimensions":
			out += AddIndent("window.resize("+params[1]+", "+params[2]+")\n", ctx.Indent*2)
		case "resizable":
			out += AddIndent("window.setResizable("+params[1]+")\n", ctx.Indent*2)
		default:
			out += "// window " + cmd[1].Data.(string) + " " + strings.Join(params, ", ") + "\n"
		}
	case "type":
		name := cmd[1].Data.(string)
		oslTypes[name] = "*OSL_" + name
		switch cmd[2].Type {
		case TKN_VAR:
			typeStr := cmd[2].Data.(string)
			out += "type OSL_" + name + " " + mapOSLTypeToGo(typeStr) + "\n"
		case TKN_BLK:
			defaults := make(map[string]*Token)
			inlines := make(map[string]*Token)
			out += "type OSL_" + cmd[1].Data.(string) + " struct {\n"
			selfTypes := make(map[string]string)
			ctx.Indent++
			for _, line := range cmd[2].Data.([][]*Token) {
				switch line[0].Type {
				case TKN_ASI:
					val := line[0]
					if val.Right != nil && val.Left != nil &&
						val.Right.Type == TKN_FNC && val.Left.Type == TKN_VAR &&
						val.Right.Data == "function" && ctx.Indent == 0 {
						ctx.Indent++
						ctx.ScopeLevel++

						savedDeclaredVars := ctx.DeclaredVars
						ctx.DeclaredVars = make(map[string]bool)

						params_string := ""
						if len(val.Right.Parameters) > 0 {
							params := val.Right.Parameters
							args := strings.Split(params[0].Data.(string), ",")
							for i, arg := range args {
								args[i] = strings.TrimSpace(arg)
								if args[i] == "" {
									continue
								}
								parts := strings.Split(args[i], " ")
								typeName := "any"
								varName := args[i]
								if len(parts) > 1 {
									typeName = mapOSLTypeToGo(parts[1])
									varName = parts[0]
								}

								ctx.DeclaredVars[varName] = true
								params_string += fmt.Sprintf("%v %v, ", varName, typeName)
							}
						}
						params_string = strings.TrimSuffix(params_string, ", ")
						returns := ""
						if val.Right.Returns != "" {
							returns = mapOSLTypeToGo(val.Right.Returns) + " "
						}

						var blockData [][]*Token = nil
						funcBody := ""
						if len(val.Right.Parameters) > 1 && val.Right.Parameters[1] != nil {
							if bd, ok := val.Right.Parameters[1].Data.([][]*Token); ok {
								blockData = bd
								funcBody = CompileBlock(blockData, ctx)
								if ctx.selfUsed {
									params_string = "OSLself any, " + params_string
									ctx.selfUsed = false
								}
							}
						}

						hasReturn := hasReturnStatement(blockData)
						if returns == "" && hasReturn {
							returns = "any "
						}

						if hasReturn && !strings.Contains(funcBody, "return") {
							funcBody += AddIndent("return nil\n", ctx.Indent*2)
						}

						out += fmt.Sprintf("func (OSLself *OSL_%v) %v%v {\n", cmd[1].Data, params_string, returns)
						out += funcBody + "\n}\n"

						ctx.DeclaredVars = savedDeclaredVars
						ctx.Indent--
						ctx.ScopeLevel--
						break
					}
					val = line[0]
					varName := val.Left.Data.(string)
					if val.Right != nil && val.Right.Type == TKN_FNC && val.Right.Data == "function" {
						inlines[varName] = val.Right
						params := val.Right.Parameters[0].Data.(string)
						parts := strings.Split(params, ",")
						paramStr := make([]string, len(parts))
						for i, part := range parts {
							parts[i] = strings.TrimSpace(part)
							if parts[i] == "" {
								continue
							}
							paramStr[i] = mapOSLTypeToGo(strings.Split(parts[i], " ")[1])
						}
						val.SetType = "func(" + strings.Join(paramStr, ", ") + ") " + val.Right.Returns
					} else {
						defaults[varName] = val.Right
					}
					if val.SetType == "" {
						val.SetType = "any"
					}
					typeStr := mapOSLTypeToGo(val.SetType)
					selfTypes[varName] = typeStr
					out += AddIndent(fmt.Sprintf("%v %v\n", varName, typeStr), ctx.Indent*2)
				}
			}
			ctx.selfTypes = selfTypes
			ctx.Indent--
			out += "}\nfunc OSL_new_" + name + "() *OSL_" + name + " {\n"
			ctx.Indent++
			out += AddIndent("OSLself = &OSL_"+name+"{\n", ctx.Indent*2)
			ctx.Indent++
			for varName, val := range defaults {
				out += AddIndent(fmt.Sprintf("%v: %v,\n", varName, CompileToken(val, ctx)), ctx.Indent*2)
			}
			ctx.Indent--
			out += AddIndent("}\n", ctx.Indent*2)
			for varName, val := range inlines {
				out += AddIndent(fmt.Sprintf("OSLself.%v = %v\n", varName, CompileToken(val, ctx)), ctx.Indent*2)
			}
			out += AddIndent("return OSLself\n", ctx.Indent*2)
			ctx.Indent--
			out += "}\n"
		}
	case "return":
		if len(cmd) < 2 {
			out += "return"
			break
		}
		if len(cmd) > 1 {
			out += fmt.Sprintf("return %v", CompileToken(cmd[1], ctx))
		}
	case "wait":
		if len(cmd) == 2 {
			out += "OSLwait(" + CompileToken(cmd[1], ctx) + ")"
		}
	case "switch":
		if len(cmd) < 3 {
			panic("Switch command requires at least 2 parameters")
		}
		out += "switch " + CompileToken(cmd[1], ctx) + " {\n"
		ctx.Indent++
		out += CompileBlock(cmd[2].Data.([][]*Token), ctx)
		ctx.Indent--
		out += AddIndent("}\n", ctx.Indent*2)
	case "case":
		if len(cmd) < 2 {
			panic("Case command requires at least 1 parameter")
		}
		out += "case " + CompileToken(cmd[1], ctx) + ":\n"
	case "def":
		if len(cmd) < 2 {
			panic("Def command requires at least 1 parameter")
		}
		cmdName := cmd[1]

		funcName := ""
		if cmdName.Type == TKN_STR {
			// This is a custom command
			if name, ok := cmdName.Data.(string); ok {
				funcName = "OSLCMD_" + name
				ctx.CustomCommands[name] = true
			}
		} else if name, ok := cmdName.Data.(string); ok {
			// This is a regular function
			funcName = name
		}

		ctx.Indent++
		ctx.ScopeLevel++

		savedDeclaredVars := ctx.DeclaredVars
		ctx.DeclaredVars = make(map[string]bool)

		var paramString strings.Builder
		if len(cmd) > 2 {
			cmdParams := cmd[2]
			switch cmdParams.Type {
			case TKN_STR:
				paramString.WriteString(fmt.Sprintf("%v any", cmdParams.Data))
				if name, ok := cmdName.Data.(string); ok {
					ctx.DeclaredVars[name] = true
				}
			case TKN_MTV:
				fields := cmdParams.Data.([]*Token)
				for i, f := range fields {
					if i > 0 {
						paramString.WriteString(", ")
					}
					if s, ok := f.Data.(string); ok {
						parts := strings.SplitN(strings.TrimSpace(s), " ", 2)
						varName := s
						goType := "any"
						if len(parts) == 2 {
							oslType := parts[0]
							varName = parts[1]
							goType = mapOSLTypeToGo(oslType)
							ctx.DeclaredVars[varName] = true
							ctx.VariableTypes[varName] = goType
						} else {
							ctx.DeclaredVars[s] = true
						}
						paramString.WriteString(fmt.Sprintf("%v %v", varName, goType))
					}
				}
			}
		}

		var funcBody string
		var blockData [][]*Token = nil
		if len(cmd) > 3 {
			cmdBody := cmd[3]
			if cmdBody.Type == TKN_BLK {
				if blk, ok := cmdBody.Data.([][]*Token); ok {
					blockData = blk
					funcBody = CompileBlock(blk, ctx)
				}
			}
		} else if len(cmd) > 2 && cmd[2].Type == TKN_BLK {
			if blk, ok := cmd[2].Data.([][]*Token); ok {
				blockData = blk
				funcBody = CompileBlock(blk, ctx)
			}
		}

		hasReturn := hasReturnStatement(blockData)
		funcResult := ""
		if hasReturn {
			funcResult = "any"
		}

		if hasReturn && !strings.Contains(funcBody, "return") {
			funcBody += AddIndent("return nil\n", ctx.Indent*2)
		}

		out = "func " + funcName + "(" + paramString.String() + ") " + funcResult + " {\n" + funcBody + "}"

		ctx.DeclaredVars = savedDeclaredVars
		ctx.Indent--
		ctx.ScopeLevel--
	case "import":
		if len(cmd) < 2 {
			panic("Import command requires at least 1 parameter")
		}
		if len(cmd) > 1 {
			importPath := cmd[1].Data.(string)
			if !ctx.Imports[importPath] {
				ctx.Imports[importPath] = true
				ctx.ImportOrder = append(ctx.ImportOrder, importPath)
			}
		}
	case "go", "defer":
		if len(cmd) < 2 {
			panic("Go and defer commands require at least 1 parameter")
		}
		out += cmd[0].Data.(string) + " "
		for i := 1; i < len(cmd); i++ {
			out += CompileToken(cmd[i], ctx)
		}
	case "void":
		if len(cmd) < 2 {
			panic("Void command requires at least 1 parameter")
		}
		for i := 1; i < len(cmd); i++ {
			out += CompileToken(cmd[i], ctx)
		}
	case "c", "color", "colour":
		if len(cmd) < 2 {
			panic("Color command requires at least 1 parameter")
		}
		out += "OSLdrawctx.Color(" + CompileToken(cmd[1], ctx) + ")"
	case "goto":
		if len(cmd) != 3 {
			panic("Goto command requires 2 parameters")
		}
		x := CompileToken(cmd[1], ctx)
		y := CompileToken(cmd[2], ctx)
		out += "OSLdrawctx.Goto(" + x + ", " + y + ")"
	case "change_x":
		if len(cmd) != 2 {
			panic("Change_x command requires 1 parameter")
		}
		out += "OSLdrawctx.Change(" + CompileToken(cmd[1], ctx) + ", 0)"
	case "change_y":
		if len(cmd) != 2 {
			panic("Change_y command requires 1 parameter")
		}
		out += "OSLdrawctx.Change(0, " + CompileToken(cmd[1], ctx) + ")"
	case "change":
		if len(cmd) != 3 {
			panic("Change command requires 2 parameters")
		}
		out += "OSLdrawctx.Change(" + CompileToken(cmd[1], ctx) + ", " + CompileToken(cmd[2], ctx) + ")"
	case "loc":
		if len(cmd) != 5 {
			panic("Loc command requires 4 parameters")
		}
		a := CompileToken(cmd[1], ctx)
		b := CompileToken(cmd[2], ctx)
		c := CompileToken(cmd[3], ctx)
		d := CompileToken(cmd[4], ctx)
		out += "OSLdrawctx.Loc(" + a + ", " + b + ", " + c + ", " + d + ")"
	case "square":
		if len(cmd) < 3 {
			panic("Square command requires at least 2 parameters")
		}
		out += "OSLdrawctx.Rect("
		for i := 1; i < len(cmd); i++ {
			out += CompileToken(cmd[i], ctx) + ", "
		}
		out += ")"
	case "icon":
		if len(cmd) != 3 {
			panic("Icon command requires 2 parameters, got: " + cmd[0].Source)
		}
		out += "OSLdrawctx.Icon(" + CompileToken(cmd[1], ctx) + ", " + CompileToken(cmd[2], ctx) + ")"
	case "text":
		if len(cmd) < 3 {
			panic("Text command requires at least 2 parameters")
		}
		out += "OSLdrawctx.Text(" + CompileToken(cmd[1], ctx) + ", " + CompileToken(cmd[2], ctx) + ")"
	case "direction":
		if len(cmd) != 2 {
			panic("Direction command requires 1 parameter")
		}
		out += "OSLdrawctx.Direction(" + CompileToken(cmd[1], ctx) + ")"
	case "turnright":
		if len(cmd) != 2 {
			panic("Turnright command requires 1 parameter")
		}
		angle := CompileToken(cmd[1], ctx)
		out += "OSLdrawctx.Turnright(" + angle + ")"
	case "turnleft":
		if len(cmd) != 2 {
			panic("Turnleft command requires 1 parameter")
		}
		angle := CompileToken(cmd[1], ctx)
		out += "OSLdrawctx.Turnleft(" + angle + ")"
	case "pointat":
		if len(cmd) != 3 {
			panic("Pointat command requires 2 parameters")
		}
		out += "OSLdrawctx.Pointat(" + CompileToken(cmd[1], ctx) + ", " + CompileToken(cmd[2], ctx) + ")"
	default:
		cmdName := ""
		if name, ok := cmd[0].Data.(string); ok {
			cmdName = name
		}

		// Check if it's a custom command
		if ctx.CustomCommands[cmdName] {
			out += "OSLCMD_" + cmdName + "("
			for i := 1; i < len(cmd); i++ {
				out += CompileToken(cmd[i], ctx)
				if i < len(cmd)-1 {
					out += ", "
				}
			}
			out += ")"
		} else {
			// Unknown command
			cmdName := cmd[0].Data.(string)

			// Go keywords that should not have parentheses
			goKeywords := map[string]bool{
				"if": true, "else": true, "for": true, "while": true,
				"switch": true, "case": true, "default": true,
				"break": true, "continue": true, "return": true,
				"goto": true, "defer": true, "go": true,
			}

			if goKeywords[cmdName] {
				// Go statement - no parentheses
				out += cmdName
				if len(cmd) > 1 {
					out += " "
					for i := 1; i < len(cmd); i++ {
						out += CompileToken(cmd[i], ctx)
						if i < len(cmd)-1 {
							out += " "
						}
					}
				}
			} else {
				// Function call - with parentheses
				out += cmdName + "("
				if len(cmd) > 1 {
					for i := 1; i < len(cmd); i++ {
						out += CompileToken(cmd[i], ctx)
						if i < len(cmd)-1 {
							out += ", "
						}
					}
				}
				out += ")"
			}
		}
	}
	return out + "\n"
}

func CompileObject(obj [][]*Token, ctx *VariableContext) string {
	var out strings.Builder
	if ctx.Indent == 0 {
		out.WriteString("newSafeMap(map[string]any{\n")
	} else {
		out.WriteString("map[string]any{\n")
	}
	for _, token := range obj {
		keyStr := ""
		switch token[0].Type {
		case TKN_VAR:
			// In object literals, variable names should be treated as string keys
			keyStr = fmt.Sprintf("\"%v\"", token[0].Data)
		case TKN_STR:
			// String keys should be used as-is
			keyStr = CompileToken(token[0], ctx)
		default:
			// Other expressions should be evaluated and used as keys
			keyStr = CompileToken(token[0], ctx)
		}
		if len(token) > 1 {
			valueStr := CompileToken(token[1], ctx)
			out.WriteString(AddIndent(fmt.Sprintf("%v: %v,\n", keyStr, valueStr), ctx.Indent*2))
		}
	}
	if ctx.Indent == 0 {
		return out.String() + "})"
	}
	return out.String() + AddIndent("}", (ctx.Indent-1)*2)
}

func CompileArray(arr []*Token, ctx *VariableContext) string {
	if len(arr) == 0 {
		return "[]any{}"
	}
	var out strings.Builder
	out.WriteString("[]any{\n")
	for _, token := range arr {
		out.WriteString(AddIndent(fmt.Sprintf("%v,\n", CompileToken(token, ctx)), ctx.Indent*2))
	}
	return out.String() + AddIndent("}", (ctx.Indent-1)*2)
}

func AddIndent(str string, indent int) string {
	var out strings.Builder
	for range indent {
		out.WriteString(" ")
	}
	out.WriteString(str)
	return out.String()
}
