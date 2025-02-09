package backendutil

import "git.terah.dev/imterah/hermes/backend/commonbackend"

type BackendInterface interface {
	StartBackend(arguments []byte) (bool, error)
	StopBackend() (bool, error)
	GetBackendStatus() (bool, error)
	StartProxy(command *commonbackend.AddProxy) (bool, error)
	StopProxy(command *commonbackend.RemoveProxy) (bool, error)
	GetAllClientConnections() []*commonbackend.ProxyClientConnection
	CheckParametersForConnections(clientParameters *commonbackend.CheckClientParameters) *commonbackend.CheckParametersResponse
	CheckParametersForBackend(arguments []byte) *commonbackend.CheckParametersResponse
}
