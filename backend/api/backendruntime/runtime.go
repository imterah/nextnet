package backendruntime

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"

	"git.terah.dev/imterah/hermes/backend/backendlauncher"
	"git.terah.dev/imterah/hermes/backend/commonbackend"
	"github.com/charmbracelet/log"
)

var (
	AvailableBackends []*Backend
	RunningBackends   map[uint]*Runtime
	TempDir           string
)

func init() {
	RunningBackends = make(map[uint]*Runtime)
}

func handleCommand(commandType string, command interface{}, sock net.Conn, rtcChan chan interface{}) {
	bytes, err := commonbackend.Marshal(commandType, command)

	if err != nil {
		log.Warnf("Failed to marshal message: %s", err.Error())
		rtcChan <- fmt.Errorf("failed to marshal message: %s", err.Error())

		return
	}

	if _, err := sock.Write(bytes); err != nil {
		log.Warnf("Failed to write message: %s", err.Error())
		rtcChan <- fmt.Errorf("failed to write message: %s", err.Error())

		return
	}

	_, data, err := commonbackend.Unmarshal(sock)

	if err != nil {
		log.Warnf("Failed to unmarshal message: %s", err.Error())
		rtcChan <- fmt.Errorf("failed to unmarshal message: %s", err.Error())

		return
	}

	rtcChan <- data
}

func (runtime *Runtime) goRoutineHandler() error {
	log.Debug("Starting up backend runtime")
	log.Debug("Running socket acquisition")

	logLevel := os.Getenv("HERMES_LOG_LEVEL")

	sockPath, sockListener, err := backendlauncher.GetUnixSocket(TempDir)

	if err != nil {
		return err
	}

	runtime.currentListener = sockListener

	log.Debugf("Acquired unix socket at: %s", sockPath)

	go func() {
		log.Debug("Creating new goroutine for socket connection handling")

		for {
			log.Debug("Waiting for Unix socket connections...")
			sock, err := runtime.currentListener.Accept()

			if err != nil {
				log.Warnf("Failed to accept Unix socket connection in a backend runtime instance: %s", err.Error())
				return
			}

			log.Debug("Recieved connection. Initializing...")

			defer sock.Close()

			for {
				commandRaw := <-runtime.RuntimeCommands

				log.Debug("Got message from server")

				switch command := commandRaw.(type) {
				case *commonbackend.AddProxy:
					handleCommand("addProxy", command, sock, runtime.RuntimeCommands)
				case *commonbackend.BackendStatusRequest:
					handleCommand("backendStatusRequest", command, sock, runtime.RuntimeCommands)
				case *commonbackend.BackendStatusResponse:
					handleCommand("backendStatusResponse", command, sock, runtime.RuntimeCommands)
				case *commonbackend.CheckClientParameters:
					handleCommand("checkClientParameters", command, sock, runtime.RuntimeCommands)
				case *commonbackend.CheckParametersResponse:
					handleCommand("checkParametersResponse", command, sock, runtime.RuntimeCommands)
				case *commonbackend.CheckServerParameters:
					handleCommand("checkServerParameters", command, sock, runtime.RuntimeCommands)
				case *commonbackend.ProxyClientConnection:
					handleCommand("proxyClientConnection", command, sock, runtime.RuntimeCommands)
				case *commonbackend.ProxyConnectionsRequest:
					handleCommand("proxyConnectionsRequest", command, sock, runtime.RuntimeCommands)
				case *commonbackend.ProxyConnectionsResponse:
					handleCommand("proxyConnectionsResponse", command, sock, runtime.RuntimeCommands)
				case *commonbackend.ProxyInstanceResponse:
					handleCommand("proxyInstanceResponse", command, sock, runtime.RuntimeCommands)
				case *commonbackend.ProxyInstanceRequest:
					handleCommand("proxyInstanceRequest", command, sock, runtime.RuntimeCommands)
				case *commonbackend.ProxyStatusRequest:
					handleCommand("proxyStatusRequest", command, sock, runtime.RuntimeCommands)
				case *commonbackend.ProxyStatusResponse:
					handleCommand("proxyStatusResponse", command, sock, runtime.RuntimeCommands)
				case *commonbackend.RemoveProxy:
					handleCommand("removeProxy", command, sock, runtime.RuntimeCommands)
				case *commonbackend.Start:
					handleCommand("start", command, sock, runtime.RuntimeCommands)
				case *commonbackend.Stop:
					handleCommand("stop", command, sock, runtime.RuntimeCommands)
				default:
					log.Warnf("Recieved unknown command type from channel: %q", command)
					runtime.RuntimeCommands <- fmt.Errorf("unknown command recieved")
				}
			}
		}
	}()

	for {
		log.Debug("Starting process...")

		ctx := context.Background()

		runtime.currentProcess = exec.CommandContext(ctx, runtime.ProcessPath)
		runtime.currentProcess.Env = append(runtime.currentProcess.Env, fmt.Sprintf("HERMES_API_SOCK=%s", sockPath), fmt.Sprintf("HERMES_LOG_LEVEL=%s", logLevel))

		runtime.currentProcess.Stdout = runtime.logger
		runtime.currentProcess.Stderr = runtime.logger

		err := runtime.currentProcess.Run()

		if err != nil {
			if err, ok := err.(*exec.ExitError); ok {
				if err.ExitCode() != -1 && err.ExitCode() != 0 {
					log.Warnf("A backend process died with exit code '%d' and with error '%s'", err.ExitCode(), err.Error())
				}
			} else {
				log.Warnf("A backend process died with error: %s", err.Error())
			}
		} else {
			log.Debug("Process exited gracefully.")
		}

		if !runtime.isRuntimeRunning {
			return nil
		}

		log.Debug("Sleeping 5 seconds, and then restarting process")
		time.Sleep(5 * time.Second)
	}
}

func (runtime *Runtime) Start() error {
	if runtime.isRuntimeRunning {
		return fmt.Errorf("runtime already running")
	}

	runtime.RuntimeCommands = make(chan interface{})

	runtime.logger = &writeLogger{
		Runtime: runtime,
	}

	go func() {
		err := runtime.goRoutineHandler()

		if err != nil {
			log.Errorf("Failed during execution of runtime: %s", err.Error())
		}
	}()

	runtime.isRuntimeRunning = true
	return nil
}

func (runtime *Runtime) Stop() error {
	if !runtime.isRuntimeRunning {
		return fmt.Errorf("runtime not running")
	}

	runtime.isRuntimeRunning = false

	if runtime.currentProcess != nil && runtime.currentProcess.Cancel != nil {
		err := runtime.currentProcess.Cancel()

		if err != nil {
			return fmt.Errorf("failed to stop process: %s", err.Error())
		}
	} else {
		log.Warn("Failed to kill process (Stop recieved), currentProcess or currentProcess.Cancel is nil")
	}

	if runtime.currentListener != nil {
		err := runtime.currentListener.Close()

		if err != nil {
			return fmt.Errorf("failed to stop listener: %s", err.Error())
		}
	} else {
		log.Warn("Failed to kill listener, as the listener is nil")
	}

	return nil
}

func NewBackend(path string) *Runtime {
	return &Runtime{
		ProcessPath: path,
	}
}

func Init(backends []*Backend) error {
	var err error
	TempDir, err = os.MkdirTemp("", "hermes-sockets-")

	if err != nil {
		return err
	}

	AvailableBackends = backends

	return nil
}
