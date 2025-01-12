package backendruntime

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"git.terah.dev/imterah/hermes/backend/backendlauncher"
	"git.terah.dev/imterah/hermes/backend/commonbackend"
	"github.com/charmbracelet/log"
)

func handleCommand(commandType string, command interface{}, sock net.Conn, rtcChan chan interface{}) error {
	bytes, err := commonbackend.Marshal(commandType, command)

	if err != nil {
		log.Warnf("Failed to marshal message: %s", err.Error())
		rtcChan <- fmt.Errorf("failed to marshal message: %s", err.Error())

		return fmt.Errorf("failed to marshal message: %s", err.Error())
	}

	if _, err := sock.Write(bytes); err != nil {
		log.Warnf("Failed to write message: %s", err.Error())
		rtcChan <- fmt.Errorf("failed to write message: %s", err.Error())

		return fmt.Errorf("failed to write message: %s", err.Error())
	}

	_, data, err := commonbackend.Unmarshal(sock)

	if err != nil {
		log.Warnf("Failed to unmarshal message: %s", err.Error())
		rtcChan <- fmt.Errorf("failed to unmarshal message: %s", err.Error())

		return fmt.Errorf("failed to unmarshal message: %s", err.Error())
	}

	rtcChan <- data

	return nil
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
		log.Debug("Created new Goroutine for socket connection handling")

		for {
			log.Debug("Waiting for Unix socket connections...")
			sock, err := runtime.currentListener.Accept()

			if err != nil {
				log.Warnf("Failed to accept Unix socket connection in a backend runtime instance: %s", err.Error())
				return
			}

			log.Debug("Recieved connection. Attempting to figure out backend state...")

			timeoutChannel := time.After(500 * time.Millisecond)

			select {
			case <-timeoutChannel:
				log.Debug("Timeout reached. Assuming backend is running.")
			case hasRestarted, ok := <-runtime.processRestartNotification:
				if !ok {
					log.Warnf("Failed to get the process restart notification state!")
				}

				if hasRestarted {
					if runtime.OnCrashCallback == nil {
						log.Warn("The backend has restarted for some reason, but we could not run the on crash callback as the callback is not set!")
					} else {
						log.Debug("We have restarted. Running the restart callback...")
						runtime.OnCrashCallback(sock)
					}

					log.Debug("Clearing caches...")
					runtime.cleanUpPendingCommandProcessingJobs()
					runtime.messageBufferLock = sync.Mutex{}
				} else {
					log.Debug("We have not restarted.")
				}
			}

			go func() {
				log.Debug("Setting up Hermes keepalive Goroutine")
				hasFailedBackendRunningCheckAlready := false

				for {
					if !runtime.isRuntimeRunning {
						return
					}

					// Asking for the backend status seems to be a "good-enough" keepalive system. Plus, it provides useful telemetry.
					// There isn't a ping command in the backend API, so we have to make do with what we have.
					//
					// To be safe here, we have to use the proper (yet annoying) facilities to prevent cross-talk, since we're in
					// a goroutine, and can't talk directly. This actually has benefits, as the OuterLoop should exit on its own, if we
					// encounter a critical error.
					statusResponse, err := runtime.ProcessCommand(&commonbackend.BackendStatusRequest{
						Type: "backendStatusRequest",
					})

					if err != nil {
						log.Warnf("Failed to get response for backend (in backend runtime keep alive): %s", err.Error())
						log.Debugf("Attempting to close socket...")
						err := sock.Close()

						if err != nil {
							log.Debugf("Failed to close socket: %s", err.Error())
						}

						continue
					}

					switch responseMessage := statusResponse.(type) {
					case *commonbackend.BackendStatusResponse:
						if !responseMessage.IsRunning {
							if hasFailedBackendRunningCheckAlready {
								if responseMessage.Message != "" {
									log.Warnf("Backend (in backend keepalive) is up but not active: %s", responseMessage.Message)
								} else {
									log.Warnf("Backend (in backend keepalive) is up but not active")
								}
							}

							hasFailedBackendRunningCheckAlready = true
						}
					default:
						log.Errorf("Got illegal response type for backend (in backend keepalive): %T", responseMessage)
						log.Debugf("Attempting to close socket...")
						err := sock.Close()

						if err != nil {
							log.Debugf("Failed to close socket: %s", err.Error())
						}
					}

					time.Sleep(5 * time.Second)
				}
			}()

		OuterLoop:
			for {
				for chanIndex, messageData := range runtime.messageBuffer {
					if messageData == nil {
						continue
					}

					switch command := messageData.Message.(type) {
					case *commonbackend.AddProxy:
						err := handleCommand("addProxy", command, sock, messageData.Channel)

						if err != nil {
							log.Warnf("failed to handle command in backend runtime instance: %s", err.Error())

							if strings.HasPrefix(err.Error(), "failed to write message") {
								break OuterLoop
							}
						}
					case *commonbackend.BackendStatusRequest:
						err := handleCommand("backendStatusRequest", command, sock, messageData.Channel)

						if err != nil {
							log.Warnf("failed to handle command in backend runtime instance: %s", err.Error())

							if strings.HasPrefix(err.Error(), "failed to write message") {
								break OuterLoop
							}
						}
					case *commonbackend.CheckClientParameters:
						err := handleCommand("checkClientParameters", command, sock, messageData.Channel)

						if err != nil {
							log.Warnf("failed to handle command in backend runtime instance: %s", err.Error())

							if strings.HasPrefix(err.Error(), "failed to write message") {
								break OuterLoop
							}
						}
					case *commonbackend.CheckServerParameters:
						err := handleCommand("checkServerParameters", command, sock, messageData.Channel)

						if err != nil {
							log.Warnf("failed to handle command in backend runtime instance: %s", err.Error())

							if strings.HasPrefix(err.Error(), "failed to write message") {
								break OuterLoop
							}
						}
					case *commonbackend.ProxyConnectionsRequest:
						err := handleCommand("proxyConnectionsRequest", command, sock, messageData.Channel)

						if err != nil {
							log.Warnf("failed to handle command in backend runtime instance: %s", err.Error())

							if strings.HasPrefix(err.Error(), "failed to write message") {
								break OuterLoop
							}
						}
					case *commonbackend.ProxyInstanceRequest:
						err := handleCommand("proxyInstanceRequest", command, sock, messageData.Channel)

						if err != nil {
							log.Warnf("failed to handle command in backend runtime instance: %s", err.Error())

							if strings.HasPrefix(err.Error(), "failed to write message") {
								break OuterLoop
							}
						}
					case *commonbackend.ProxyStatusRequest:
						command.Message()
						err := handleCommand("proxyStatusRequest", command, sock, messageData.Channel)

						if err != nil {
							log.Warnf("failed to handle command in backend runtime instance: %s", err.Error())

							if strings.HasPrefix(err.Error(), "failed to write message") {
								break OuterLoop
							}
						}
					case *commonbackend.RemoveProxy:
						err := handleCommand("removeProxy", command, sock, messageData.Channel)

						if err != nil {
							log.Warnf("failed to handle command in backend runtime instance: %s", err.Error())

							if strings.HasPrefix(err.Error(), "failed to write message") {
								break OuterLoop
							}
						}
					case *commonbackend.Start:
						err := handleCommand("start", command, sock, messageData.Channel)

						if err != nil {
							log.Warnf("failed to handle command in backend runtime instance: %s", err.Error())

							if strings.HasPrefix(err.Error(), "failed to write message") {
								break OuterLoop
							}
						}
					case *commonbackend.Stop:
						err := handleCommand("stop", command, sock, messageData.Channel)

						if err != nil {
							log.Warnf("failed to handle command in backend runtime instance: %s", err.Error())

							if strings.HasPrefix(err.Error(), "failed to write message") {
								break OuterLoop
							}
						}
					default:
						log.Warnf("Recieved unknown command type from channel: %T", command)
						messageData.Channel <- fmt.Errorf("unknown command recieved")
					}

					runtime.messageBuffer[chanIndex] = nil
				}
			}

			sock.Close()
		}
	}()

	runtime.processRestartNotification <- false

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

		// NOTE(imterah): This could cause hangs if we're not careful. If the process dies so much that we can't keep up, it should deserve to be hung, really.
		// There's probably a better way to do this, but this works.
		//
		// If this does turn out to be a problem, just increase the Goroutine buffer size.
		runtime.processRestartNotification <- true

		log.Debug("Sent off notification.")
	}
}

