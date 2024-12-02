package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"git.greysoh.dev/imterah/nextnet/backendlauncher"
	"github.com/charmbracelet/log"
)

type WriteLogger struct {
	UseError bool
}

func (writer WriteLogger) Write(p []byte) (n int, err error) {
	logSplit := strings.Split(string(p), "\n")

	for _, line := range logSplit {
		if writer.UseError {
			log.Errorf("application: %s", line)
		} else {
			log.Infof("application: %s", line)
		}
	}

	return len(p), err
}

func main() {
	tempDir, err := os.MkdirTemp("", "nextnet-sockets-")
	logLevel := os.Getenv("NEXTNET_LOG_LEVEL")

	if logLevel != "" {
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
	}

	if len(os.Args) != 3 {
		log.Fatalf("missing arguments! example: ./externalbackendlauncher <backend executable> <file with backend arguments>")
	}

	executablePath := os.Args[1]
	executableParamsPath := os.Args[2]

	_, err = os.ReadFile(executableParamsPath)

	if err != nil {
		log.Fatalf("could not read backend parameters: %s", err.Error())
	}

	_, err = os.Stat(executablePath)

	if err != nil {
		log.Fatalf("failed backend checks: %s", err.Error())
	}

	log.Debug("running socket acquisition")

	sockPath, sockListener, err := backendlauncher.GetUnixSocket(tempDir)

	if err != nil {
		log.Fatalf("failed to acquire unix socket: %s", err.Error())
	}

	log.Debugf("acquisition was successful: %s", sockPath)

	go func() {
		log.Debug("entering execution loop (in auxiliary goroutine)...")

		for {
			log.Info("waiting for Unix socket connections...")
			sock, err := sockListener.Accept()

			if err != nil {
				log.Warnf("failed to accept socket connection: %s", err.Error())
			}

			defer sock.Close()
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
		cmd.Env = append(cmd.Env, fmt.Sprintf("NEXTNET_API_SOCK=%s", sockPath))

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
