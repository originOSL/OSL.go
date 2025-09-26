package main

import (
	"embed"
	"fmt"
	"os"
	"strings"
)

//go:embed std/*.go
var typesFS embed.FS

//go:embed packages/*.go
var packagesFS embed.FS

type VariableContext struct {
	Globals      map[string]any
	Locals       map[string]any
	Indent       int
	Prepend      map[string]string
	Imports      map[string]bool
	DeclaredVars map[string]bool
}

func include(path string) string {
	data, err := typesFS.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(data) + "\n"
}

func mapOSLTypeToGo(oslType string) string {
	switch oslType {
	case "string":
		return "string"
	case "int":
		return "int"
	case "number":
		return "float64"
	case "boolean":
		return "bool"
	case "object":
		return "map[string]any"
	case "array":
		return "[]any"
	default:
		if strings.HasSuffix(oslType, "[]") {
			return "[]" + mapOSLTypeToGo(strings.TrimSuffix(oslType, "[]"))
		}
		return oslType
	}
}

func Compile(ast [][]*Token) string {
	ctx := VariableContext{
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
		},
		DeclaredVars: make(map[string]bool),
	}
	compiled := CompileBlock(ast, ctx)
	prepend := ""
	for _, v := range ctx.Prepend {
		prepend += fmt.Sprintf("%v", v)
	}

	imported := make(map[string]bool)
	goImports := []string{}

	for k, v := range ctx.Imports {
		if !v {
			continue
		}

		if imported[k] {
			continue
		}
		imported[k] = true

		if after, ok := strings.CutPrefix(k, "osl/"); ok {
			data, err := packagesFS.ReadFile("packages/" + after + ".go")
			if err != nil {
				panic(err)
			}
			file := string(data)
			lines := strings.Split(file, "\n")
			for _, line := range lines {
				if requires, ok := strings.CutPrefix(line, "// requires: "); ok {
					parts := strings.Split(requires, ", ")
					for _, part := range parts {
						if imported[part] {
							continue
						}
						part = strings.TrimSpace(part)
						imported[part] = true
						goImports = append(goImports, part)
					}
				}
			}
			compiled = "\n" + strings.TrimSpace(file) + compiled
			continue
		}

		if strings.HasPrefix(k, "./") && strings.HasSuffix(k, ".osl") {
			data, err := os.ReadFile(strings.TrimPrefix(k, "./"))
			if err != nil {
				panic(err)
			}
			compiled = "\n" + CompileBlock(scriptToAst(string(data)), ctx) + compiled
			continue
		}

		goImports = append(goImports, k)
	}

	if len(goImports) > 0 {
		prepend += "import (\n"
		for _, pkg := range goImports {
			prepend += fmt.Sprintf("\t%q\n", pkg)
		}
		prepend += ")\n\n"
	}

	prepend += include("std/general.go")

	return prepend + compiled
}

func CompileBlock(block [][]*Token, ctx VariableContext) string {
	var out string
	for _, line := range block {
		out += AddIndent(CompileLine(line, ctx), ctx.Indent*2)
	}
	return out
}

func CompileLine(line []*Token, ctx VariableContext) string {
	var out string
	if line[0].Type == TKN_CMD {
		out += CompileCmd(line, ctx)
		return out
	}
	for _, token := range line {
		out += CompileToken(token, ctx)
	}
	return out + "\n"
}

