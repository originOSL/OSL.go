// name: process
// description: Process management utilities
// author: roturbot
// requires: os/exec, os, time, strconv

type Process struct{}

type OSLProcess struct {
	cmd  *exec.Cmd
	PID  int
	Name string
}

func (Process) spawn(command any, args ...any) *OSLProcess {
	cmdStr := OSLtoString(command)
	argsSlice := make([]string, len(args))

	for i, arg := range args {
		argsSlice[i] = OSLtoString(arg)
	}

	cmd := exec.Command(cmdStr, argsSlice...)

	return &OSLProcess{
		cmd:  cmd,
		Name: cmdStr,
	}
}

func (Process) spawnShell(command any) *OSLProcess {
	cmdStr := OSLtoString(command)
	cmd := exec.Command("sh", "-c", cmdStr)

	return &OSLProcess{
		cmd:  cmd,
		Name: "sh",
	}
}

func (p *OSLProcess) run() map[string]any {
	if p == nil || p.cmd == nil {
		return map[string]any{"success": false, "error": "process not initialized"}
	}

	output, err := p.cmd.CombinedOutput()

	success := err == nil
	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
	}

	return map[string]any{
		"success": success,
		"output":  string(output),
		"error":   errorMsg,
		"code":    p.cmd.ProcessState.ExitCode(),
	}
}

func (p *OSLProcess) start() bool {
	if p == nil || p.cmd == nil {
		return false
	}

	err := p.cmd.Start()
	if err != nil {
		return false
	}

	p.PID = p.cmd.Process.Pid
	return true
}

func (p *OSLProcess) wait() map[string]any {
	if p == nil || p.cmd == nil {
		return map[string]any{"success": false, "error": "process not initialized"}
	}

	err := p.cmd.Wait()

	output, _ := p.cmd.CombinedOutput()

	success := err == nil
	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
	}

	return map[string]any{
		"success": success,
		"output":  string(output),
		"error":   errorMsg,
		"code":    p.cmd.ProcessState.ExitCode(),
	}
}

func (p *OSLProcess) kill() bool {
	if p == nil || p.cmd == nil || p.cmd.Process == nil {
		return false
	}

	err := p.cmd.Process.Kill()
	return err == nil
}

func (p *OSLProcess) signal(sig any) bool {
	if p == nil || p.cmd == nil || p.cmd.Process == nil {
		return false
	}

	sigStr := strings.ToUpper(OSLtoString(sig))

	var signal os.Signal
	switch sigStr {
	case "INT", "SIGINT":
		signal = os.Interrupt
	case "TERM", "SIGTERM":
		signal = syscall.SIGTERM
	case "KILL", "SIGKILL":
		signal = syscall.SIGKILL
	case "HUP", "SIGHUP":
		signal = syscall.SIGHUP
	default:
		return false
	}

	err := p.cmd.Process.Signal(signal)
	return err == nil
}

func (p *OSLProcess) isRunning() bool {
	if p == nil || p.cmd == nil || p.cmd.Process == nil {
		return false
	}

	err := p.cmd.Process.Signal(syscall.Signal(0))
	return err == nil
}

func (p *OSLProcess) getPID() int {
	if p == nil || p.cmd == nil || p.cmd.Process == nil {
		return 0
	}

	return p.cmd.Process.Pid
}

func (p *OSLProcess) setTimeout(timeout any) bool {
	if p == nil || p.cmd == nil {
		return false
	}

	duration := time.Duration(OSLcastNumber(timeout)) * time.Second
	p.cmd.SysProcAttr = &syscall.SysProcAttr{}
	return true
}

func (Process) getPID() int {
	return os.Getpid()
}

func (Process) getPPID() int {
	return os.Getppid()
}

func (Process) killPID(pid any) bool {
	pidInt := int(OSLcastNumber(pid))

	process, err := os.FindProcess(pidInt)
	if err != nil {
		return false
	}

	err = process.Kill()
	return err == nil
}

func (Process) signalPID(pid any, sig any) bool {
	pidInt := int(OSLcastNumber(pid))
	sigStr := strings.ToUpper(OSLtoString(sig))

	process, err := os.FindProcess(pidInt)
	if err != nil {
		return false
	}

	var signal os.Signal
	switch sigStr {
	case "INT", "SIGINT":
		signal = os.Interrupt
	case "TERM", "SIGTERM":
		signal = syscall.SIGTERM
	case "KILL", "SIGKILL":
		signal = syscall.SIGKILL
	case "HUP", "SIGHUP":
		signal = syscall.SIGHUP
	default:
		return false
	}

	err = process.Signal(signal)
	return err == nil
}

