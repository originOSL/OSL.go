package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/browser"
)

const (
	OSL_VERSION  = "0.1.0"
	OSL_NAME     = "OSL.go"
	OSL_AUTHOR   = "Mistium"
	OSL_LICENSE  = "MIT"
	OSL_URL      = "https://github.com/Mistium/OSL.go"
	HELP_MESSAGE = `OSL (Origin Scripting Language) CLI v%v

Usage:
  osl <command> [options]

Commands:
  setup                 Setup OSL.go environment
  compile <file.osl>    Compile OSL file
  transpile <file.osl>  Transpile OSL file to Go and print to stdout
  run <file.osl>        Compile and run OSL file
  ast <file.osl>        Generate AST for OSL file
  uninstall             Uninstall OSL.go
  origin                Open Origin website (https://origin.mistium.com)
  help                  Show this help message
  version               Show version information

For more information, visit: https://origin.mistium.com`
)

func setup() {
	fmt.Println("Setting up OSL.go environment...")

	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		panic(err)
	}

	dest := "/usr/local/bin/osl"
	if runtime.GOOS == "windows" {
		fmt.Println("Please manually copy", exePath, "to a folder in your PATH.")
		return
	}

	cmd := exec.Command("sudo", "mv", exePath, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("Failed to move binary:", err)
		return
	}

	fmt.Println("Installed OSL.go to", dest)
}

func uninstall() {
	fmt.Println("Uninstalling OSL.go...")

	dest := "/usr/local/bin/osl"
	if runtime.GOOS == "windows" {
		fmt.Println("Please manually remove osl.exe from the folder you installed it to and remove it from PATH if necessary.")
		return
	}

	if _, err := os.Stat(dest); os.IsNotExist(err) {
		fmt.Println("No installation found at", dest)
		return
	}

	cmd := exec.Command("sudo", "rm", dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("Failed to remove binary:", err)
		return
	}

	fmt.Println("✅ Uninstalled OSL.go from", dest)
}

func open(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(data)
}

func scriptToAst(script string) [][]*Token {
	parser := NewOSLUtils()
	ast := parser.GenerateFullAST(script, true)
	return ast
}

func scriptToGo(script string) string {
	ast := scriptToAst(script)
	return Compile(ast)
}

func transpile(args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: osl transpile <file.osl>")
		return
	}

	scriptPath := args[0]
	script := open(scriptPath)
	fmt.Println(scriptToGo(script))
}

func compile(args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: osl compile <file.osl>")
		return
	}

	inputFile := args[0]
	script := open(inputFile)

	tmpDir, err := os.MkdirTemp("", "osl-compile-*")
	if err != nil {
		fmt.Println("Failed to create temp dir:", err)
		return
	}
	defer os.RemoveAll(tmpDir)

	tmpGoFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(tmpGoFile, []byte(scriptToGo(script)), 0644); err != nil {
		fmt.Println("Failed to write temp Go file:", err)
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Failed to get current working directory:", err)
		return
	}

	outputName := strings.TrimSuffix(filepath.Base(inputFile), ".osl")
	outputPath := filepath.Join(cwd, outputName)

	buildCmd := exec.Command("go", "build", "-o", outputPath, tmpGoFile)
	buildCmd.Dir = tmpDir
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		fmt.Println("Build failed!")
		fmt.Println(string(output))
		return
	}

	fmt.Printf("Compiled binary: %s\n", outputPath)
}

func run(args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: osl run <file.osl>")
		return
	}

	scriptPath := args[0]
	script := open(scriptPath)
	ast := scriptToAst(script)

	tmpDir, err := os.MkdirTemp("", "osl-build-*")
	if err != nil {
		fmt.Println("Failed to create temp dir:", err)
		return
	}
	defer os.RemoveAll(tmpDir)

	tmpGoFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(tmpGoFile, []byte(Compile(ast)), 0644); err != nil {
		fmt.Println("Failed to write temp Go file:", err)
		return
	}

	binaryPath := filepath.Join(tmpDir, "program")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, tmpGoFile)
	buildCmd.Dir = tmpDir
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		fmt.Println("Build failed!")
		fmt.Println(string(output))
		return
	}

	runCmd := exec.Command(binaryPath)
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr
	if err := runCmd.Run(); err != nil {
		fmt.Println("Error running binary:", err)
	}
}

func main() {
	args := os.Args
	if len(args) < 2 {
		response := HELP_MESSAGE
		fmt.Println(fmt.Sprintf(response, OSL_VERSION))
		return
	}

	switch args[1] {
	case "setup":
		setup()
	case "compile":
		compile(args[2:])
	case "transpile":
		transpile(args[2:])
	case "ast":
		os.WriteFile(args[2]+".ast.json", []byte(JsonStringify(scriptToAst(open(args[2])))), os.ModePerm)
	case "run":
		run(args[2:])
	case "uninstall":
		uninstall()
	case "origin":
		browser.OpenURL("https://origin.mistium.com")
	case "help":
		response := HELP_MESSAGE
		fmt.Println(fmt.Sprintf(response, OSL_VERSION))
	case "version":
		fmt.Println(OSL_VERSION)
	default:
		fmt.Println("Usage: osl <script> -o <output> -a (true/false generate ast)")
		return
	}
}
