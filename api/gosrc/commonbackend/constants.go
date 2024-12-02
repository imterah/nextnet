package commonbackend

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

type ClientConnection struct {
	SourceIP   string
	SourcePort uint16
	DestPort   uint16
	ClientIP   string
	ClientPort uint16
}

type GetAllConnections struct {
	Type        string              // Will be 'getAllConnections' always
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
	Type      string // Will be 'checkParametersResponse' always
	InReplyTo string // Will be either 'checkClientParameters' or 'checkServerParameters'
	IsValid   bool   // If true, valid, and if false, invalid
	Message   string // String message from the client (ex. failed to unmarshal JSON: x is not defined)
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

	IPv4 = 4
	IPv6 = 6

	IPv4Size = 4
	IPv6Size = 16
)
