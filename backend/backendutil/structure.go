package backendutil

import (
	"capnproto.org/go/capnp/v3"
	"git.terah.dev/imterah/hermes/backend/commonbackend"
)

type BackendInterface interface {
	StartBackend(arguments []byte) (bool, error)
	StopBackend() (bool, error)
	GetBackendStatus() (bool, error)
	StartProxy(command *commonbackend.AddProxy) (bool, error)
	StopProxy(command *commonbackend.RemoveProxy) (bool, error)
	GetAllClientConnections(seg *capnp.Segment) []*commonbackend.Connection
	CheckParametersForConnections(clientParameters *commonbackend.CheckClientParameters) *CheckParametersResponse
	CheckParametersForBackend(arguments []byte) *CheckParametersResponse
}

// Sent as a response to either CheckClientParameters or CheckBackendParameters
type CheckParametersResponse struct {
	Type         string // Will be 'checkParametersResponse' always
	InResponseTo string // Will be either 'checkClientParameters' or 'checkServerParameters'
	IsValid      bool   // If true, valid, and if false, invalid
	Message      string // String message from the client (ex. failed to unmarshal JSON: x is not defined)
}