func CompileToken(token *Token, ctx VariableContext) string {
	if token == nil {
		return ""
	}
	switch token.Type {
	case TKN_ASI:
		if token.Right.Type == TKN_FNC && token.Left.Type == TKN_VAR {
			ctx.Indent++
			params_string := ""
			params_token := strings.TrimSpace(token.Right.Parameters[0].Data.(string))
			if len(params_token) > 0 {
				params := strings.Split(params_token, ",")
				for i, param := range params {
					params_string += param
					if i < len(params)-1 {
						params_string += ", "
					}
				}
			}
			returns := ""
			if token.Right.Returns != "" {
				returns = mapOSLTypeToGo(token.Right.Returns) + " "
			}
			out := "func " + token.Left.Data.(string) + "(" +
				params_string + ") " + returns + "{\n" +
				CompileBlock(token.Right.Parameters[1].Data.([][]*Token), ctx) + "}"
			ctx.Indent--
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
			varName = token.Left.Data.(string)
		}

		if varName != "" {
			declared := ctx.DeclaredVars[varName]
			if !declared {
				ctx.DeclaredVars[varName] = true
			}
			if token.SetType != "" {
				tokenType := token.SetType
				switch tokenType {
				case "string":
					compiledRight = fmt.Sprintf("string(%v)", compiledRight)
				case "int":
					compiledRight = fmt.Sprintf("int(%v)", compiledRight)
				case "number":
					compiledRight = fmt.Sprintf("float64(%v)", compiledRight)
				case "boolean":
					compiledRight = fmt.Sprintf("bool(%v)", compiledRight)
				case "array":
					if !strings.HasSuffix(compiledRight, "{}") {
						compiledRight = fmt.Sprintf("%v.([]any)", compiledRight)
					}
				}
				goType := mapOSLTypeToGo(tokenType)
				return fmt.Sprintf("var %v %v %v %v", compiledLeft, goType, op, compiledRight)
			}
			if !declared {
				return fmt.Sprintf("var %v %v %v", compiledLeft, op, compiledRight)
			}
		}

		return fmt.Sprintf("%v %v %v", compiledLeft, op, compiledRight)
	case TKN_OPR:
		compiledLeft := CompileToken(token.Left, ctx)
		compiledRight := CompileToken(token.Right, ctx)

		switch token.Data {
		case "??":
			return fmt.Sprintf("OSLnullishCoaless(%v, %v)", compiledLeft, compiledRight)
		case "+":
			return fmt.Sprintf("%v + %v", compiledLeft, compiledRight)
		case "++":
			return fmt.Sprintf("OSLjoin(%v, %v)", compiledLeft, compiledRight)
		case "-", "/", "%":
			return fmt.Sprintf("(%v %v %v)", compiledLeft, token.Data, compiledRight)
		case "*":
			return fmt.Sprintf("OSLmultiply(%v, %v)", compiledLeft, compiledRight)
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
		}
		return fmt.Sprintf("%v %v %v", compiledLeft, token.Data, compiledRight)
	case TKN_LOG:
		op := ""
		switch token.Data {
		case "and":
			op = "&&"
		case "or":
			op = "||"
		}
		return fmt.Sprintf("%v %v %v", CompileToken(token.Left, ctx), op, CompileToken(token.Right, ctx))
	case TKN_BIT:
		return fmt.Sprintf("%v %v %v", CompileToken(token.Left, ctx), token.Data, CompileToken(token.Right, ctx))
	case TKN_STR:
		return JsonStringify(token.Data)
	case TKN_NUM:
		return fmt.Sprintf("%v", token.Data)
	case TKN_VAR:
		varName := token.Data.(string)
		if strings.HasPrefix(varName, "OSL") {
			panic("Cannot use reserved variable name: " + varName)
		}
		switch varName {
		case "null":
			return "nil"
		case "timestamp":
			return "time.Now().UnixMilli()"
		}
		return varName
	case TKN_BLK:
		ctx.Indent++
		blk := CompileBlock(token.Data.([][]*Token), ctx)
		ctx.Indent--
		return fmt.Sprintf("(\n%v\n)", blk)
	case TKN_ARR:
		ctx.Indent++
		arr := CompileArray(token.Data.([]*Token), ctx)
		ctx.Indent--
		return arr
	case TKN_OBJ:
		ctx.Indent++
		if token.Data == nil {
			return "map[string]any{}"
		}
		obj := CompileObject(token.Data.([][]*Token), ctx)
		ctx.Indent--
		return AddIndent(obj, 1)
	case TKN_RMT:
		path := token.Data.([]*Token)
		out := ""
		out += CompileToken(path[0], ctx)

		if token.Final != nil {
			if token.Final.Type == TKN_MTV && token.Final.Data == "item" {
				out += fmt.Sprintf("[OSLcastString(%v)]", CompileToken(token.Final.Parameters[0], ctx))
			}
		}

		return out
	case TKN_FNC:
		params := token.Parameters
		if params == nil {
			params = []*Token{}
		}
		switch token.Data {
		case "function":
			ctx.Indent++
			paramString := ""
			if len(params) > 0 {
				paramString = params[0].Data.(string)
			}
			out := fmt.Sprintf("func(%v) {\n", paramString)
			ctx.Indent++
			blk := params[1]
			if blk != nil {
				out += CompileBlock(blk.Data.([][]*Token), ctx)
			}
			out += AddIndent("}", (ctx.Indent-2)*2)
			ctx.Indent--
			return out
		case "typeof":
			return fmt.Sprintf("OSLtypeof(%v)", CompileToken(token.Parameters[0], ctx))
		default:
			// params := []string{}
			paramString := ""
			if len(token.Parameters) > 0 {
				for i, p := range params {
					paramString += CompileToken(p, ctx)
					if i < len(token.Parameters)-1 {
						paramString += ", "
					}
				}
			}
			return fmt.Sprintf("%v(%v)", token.Data, paramString)
		}
	case TKN_URY:
		return fmt.Sprintf("%v(%v)", token.Data, CompileToken(token.Right, ctx))
	case TKN_MTD:
		out := ""
		parts := token.Data.([]*Token)
		first := parts[0]
		if first.Type == TKN_VAR {
			if strings.HasPrefix(first.Data.(string), "OSL") {
				panic("Cannot use reserved variable name: " + first.Data.(string))
			}
		}
		out = CompileToken(first, ctx)
		parts = parts[1:]
		for _, part := range parts {
			name := part.Data.(string)
			switch part.Type {
			case TKN_VAR:
				switch name {
				case "len":
					out = fmt.Sprintf("OSLlen(%v)", out)
				default:
					out = fmt.Sprintf("%v.%v", out, name)
				}
			case TKN_MTV:
				params := make([]string, len(part.Parameters))
				for i, p := range part.Parameters {
					params[i] = CompileToken(p, ctx)
				}
				switch name {
				case "toStr":
					out = fmt.Sprintf("OSLcastString(%v)", out)
				case "toInt":
					out = fmt.Sprintf("OSLcastInt(%v)", out)
				case "toNum":
					out = fmt.Sprintf("OSLcastFloat(%v)", out)
				case "toBool":
					out = fmt.Sprintf("OSLcastBool(%v)", out)
				case "item":
					out = fmt.Sprintf("OSLgetItem(%v, %v)", out, params[0])
				case "ask":
					out = fmt.Sprintf("input(%v)", out)
				case "chr":
					out = fmt.Sprintf("string(rune(OSLcastInt(%v)))", out)
				case "ord":
					out = fmt.Sprintf("int(OSLcastString(%v)[0])", out)
				case "toLower":
					out = fmt.Sprintf("strings.ToLower(%v)", out)
				case "toUpper":
					out = fmt.Sprintf("strings.ToUpper(%v)", out)
				case "startsWith":
					out = fmt.Sprintf("strings.HasPrefix(%v, %v)", out, params[0])
				case "endsWith":
					out = fmt.Sprintf("strings.HasSuffix(%v, %v)", out, params[0])
				case "contains":
					out = fmt.Sprintf("strings.Contains(%v, %v)", out, params[0])
				case "index":
					out = fmt.Sprintf("strings.Index(%v, %v)", out, params[0])
				case "strip":
					out = fmt.Sprintf("strings.TrimSpace(%v)", out)
				case "append":
					out = fmt.Sprintf("append(%v, %v)", out, params[0])
				case "prepend":
					out = fmt.Sprintf("append(%v, %v)", params[0], out)
				case "join":
					out = fmt.Sprintf("strings.Join(%v, %v)", out, params[0])
				case "split":
					out = fmt.Sprintf("strings.Split(%v, %v)", out, params[0])
				case "trim":
					out = fmt.Sprintf("OSLtrim(%v, %v, %v)", out, params[0], params[1])
				case "JsonStringify":
					out = fmt.Sprintf("JsonStringify(%v)", out)
				case "JsonParse":
					out = fmt.Sprintf("JsonParse(%v)", out)
				case "@": // assert type
					params[0] = strings.Trim(params[0], "\"")
					params[0] = mapOSLTypeToGo(params[0])
					out = fmt.Sprintf("%v.(%v)", out, params[0])
				default:
					if len(part.Parameters) == 0 {
						out = fmt.Sprintf("%v.%v()", out, name)
						break
					}
					out = fmt.Sprintf("%v.%v(%v)", out, name, strings.Join(params, ", "))
				}
			}
		}
		return out
	}
	return fmt.Sprintf("<%v>", token.Type)
}

