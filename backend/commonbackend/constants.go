package commonbackend

type Start struct {
	Type      string // Will be 'start' always
	Arguments []byte
}

type Stop struct {
	Type string // Will be 'stop' always
}

type AddProxy struct {
	Type       string // Will be 'addProxy' always
	SourceIP   string
	SourcePort uint16
	DestPort   uint16
	Protocol   string // Will be either 'tcp' or 'udp'
}

type RemoveProxy struct {
	Type       string // Will be 'removeProxy' always
	SourceIP   string
	SourcePort uint16
	DestPort   uint16
	Protocol   string // Will be either 'tcp' or 'udp'
}

type ProxyStatusRequest struct {
	Type       string // Will be 'proxyStatusRequest' always
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

type ProxyInstance struct {
	SourceIP   string
	SourcePort uint16
	DestPort   uint16
	Protocol   string // Will be either 'tcp' or 'udp'
}

type ProxyInstanceResponse struct {
	Type    string           // Will be 'proxyConnectionResponse' always
	Proxies []*ProxyInstance // List of connections
}

type ProxyInstanceRequest struct {
	Type string // Will be 'proxyConnectionRequest' always
}

type BackendStatusResponse struct {
	Type       string // Will be 'backendStatusResponse' always
	IsRunning  bool   // True if running, false if not running
	StatusCode int    // Either the 'Success' or 'Failure' constant
	Message    string // String message from the client (ex. failed to dial TCP)
}

type BackendStatusRequest struct {
	Type string // Will be 'backendStatusRequest' always
}

type ProxyConnectionsRequest struct {
	Type string // Will be 'proxyConnectionsRequest' always
}

// Client's connection to a specific proxy
type ProxyClientConnection struct {
	SourceIP   string
	SourcePort uint16
	DestPort   uint16
	ClientIP   string
	ClientPort uint16
}

type ProxyConnectionsResponse struct {
	Type        string                   // Will be 'proxyConnectionsResponse' always
	Connections []*ProxyClientConnection // List of connections
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
	StartID = iota
	StopID
	AddProxyID
	RemoveProxyID
	ProxyConnectionsResponseID
	CheckClientParametersID
	CheckServerParametersID
	CheckParametersResponseID
	ProxyConnectionsRequestID
	BackendStatusResponseID
	BackendStatusRequestID
	ProxyStatusRequestID
	ProxyStatusResponseID
	ProxyInstanceResponseID
	ProxyInstanceRequestID
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
