// name: sys
// description: OS utilities
// author: Mist
// requires: os/user, os, os/exec

type OS struct{}

func (OS) GetArgs() []string {
	return os.Args
}

func (OS) GetEnv(key string) string {
	return os.Getenv(key)
}

func (OS) SetEnv(key string, value string) bool {
	err := os.Setenv(key, value)
	return err == nil
}

func (OS) GetCwd() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}

func (OS) Chdir(path string) bool {
	err := os.Chdir(path)
	return err == nil
}

func (OS) GetPid() int {
	return os.Getpid()
}

func (OS) GetPpid() int {
	ppid := os.Getppid()
	if ppid == 0 {
		return -1
	}
	return ppid
}

func (OS) GetUid() int {
	uid := os.Getuid()
	if uid == 0 {
		return -1
	}
	return uid
}

func (OS) GetGid() int {
	gid := os.Getgid()
	if gid == 0 {
		return -1
	}
	return gid
}

func (OS) GetUsername() string {
	user, err := user.Current()
	if err != nil {
		return ""
	}
	return user.Username
}

func (OS) GetHomeDir() string {
	user, err := user.Current()
	if err != nil {
		return ""
	}
	return user.HomeDir
}

func (OS) CMD(cmd string, args ...string) string {
	cmdArgs := append([]string{cmd}, args...)
	out, err := exec.Command(cmdArgs[0], cmdArgs[1:]...).Output()
	if err != nil {
		return ""
	}
	return string(out)
}

var sys = OS{}