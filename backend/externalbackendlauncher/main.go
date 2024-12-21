package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"git.terah.dev/imterah/hermes/backendlauncher"
	"git.terah.dev/imterah/hermes/commonbackend"
	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v2"
)

type ProxyInstance struct {
	SourceIP   string `json:"sourceIP"`
	SourcePort uint16 `json:"sourcePort"`
	DestPort   uint16 `json:"destPort"`
	Protocol   string `json:"protocol"`
}

type WriteLogger struct {
	UseError bool
}

// TODO: deprecate UseError switching
func (writer WriteLogger) Write(p []byte) (n int, err error) {
	logSplit := strings.Split(string(p), "\n")

	for _, line := range logSplit {
		if line == "" {
			continue
		}

		if writer.UseError {
			log.Errorf("application: %s", line)
		} else {
			log.Infof("application: %s", line)
		}
	}

	return len(p), err
}

var (
	tempDir  string
	logLevel string
)

func entrypoint(cCtx *cli.Context) error {
	executablePath := cCtx.Args().Get(0)

	if executablePath == "" {
		return fmt.Errorf("executable file is not set")
	}

	executableParamsPath := cCtx.String("params-path")

	if executablePath == "" {
		return fmt.Errorf("executable parameters is not set")
	}

	proxyFilePath := cCtx.String("proxies")
	proxies := []ProxyInstance{}

	if proxyFilePath != "" {
		proxyFile, err := os.ReadFile(proxyFilePath)

		if err != nil {
			return fmt.Errorf("failed to read proxy file: %s", err.Error())
		}

		err = json.Unmarshal(proxyFile, &proxies)

		if err != nil {
			return fmt.Errorf("failed to parse proxy file: %s", err.Error())
		}
	}

	log.Debugf("discovered %d proxies.", len(proxies))

	backendParameters, err := os.ReadFile(executableParamsPath)

	if err != nil {
		return fmt.Errorf("could not read backend parameters: %s", err.Error())
	}

	_, err = os.Stat(executablePath)

	if err != nil {
		return fmt.Errorf("failed to get backend executable information: %s", err.Error())
	}

	log.Debug("running socket acquisition")

	sockPath, sockListener, err := backendlauncher.GetUnixSocket(tempDir)

	if err != nil {
		return fmt.Errorf("failed to acquire unix socket: %s", err.Error())
	}

	log.Debugf("acquisition was successful: %s", sockPath)

	go func() {
		log.Debug("entering execution loop (in auxiliary goroutine)...")

		for {
			log.Info("waiting for Unix socket connections...")
			sock, err := sockListener.Accept()
			log.Info("recieved connection. initializing...")

			if err != nil {
				log.Warnf("failed to accept socket connection: %s", err.Error())
			}

			defer sock.Close()

			startCommand := &commonbackend.Start{
				Type:      "start",
				Arguments: backendParameters,
			}

			startMarshalledCommand, err := commonbackend.Marshal("start", startCommand)

			if err != nil {
				log.Errorf("failed to generate start command: %s", err.Error())
				continue
			}

			if _, err = sock.Write(startMarshalledCommand); err != nil {
				log.Errorf("failed to write to socket: %s", err.Error())
				continue
			}

			commandType, commandRaw, err := commonbackend.Unmarshal(sock)

			if err != nil {
				log.Errorf("failed to read from/unmarshal from socket: %s", err.Error())
				continue
			}

			if commandType != "backendStatusResponse" {
				log.Errorf("recieved commandType '%s', expecting 'backendStatusResponse'", commandType)
				continue
			}

			command, ok := commandRaw.(*commonbackend.BackendStatusResponse)

			if !ok {
				log.Error("failed to typecast response")
				continue
			}

			if !command.IsRunning {
				var status string

				if command.StatusCode == commonbackend.StatusSuccess {
					status = "Success"
				} else {
					status = "Failure"
				}

				log.Errorf("failed to start backend (status: %s): %s", status, command.Message)
				continue
			}

			log.Info("successfully started backend.")

			hasAnyFailed := false

			for _, proxy := range proxies {
				log.Infof("initializing proxy %s:%d -> remote:%d", proxy.SourceIP, proxy.SourcePort, proxy.DestPort)

				proxyAddCommand := &commonbackend.AddProxy{
					Type:       "addProxy",
					SourceIP:   proxy.SourceIP,
					SourcePort: proxy.SourcePort,
					DestPort:   proxy.DestPort,
					Protocol:   proxy.Protocol,
				}

				marshalledProxyCommand, err := commonbackend.Marshal("addProxy", proxyAddCommand)

				if err != nil {
					log.Errorf("failed to generate start command: %s", err.Error())
					hasAnyFailed = true
					continue
				}

				if _, err = sock.Write(marshalledProxyCommand); err != nil {
					log.Errorf("failed to write to socket: %s", err.Error())
					hasAnyFailed = true
					continue
				}

				commandType, commandRaw, err := commonbackend.Unmarshal(sock)

				if err != nil {
					log.Errorf("failed to read from/unmarshal from socket: %s", err.Error())
					hasAnyFailed = true
					continue
				}

				if commandType != "proxyStatusResponse" {
					log.Errorf("recieved commandType '%s', expecting 'proxyStatusResponse'", commandType)
					hasAnyFailed = true
					continue
				}

				command, ok := commandRaw.(*commonbackend.ProxyStatusResponse)

				if !ok {
					log.Error("failed to typecast response")
					hasAnyFailed = true
					continue
				}

				if !command.IsActive {
					log.Error("failed to activate: isActive is false in response to AddProxy{} call")
					hasAnyFailed = true
					continue
				}

				log.Infof("successfully initialized proxy %s:%d -> remote:%d", proxy.SourceIP, proxy.SourcePort, proxy.DestPort)
			}

			if hasAnyFailed {
				log.Error("failed to initialize all proxies (read logs above)")
			} else {
				log.Info("successfully initialized all proxies")
			}

			log.Debug("entering infinite keepalive loop...")

			for {
			}
		}
	}()

	log.Debug("entering execution loop (in main goroutine)...")

	stdout := WriteLogger{
		UseError: false,
	}

	stderr := WriteLogger{
		UseError: true,
	}

	for {
		log.Info("starting process...")
		// TODO: can we reuse cmd?

		cmd := exec.Command(executablePath)
		cmd.Env = append(cmd.Env, fmt.Sprintf("HERMES_API_SOCK=%s", sockPath), fmt.Sprintf("HERMES_LOG_LEVEL=%s", logLevel))

		cmd.Stdout = stdout
		cmd.Stderr = stderr

		err := cmd.Run()

		if err != nil {
			if err, ok := err.(*exec.ExitError); ok {
				log.Warnf("backend died with exit code '%d' and with error '%s'", err.ExitCode(), err.Error())
			} else {
				log.Warnf("backend died with error: %s", err.Error())
			}
		} else {
			log.Info("process exited gracefully.")
		}

		log.Info("sleeping 5 seconds, and then restarting process")
		time.Sleep(5 * time.Millisecond)
	}
}

func main() {
	logLevel = os.Getenv("HERMES_LOG_LEVEL")

	if logLevel == "" {
		logLevel = "fatal"
	}

	switch logLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)

	case "info":
		log.SetLevel(log.InfoLevel)

	case "warn":
		log.SetLevel(log.WarnLevel)

	case "error":
		log.SetLevel(log.ErrorLevel)

	case "fatal":
		log.SetLevel(log.FatalLevel)
	}

	var err error
	tempDir, err = os.MkdirTemp("", "hermes-sockets-")

	if err != nil {
		log.Fatalf("failed to create sockets directory: %s", err.Error())
	}

	app := &cli.App{
		Name:   "externalbackendlauncher",
		Usage:  "for development purposes only -- external backend launcher for Hermes",
		Action: entrypoint,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "params-path",
				Aliases:  []string{"params", "pp"},
				Usage:    "file containing the parameters that are sent to the backend",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "proxies",
				Aliases: []string{"p"},
				Usage:   "file that contains the list of proxies to setup in JSON format",
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