func CompileCmd(cmd []*Token, ctx VariableContext) string {
	var out string
	switch cmd[0].Data {
	case "if":
		if len(cmd) < 3 {
			panic("If command requires at least 2 parameters")
		}
		blk := cmd[2]
		if blk.Type != TKN_BLK {
			panic("If command requires a block")
		}
		condition := CompileToken(cmd[1], ctx)

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
				condition := CompileToken(cmd[i+2], ctx)
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
				out += CompileBlock(cmd[i+1].Data.([][]*Token), ctx)
				ctx.Indent--
				out += AddIndent("}", ctx.Indent*2)
				i += 2
				break
			} else {
				break
			}
		}
	case "for":
		if len(cmd) < 3 {
			panic("For command requires at least 2 parameters")
		}
		iteratorVar := cmd[1].Data.(string)
		loopNumber := CompileToken(cmd[2], ctx)
		blk := cmd[3]
		ctx.Indent++
		out += fmt.Sprintf("for %v := 1; %v <= %v; %v++ {\n", iteratorVar, iteratorVar, loopNumber, iteratorVar)
		out += CompileBlock(blk.Data.([][]*Token), ctx)
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
		out += fmt.Sprintf("for %v {\n", CompileToken(cmd[1], ctx))
		ctx.Indent++
		out += CompileBlock(blk.Data.([][]*Token), ctx)
		ctx.Indent--
		out += AddIndent("}", ctx.Indent*2)
	case "log":
		if len(cmd) < 2 {
			panic("Log command requires at least 1 parameter")
		}
		out = "fmt.Println("
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
	case "return":
		if len(cmd) < 2 {
			out += "return"
			break
		}
		out += fmt.Sprintf("return %v", CompileToken(cmd[1], ctx))
	case "import":
		if len(cmd) < 2 {
			panic("Import command requires at least 1 parameter")
		}
		ctx.Imports[cmd[1].Data.(string)] = true
	case "void":
		for i := 1; i < len(cmd); i++ {
			out += CompileToken(cmd[i], ctx)
		}
	}
	return out + "\n"
}

func CompileObject(obj [][]*Token, ctx VariableContext) string {
	var out string = "map[string]any{\n"
	for _, token := range obj {
		keyStr := ""
		if token[0].Type == TKN_VAR {
			keyStr = fmt.Sprintf("\"%v\"", token[0].Data)
		} else {
			keyStr = CompileToken(token[0], ctx)
		}
		out += AddIndent(fmt.Sprintf("%v: %v,\n", keyStr, CompileToken(token[1], ctx)), ctx.Indent*2)
	}
	return out + AddIndent("}", (ctx.Indent-1)*2)
}

func CompileArray(arr []*Token, ctx VariableContext) string {
	if len(arr) == 0 {
		return "[]any{}"
	}
	var out string = "[]any{\n"
	for _, token := range arr {
		out += AddIndent(fmt.Sprintf("%v,\n", CompileToken(token, ctx)), ctx.Indent*2)
	}
	return out + AddIndent("}", (ctx.Indent-1)*2)
}

func AddIndent(str string, indent int) string {
	var out string
	for range indent {
		out += " "
	}
	out += str
	return out
}
