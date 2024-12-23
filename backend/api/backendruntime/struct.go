package backendruntime

import (
	"net"
	"os/exec"
	"strings"
)

type Backend struct {
	Name string `validate:"required"`
	Path string `validate:"required"`
}

type Runtime struct {
	isRuntimeRunning bool
	logger           *writeLogger
	currentProcess   *exec.Cmd
	currentListener  net.Listener

	ProcessPath     string
	Logs            []string
	RuntimeCommands chan interface{}
}

type writeLogger struct {
	Runtime *Runtime
}

func (writer writeLogger) Write(p []byte) (n int, err error) {
	logSplit := strings.Split(string(p), "\n")
	writer.Runtime.Logs = append(writer.Runtime.Logs, logSplit...)

	return len(p), err
}
