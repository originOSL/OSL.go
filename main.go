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
	OSL_VERSION  = "0.2.0"
	OSL_NAME     = "OSL.go"
	OSL_AUTHOR   = "Mistium"
	OSL_LICENSE  = "MIT"
	OSL_URL      = "https://github.com/Mistium/OSL.go"
	HELP_MESSAGE = `OSL (Origin Scripting Language) CLI v%v

Usage:
  osl <command> [options]

Commands:
  setup                  Setup OSL.go environment
  compile <file.osl>     Compile OSL file
  compile-max <file.osl> Compile OSL file with maximum optimizations
  transpile <file.osl>   Transpile OSL file to Go and print to stdout
  run <file.osl>         Compile and run OSL file
  ast <file.osl>         Generate AST for OSL file
  uninstall              Uninstall OSL.go
  origin                 Open Origin website (https://origin.mistium.com)
  help                   Show this help message
  version                Show version information

For more information, visit: https://origin.mistium.com`
)

func setup() {
	fmt.Println("Setting up OSL.go environment...")

	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("Error getting executable path:", err)
		return
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		fmt.Println("Error resolving symlinks:", err)
		return
	}

	dest := "/usr/local/bin/osl"
	if runtime.GOOS == "windows" {
		fmt.Println("Please manually copy", exePath, "to a folder in your PATH.")
		return
	}

	data, err := os.ReadFile(exePath)
	if err != nil {
		fmt.Println("Failed to read binary:", err)
		return
	}

	if err := os.WriteFile(dest, data, 0755); err != nil {
		fmt.Println("Failed to install binary (try running with sudo):", err)
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

	fmt.Println("Uninstalled OSL.go from", dest)
}

func openFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Failed to read file:", err)
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
	return "package main\n\n" + Compile(ast)
}

func copyGoModFiles(srcDir, dstDir string) error {
	files := []string{"go.mod", "go.sum"}
	for _, f := range files {
		src := filepath.Join(srcDir, f)
		if _, err := os.Stat(src); err == nil {
			data, err := os.ReadFile(src)
			if err != nil {
				return fmt.Errorf("failed to read %s: %w", src, err)
			}
			dst := filepath.Join(dstDir, f)
			if err := os.WriteFile(dst, data, 0644); err != nil {
				return fmt.Errorf("failed to write %s: %w", dst, err)
			}
		}
	}
	return nil
}

func transpile(args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: osl transpile <file.osl>")
		return
	}

	scriptPath := args[0]

	scriptDir := filepath.Dir(scriptPath)
	originalDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Failed to get current directory:", err)
		return
	}

	if err := os.Chdir(scriptDir); err != nil {
		fmt.Println("Failed to change to script directory:", err)
		return
	}
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	script := openFile(filepath.Base(scriptPath))
	if script == "" {
		return
	}
	fmt.Println(scriptToGo(script))
}

func compile(main_args []string, max bool) {
	if len(main_args) != 1 {
		fmt.Println("Usage: osl compile <file.osl>")
		return
	}

	inputFile := main_args[0]

	scriptDir := filepath.Dir(inputFile)
	originalDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Failed to get current directory:", err)
		return
	}

	if err := os.Chdir(scriptDir); err != nil {
		fmt.Println("Failed to change to script directory:", err)
		return
	}
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	script := openFile(filepath.Base(inputFile))
	if script == "" {
		return
	}

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

	if err := copyGoModFiles(cwd, tmpDir); err != nil {
		fmt.Println("Warning: failed to copy go.mod/go.sum:", err)
	}

	outputName := strings.TrimSuffix(filepath.Base(inputFile), ".osl")
	outputPath := filepath.Join(cwd, outputName)

	args := []string{}
	if max {
		args = append(args, "-ldflags", "-s -w")
	}
	args = append(args, "-o", outputPath, tmpGoFile)

	buildCmd := exec.Command("go", append([]string{"build"}, args...)...)
	buildCmd.Dir = tmpDir
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		fmt.Println("Build failed!")
		fmt.Println(string(output))
		return
	}

	fmt.Printf("Compiled binary: %s\n", outputPath)
}

func ast(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: osl ast <file.osl>")
		return
	}
	script := openFile(args[0])
	if script == "" {
		return
	}
	ast := scriptToAst(script)
	jsonStr := JsonStringify(ast)
	outPath := args[0] + ".ast.json"
	if err := os.WriteFile(outPath, []byte(jsonStr), 0644); err != nil {
		fmt.Println("Failed to write AST file:", err)
		return
	}
	fmt.Println("Wrote AST to", outPath)
}

func run(args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: osl run <file.osl>")
		return
	}

	scriptPath := args[0]

	scriptDir := filepath.Dir(scriptPath)
	originalDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Failed to get current directory:", err)
		return
	}

	if err := os.Chdir(scriptDir); err != nil {
		fmt.Println("Failed to change to script directory:", err)
		return
	}
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	script := openFile(filepath.Base(scriptPath))
	if script == "" {
		return
	}

	tmpDir, err := os.MkdirTemp("", "osl-build-*")
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
	if err == nil {
		if err := copyGoModFiles(cwd, tmpDir); err != nil {
			// not fatal
			fmt.Println("Warning: failed to copy go.mod/go.sum:", err)
		}
	}

	// platform-specific binary name
	binName := "program"
	if runtime.GOOS == "windows" {
		binName = "program.exe"
	}
	binaryPath := filepath.Join(tmpDir, binName)

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
		compile(args[2:], false)
	case "compile-max":
		compile(args[2:], true)
	case "transpile":
		transpile(args[2:])
	case "ast":
		ast(args[2:])
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
