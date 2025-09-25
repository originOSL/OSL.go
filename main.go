package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func open(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(data)
}

func main() {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)

	parser := NewOSLUtils()

	script := open("./examples/brainfuck-api.osl")

	outputFile := "./build/built.go"
	outputAst := "./build/ast.json"
	ast := parser.GenerateFullAST(script, true)
	os.Mkdir("./build", os.ModeDir)
	os.WriteFile(outputAst, []byte(JsonStringify(ast)), os.ModePerm)
	os.WriteFile(outputFile, []byte(Compile(ast)), os.ModePerm)
	// Change into build directory
	err := os.Chdir("./build")
	if err != nil {
		fmt.Println("Failed to change directory:", err)
		return
	}

	buildCmd := exec.Command("go", "build", "-o", "built", filepath.Base(outputFile))
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		fmt.Println("Build failed!")
		fmt.Println(string(output))
		return
	}

	runCmd := exec.Command("./built")
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr
	if err := runCmd.Run(); err != nil {
		fmt.Println("Error running binary:", err)
	}
}
