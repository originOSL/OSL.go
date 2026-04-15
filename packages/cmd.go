// name: cmd
// description: Command line utilities
// author: Mist
// requires: os, strings, os/exec

type Cmd struct{}

func (Cmd) Run(cmd string, args ...string) string {
	output, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		return ""
	}
	return string(output)
}

var cmd = Cmd{}