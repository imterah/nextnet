package backendutil

import "git.greysoh.dev/imterah/nextnet/commonbackend"

type BackendInterface interface {
	StartBackend() (bool, error)
	StopBackend() (bool, error)
	AddConnection(command *commonbackend.AddConnectionCommand) (bool, error)
	RemoveConnection(command *commonbackend.RemoveConnectionCommand) (bool, error)
	GetAllConnections() []*commonbackend.ClientConnection
	CheckParametersForConnections(clientParameters *commonbackend.CheckClientParameters) *commonbackend.CheckParametersResponse
	CheckParametersForBackend(arguments []byte) *commonbackend.CheckParametersResponse
}
