package commonbackend

// Not all of these structs are implemented commands.
// Currently unimplemented commands:
//   GetAllConnectionsRequest
//   BackendStatusResponse
//   BackendStatusRequest
//   ProxyStatusRequest
//   ProxyStatusResponse
//   GetAllConnectionsRequest

// TODO (imterah): Rename AddConnectionCommand/RemoveConnectionCommand to AddProxyCommand/RemoveProxyCommand
// and their associated function calls

type StartCommand struct {
	Type      string // Will be 'start' always
	Arguments []byte
}

type StopCommand struct {
	Type string // Will be 'stop' always
}

type AddConnectionCommand struct {
	Type       string // Will be 'addConnection' always
	SourceIP   string
	SourcePort uint16
	DestPort   uint16
	Protocol   string // Will be either 'tcp' or 'udp'
}

type RemoveConnectionCommand struct {
	Type       string // Will be 'removeConnection' always
	SourceIP   string
	SourcePort uint16
	DestPort   uint16
	Protocol   string // Will be either 'tcp' or 'udp'
}

type GetProxyStatus struct {
	Type       string // Will be 'getProxyStatus' always
	SourceIP   string
	SourcePort uint16
	DestPort   uint16
	Protocol   string // Will be either 'tcp' or 'udp'
}

type ProxyStatusResponse struct {
	Type       string // Will be 'proxyStatusResponse' always
	SourceIP   string
	SourcePort uint16
	DestPort   uint16
	Protocol   string // Will be either 'tcp' or 'udp'
	IsActive   bool
}

type ProxyConnection struct {
	SourceIP   string
	SourcePort uint16
	DestPort   uint16
	Protocol   string // Will be either 'tcp' or 'udp'
}

type ProxyConnectionResponse struct {
	Type        string             // Will be 'proxyConnectionResponse' always
	Connections []*ProxyConnection // List of connections
}

type BackendStatusResponse struct {
	Type         string // Will be 'backendStatusResponse' always
	InResponseTo string // Can be either for 'start' or 'stop'
	StatusCode   int    // Either the 'Success' or 'Failure' constant
	Message      string // String message from the client (ex. failed to dial TCP)
}

type BackendStatusRequest struct {
	Type        string // Will be 'backendStatusRequest' always
	ForProperty string // Can be either for 'start' or 'stop'
}

type GetAllConnectionsRequest struct {
	Type string // Will be 'getAllConnectionsRequest' always
}

type ClientConnection struct {
	SourceIP   string
	SourcePort uint16
	DestPort   uint16
	ClientIP   string
	ClientPort uint16
}

type ConnectionsResponse struct {
	Type        string              // Will be 'connectionsResponse' always
	Connections []*ClientConnection // List of connections
}

type CheckClientParameters struct {
	Type       string // Will be 'checkClientParameters' always
	SourceIP   string
	SourcePort uint16
	DestPort   uint16
	Protocol   string // Will be either 'tcp' or 'udp'
}

type CheckServerParameters struct {
	Type      string // Will be 'checkServerParameters' always
	Arguments []byte
}

// Sent as a response to either CheckClientParameters or CheckBackendParameters
type CheckParametersResponse struct {
	Type         string // Will be 'checkParametersResponse' always
	InResponseTo string // Will be either 'checkClientParameters' or 'checkServerParameters'
	IsValid      bool   // If true, valid, and if false, invalid
	Message      string // String message from the client (ex. failed to unmarshal JSON: x is not defined)
}

const (
	StartCommandID = iota
	StopCommandID
	AddConnectionCommandID
	RemoveConnectionCommandID
	ClientConnectionID
	GetAllConnectionsID
	CheckClientParametersID
	CheckServerParametersID
	CheckParametersResponseID
)

const (
	TCP = iota
	UDP
)

const (
	StatusSuccess = iota
	StatusFailure
)

const (
	// IP versions
	IPv4 = 4
	IPv6 = 6

	// TODO: net has these constants defined already. We should switch to these
	IPv4Size = 4
	IPv6Size = 16
)
