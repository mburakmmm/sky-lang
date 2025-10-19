package skylib

import (
	"os"
	"os/exec"
	"runtime"
)

// GetEnv gets an environment variable
func OSGetEnv(key string) string {
	return os.Getenv(key)
}

// SetEnv sets an environment variable
func OSSetEnv(key, value string) error {
	return os.Setenv(key, value)
}

// Environ returns all environment variables
func OSEnviron() []string {
	return os.Environ()
}

// Getcwd returns current working directory
func OSGetcwd() (string, error) {
	return os.Getwd()
}

// Chdir changes current working directory
func OSChdir(path string) error {
	return os.Chdir(path)
}

// CPUCount returns number of CPUs
func OSCPUCount() int {
	return runtime.NumCPU()
}

// PID returns process ID
func OSPID() int {
	return os.Getpid()
}

// Platform returns OS name
func OSPlatform() string {
	return runtime.GOOS
}

// Arch returns architecture
func OSArch() string {
	return runtime.GOARCH
}

// Hostname returns hostname
func OSHostname() (string, error) {
	return os.Hostname()
}

// Exec executes a command and returns output
func OSExec(cmd string, args []string) (string, string, error) {
	command := exec.Command(cmd, args...)

	output, err := command.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", string(exitErr.Stderr), err
		}
		return "", "", err
	}

	return string(output), "", nil
}

// ExecInteractive executes a command with inherited stdin/stdout/stderr
func OSExecInteractive(cmd string, args []string) error {
	command := exec.Command(cmd, args...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}

// Exit exits the program
func OSExit(code int) {
	os.Exit(code)
}

// Args returns command line arguments
func OSArgs() []string {
	return os.Args
}

// Getuid returns user ID (Unix only)
func OSGetuid() int {
	return os.Getuid()
}

// Getgid returns group ID (Unix only)
func OSGetgid() int {
	return os.Getgid()
}

// TempDir returns temporary directory
func OSTempDir() string {
	return os.TempDir()
}

// ExpandEnv expands ${VAR} in string
func OSExpandEnv(s string) string {
	return os.ExpandEnv(s)
}

// Which finds executable in PATH
func OSWhich(name string) (string, error) {
	return exec.LookPath(name)
}
