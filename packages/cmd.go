// name: cmd
// description: Command line utilities
// author: Mist
// requires: os, strings

type Cmd struct{}

func (Cmd) Run(cmd string, args ...string) string {
	cmdArgs := append([]string{cmd}, args...)
	output, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		return ""
	}
	return string(output)
}

var cmd = Cmd{}