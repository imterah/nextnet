package backendutil

import "git.greysoh.dev/imterah/nextnet/commonbackend"

type BackendInterface interface {
	StartBackend(arguments []byte) (bool, error)
	StopBackend() (bool, error)
	StartProxy(command *commonbackend.AddConnectionCommand) (bool, error)
	StopProxy(command *commonbackend.RemoveConnectionCommand) (bool, error)
	GetAllClientConnections() []*commonbackend.ClientConnection
	CheckParametersForConnections(clientParameters *commonbackend.CheckClientParameters) *commonbackend.CheckParametersResponse
	CheckParametersForBackend(arguments []byte) *commonbackend.CheckParametersResponse
}
