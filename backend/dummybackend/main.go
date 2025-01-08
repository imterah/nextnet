package main

import (
	"os"

	"git.terah.dev/imterah/hermes/backend/backendutil"
	"git.terah.dev/imterah/hermes/backend/commonbackend"
	"github.com/charmbracelet/log"
)

type DummyBackend struct {
}

func (backend *DummyBackend) StartBackend(byte []byte) (bool, error) {
	return true, nil
}

func (backend *DummyBackend) StopBackend() (bool, error) {
	return true, nil
}

func (backend *DummyBackend) GetBackendStatus() (bool, error) {
	return true, nil
}

func (backend *DummyBackend) StartProxy(command *commonbackend.AddProxy) (bool, error) {
	return true, nil
}

func (backend *DummyBackend) StopProxy(command *commonbackend.RemoveProxy) (bool, error) {
	return true, nil
}

func (backend *DummyBackend) GetAllClientConnections() []*commonbackend.ProxyClientConnection {
	return []*commonbackend.ProxyClientConnection{}
}

func (backend *DummyBackend) CheckParametersForConnections(clientParameters *commonbackend.CheckClientParameters) *commonbackend.CheckParametersResponse {
	// You don't have to specify Type and InReplyTo. Those will be handled for you.
	// Message is optional.
	return &commonbackend.CheckParametersResponse{
		IsValid: true,
		Message: "Valid!",
	}
}

func (backend *DummyBackend) CheckParametersForBackend(arguments []byte) *commonbackend.CheckParametersResponse {
	// You don't have to specify Type and InReplyTo. Those will be handled for you.
	// Message is optional.
	return &commonbackend.CheckParametersResponse{
		IsValid: true,
		Message: "Valid!",
	}
}

func main() {
	// When using logging, you should use charmbracelet/log, because that's what everything else uses in this ecosystem of a project. - imterah
	logLevel := os.Getenv("HERMES_LOG_LEVEL")

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

	backend := &DummyBackend{}

	application := backendutil.NewHelper(backend)
	err := application.Start()

	if err != nil {
		log.Fatalf("failed execution in application: %s", err.Error())
	}
}