func (runtime *Runtime) Start() error {
	if runtime.isRuntimeRunning {
		return fmt.Errorf("runtime already running")
	}

	runtime.messageBuffer = make([]*messageForBuf, 10)
	runtime.messageBufferLock = sync.Mutex{}

	runtime.processRestartNotification = make(chan bool, 1)

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

func (runtime *Runtime) ProcessCommand(command interface{}) (interface{}, error) {
	schedulingAttempts := 0
	var commandChannel chan interface{}

SchedulingLoop:
	for {
		if !runtime.isRuntimeRunning {
			time.Sleep(10 * time.Millisecond)
		}

		if schedulingAttempts > 50 {
			return nil, fmt.Errorf("failed to schedule message transmission after 50 tries (REPORT THIS ISSUE)")
		}

		runtime.messageBufferLock.Lock()

		// Attempt to find spot in buffer to schedule message transmission
		for i, message := range runtime.messageBuffer {
			if message != nil {
				continue
			}

			commandChannel = make(chan interface{})

			runtime.messageBuffer[i] = &messageForBuf{
				Channel: commandChannel,
				Message: command,
			}

			runtime.messageBufferLock.Unlock()
			break SchedulingLoop
		}

		runtime.messageBufferLock.Unlock()
		time.Sleep(100 * time.Millisecond)

		schedulingAttempts++
	}

	// Fetch response and close Channel
	response, ok := <-commandChannel

	if !ok {
		return nil, fmt.Errorf("failed to read from command channel: recieved signal that is not OK")
	}

	close(commandChannel)

	err, ok := response.(error)

	if ok {
		return nil, err
	}

	return response, nil
}

func (runtime *Runtime) cleanUpPendingCommandProcessingJobs() {
	for messageIndex, message := range runtime.messageBuffer {
		if message == nil {
			continue
		}

		timeoutChannel := time.After(100 * time.Millisecond)

		select {
		case <-timeoutChannel:
			log.Warn("Message channel is likely running (timed out reading from it without an error)")
			close(message.Channel)
		case _, ok := <-message.Channel:
			if ok {
				log.Warn("Message channel is running, but should be stopped (since message is NOT nil!)")
				close(message.Channel)
			}
		}

		runtime.messageBuffer[messageIndex] = nil
	}
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
