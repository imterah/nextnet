package backendruntime

import (
	"net"
	"os/exec"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
)

type Backend struct {
	Name string `validate:"required"`
	Path string `validate:"required"`
}

type messageForBuf struct {
	Channel chan interface{}
	Message interface{}
}

type Runtime struct {
	isRuntimeRunning           bool
	logger                     *writeLogger
	currentProcess             *exec.Cmd
	currentListener            net.Listener
	processRestartNotification chan bool

	messageBufferLock sync.Mutex
	messageBuffer     []*messageForBuf

	ProcessPath string
	Logs        []string

	OnCrashCallback func(sock net.Conn)
}

type writeLogger struct {
	Runtime *Runtime
}

func (writer writeLogger) Write(p []byte) (n int, err error) {
	logSplit := strings.Split(string(p), "\n")

	if isDevelopmentMode {
		for _, logLine := range logSplit {
			if logLine == "" {
				continue
			}

			log.Debug("spawned backend logs: " + logLine)
		}
	}

	writer.Runtime.Logs = append(writer.Runtime.Logs, logSplit...)

	return len(p), err
}