func (Process) isPIDRunning(pid any) bool {
	pidInt := int(OSLcastNumber(pid))

	process, err := os.FindProcess(pidInt)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func (Process) list() []map[string]any {
	cmd := exec.Command("ps", "aux")
	output, _ := cmd.CombinedOutput()

	lines := strings.Split(string(output), "\n")
	result := make([]map[string]any, 0, len(lines))

	for i, line := range lines {
		if i == 0 || line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 11 {
			user := fields[0]
			pidStr := fields[1]
			pid, _ := strconv.Atoi(pidStr)
			cpu, _ := strconv.ParseFloat(fields[2], 64)
			mem, _ := strconv.ParseFloat(fields[3], 64)
			vsz := fields[4]
			rss := fields[5]
			stat := fields[7]
			start := fields[8]
			time := fields[9]
			command := strings.Join(fields[10:], " ")

			result = append(result, map[string]any{
				"user":     user,
				"pid":      pid,
				"cpu":      cpu,
				"mem":      mem,
				"vsz":      vsz,
				"rss":      rss,
				"stat":     stat,
				"start":    start,
				"time":     time,
				"command":  command,
			})
		}
	}

	return result
}

func (Process) findByPID(pid any) map[string]any {
	pidInt := int(OSLcastNumber(pid))

	processes := process.list()
	for _, proc := range processes {
		if proc["pid"] == pidInt {
			return proc
		}
	}

	return nil
}

func (Process) findByName(name any) []map[string]any {
	nameStr := strings.ToLower(OSLtoString(name))

	processes := process.list()
	result := make([]map[string]any, 0)

	for _, proc := range processes {
		if strings.Contains(strings.ToLower(OSLtoString(proc["command"])), nameStr) {
			result = append(result, proc)
		}
	}

	return result
}

func (Process) killByName(name any) int {
	killed := 0
	processes := process.findByName(name)

	for _, proc := range processes {
		if process.killByPID(proc["pid"]) {
			killed++
		}
	}

	return killed
}

func (Process) environment() map[string]string {
	return env.getAll()
}

func (Process) setEnvironment(key any, value any) bool {
	return env.set(key, value)
}

func (Process) getEnvironment(key any) string {
	return env.get(key)
}

func (Process) workingDir() string {
	return sys.GetCwd()
}

func (Process) setWorkingDir(path any) bool {
	return sys.Chdir(path)
}

func (Process) getArguments() []any {
	return sys.GetArgs()
}

func (Process) getArg(index any) string {
	indexInt := OSLcastInt(index) - 1
	args := sys.GetArgs()

	if indexInt < 0 || indexInt >= len(args) {
		return ""
	}

	return OSLtoString(args[indexInt])
}

func (Process) getExecutablePath() string {
	return sys.GetExecutablePath()
}

func (Process) exec(command any, args ...any) string {
	return sys.CMD(OSLtoString(command), args...)
}

func (Process) execAsUser(user any, command any, args ...any) string {
	userStr := OSLtoString(user)
	cmdStr := OSLtoString(command)
	argsSlice := make([]string, len(args))

	for i, arg := range args {
		argsSlice[i] = OSLtoString(arg)
	}

	cmd := exec.Command("sudo", "-u", userStr, cmdStr)
	cmd.Args = append([]string{"sudo", "-u", userStr, cmdStr}, argsSlice...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "Error: " + err.Error()
	}

	return string(output)
}

func (Process) pipe(process1 *OSLProcess, process2 *OSLProcess) bool {
	if process1 == nil || process1.cmd == nil || process2 == nil || process2.cmd == nil {
		return false
	}

	stdout1, err := process1.cmd.StdoutPipe()
	if err != nil {
		return false
	}
	process1.cmd.Stdout = stdout1

	process2.cmd.Stdin = stdout1

	return true
}

func (Process) background(command any, args ...any) *OSLProcess {
	p := process.spawn(command, args...)
	
	if p.start() {
		return p
	}
	
	return nil
}

func (Process) daemonize(command any, args ...any) bool {
	p := process.spawn(command, args...)
	
	if p.start() {
		os.Exit(0)
		return true
	}
	
	return false
}

func (Process) fork() map[string]any {
	return map[string]any{"type": "error", "error": "fork not supported on this platform"}
}

func (Process) waitPID(pid any) map[string]any {
	pidInt := int(OSLcastNumber(pid))

	process, err := os.FindProcess(pidInt)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}

	state, err := process.Wait()
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}

	return map[string]any{
		"success": true,
		"pid":     pidInt,
		"code":    state.ExitCode(),
	}
}

func (Process) getMemoryMB() float64 {
	var memStat runtime.MemStats
	runtime.ReadMemStats(&memStat)

	return float64(memStat.Alloc) / 1024 / 1024
}

func (Process) getCPUTime() float64 {
	rusage := &syscall.Rusage{}
	syscall.Getrusage(syscall.RUSAGE_SELF, rusage)

	userTime := float64(rusage.Utime.Sec) + float64(rusage.Utime.Usec)/1000000.0
	systemTime := float64(rusage.Stime.Sec) + float64(rusage.Stime.Usec)/1000000.0

	return userTime + systemTime
}

func (Process) getNumGoroutines() int {
	return runtime.NumGoroutine()
}

func (Process) getNumCPU() int {
	return runtime.NumCPU()
}

func (Process) setNumCPU(n any) {
	nInt := int(OSLcastNumber(n))
	if nInt > 0 {
		runtime.GOMAXPROCS(nInt)
	}
}

func (Process) sleep(seconds any) float64 {
	sec := OSLcastNumber(seconds)
	if sec > 0 {
		time.Sleep(time.Duration(sec) * time.Second)
	}
	return sec
}

func (Process) exit(code any) {
	codeInt := int(OSLcastNumber(code))
	os.Exit(codeInt)
}

func (Process) getExitCode() int {
	return 0
}

var process = Process{}
