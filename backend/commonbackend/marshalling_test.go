package commonbackend

import (
	"bytes"
	"log"
	"os"
	"testing"
)

var logLevel = os.Getenv("HERMES_LOG_LEVEL")

func TestStartCommandMarshalSupport(t *testing.T) {
	commandInput := &Start{
		Type:      "start",
		Arguments: []byte("Hello from automated testing"),
	}

	commandMarshalled, err := Marshal(commandInput.Type, commandInput)

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	if err != nil {
		t.Fatal(err.Error())
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandType, commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	if commandType != commandInput.Type {
		t.Fail()
		log.Print("command type does not match up!")
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*Start)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.Type != commandUnmarshalled.Type {
		t.Fail()
		log.Printf("Types are not equal (orig: %s, unmsh: %s)", commandInput.Type, commandUnmarshalled.Type)
	}

	if !bytes.Equal(commandInput.Arguments, commandUnmarshalled.Arguments) {
		log.Fatalf("Arguments are not equal (orig: '%s', unmsh: '%s')", string(commandInput.Arguments), string(commandUnmarshalled.Arguments))
	}
}

func TestStopCommandMarshalSupport(t *testing.T) {
	commandInput := &Stop{
		Type: "stop",
	}

	commandMarshalled, err := Marshal(commandInput.Type, commandInput)

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	if err != nil {
		t.Fatal(err.Error())
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandType, commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	if commandType != commandInput.Type {
		t.Fail()
		log.Print("command type does not match up!")
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*Stop)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.Type != commandUnmarshalled.Type {
		t.Fail()
		log.Printf("Types are not equal (orig: %s, unmsh: %s)", commandInput.Type, commandUnmarshalled.Type)
	}
}

func TestAddConnectionCommandMarshalSupport(t *testing.T) {
	commandInput := &AddProxy{
		Type:       "addProxy",
		SourceIP:   "192.168.0.139",
		SourcePort: 19132,
		DestPort:   19132,
		Protocol:   "tcp",
	}

	commandMarshalled, err := Marshal(commandInput.Type, commandInput)

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	if err != nil {
		t.Fatal(err.Error())
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandType, commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	if commandType != commandInput.Type {
		t.Fail()
		log.Print("command type does not match up!")
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*AddProxy)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.Type != commandUnmarshalled.Type {
		t.Fail()
		log.Printf("Types are not equal (orig: %s, unmsh: %s)", commandInput.Type, commandUnmarshalled.Type)
	}

	if commandInput.SourceIP != commandUnmarshalled.SourceIP {
		t.Fail()
		log.Printf("SourceIP's are not equal (orig: %s, unmsh: %s)", commandInput.SourceIP, commandUnmarshalled.SourceIP)
	}

	if commandInput.SourcePort != commandUnmarshalled.SourcePort {
		t.Fail()
		log.Printf("SourcePort's are not equal (orig: %d, unmsh: %d)", commandInput.SourcePort, commandUnmarshalled.SourcePort)
	}

	if commandInput.DestPort != commandUnmarshalled.DestPort {
		t.Fail()
		log.Printf("DestPort's are not equal (orig: %d, unmsh: %d)", commandInput.DestPort, commandUnmarshalled.DestPort)
	}

	if commandInput.Protocol != commandUnmarshalled.Protocol {
		t.Fail()
		log.Printf("Protocols are not equal (orig: %s, unmsh: %s)", commandInput.Protocol, commandUnmarshalled.Protocol)
	}
}

func TestRemoveConnectionCommandMarshalSupport(t *testing.T) {
	commandInput := &RemoveProxy{
		Type:       "removeProxy",
		SourceIP:   "192.168.0.139",
		SourcePort: 19132,
		DestPort:   19132,
		Protocol:   "tcp",
	}

	commandMarshalled, err := Marshal(commandInput.Type, commandInput)

	if err != nil {
		t.Fatal(err.Error())
	}

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandType, commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	if commandType != commandInput.Type {
		t.Fail()
		log.Print("command type does not match up!")
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*RemoveProxy)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.Type != commandUnmarshalled.Type {
		t.Fail()
		log.Printf("Types are not equal (orig: %s, unmsh: %s)", commandInput.Type, commandUnmarshalled.Type)
	}

	if commandInput.SourceIP != commandUnmarshalled.SourceIP {
		t.Fail()
		log.Printf("SourceIP's are not equal (orig: %s, unmsh: %s)", commandInput.SourceIP, commandUnmarshalled.SourceIP)
	}

	if commandInput.SourcePort != commandUnmarshalled.SourcePort {
		t.Fail()
		log.Printf("SourcePort's are not equal (orig: %d, unmsh: %d)", commandInput.SourcePort, commandUnmarshalled.SourcePort)
	}

	if commandInput.DestPort != commandUnmarshalled.DestPort {
		t.Fail()
		log.Printf("DestPort's are not equal (orig: %d, unmsh: %d)", commandInput.DestPort, commandUnmarshalled.DestPort)
	}

	if commandInput.Protocol != commandUnmarshalled.Protocol {
		t.Fail()
		log.Printf("Protocols are not equal (orig: %s, unmsh: %s)", commandInput.Protocol, commandUnmarshalled.Protocol)
	}
}

func TestGetAllConnectionsCommandMarshalSupport(t *testing.T) {
	commandInput := &ProxyConnectionsResponse{
		Type: "proxyConnectionsResponse",
		Connections: []*ProxyClientConnection{
			{
				SourceIP:   "127.0.0.1",
				SourcePort: 19132,
				DestPort:   19132,
				ClientIP:   "127.0.0.1",
				ClientPort: 12321,
			},
			{
				SourceIP:   "127.0.0.1",
				SourcePort: 19132,
				DestPort:   19132,
				ClientIP:   "192.168.0.168",
				ClientPort: 23457,
			},
			{
				SourceIP:   "127.0.0.1",
				SourcePort: 19132,
				DestPort:   19132,
				ClientIP:   "68.42.203.47",
				ClientPort: 38721,
			},
		},
	}

	commandMarshalled, err := Marshal(commandInput.Type, commandInput)

	if err != nil {
		t.Fatal(err.Error())
	}

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandType, commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	if commandType != commandInput.Type {
		t.Fail()
		log.Print("command type does not match up!")
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*ProxyConnectionsResponse)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.Type != commandUnmarshalled.Type {
		t.Fail()
		log.Printf("Types are not equal (orig: %s, unmsh: %s)", commandInput.Type, commandUnmarshalled.Type)
	}

	for commandIndex, originalConnection := range commandInput.Connections {
		remoteConnection := commandUnmarshalled.Connections[commandIndex]

		if originalConnection.SourceIP != remoteConnection.SourceIP {
			t.Fail()
			log.Printf("(in #%d) SourceIP's are not equal (orig: %s, unmsh: %s)", commandIndex, originalConnection.SourceIP, remoteConnection.SourceIP)
		}

		if originalConnection.SourcePort != remoteConnection.SourcePort {
			t.Fail()
			log.Printf("(in #%d) SourcePort's are not equal (orig: %d, unmsh: %d)", commandIndex, originalConnection.SourcePort, remoteConnection.SourcePort)
		}

		if originalConnection.DestPort != remoteConnection.DestPort {
			t.Fail()
			log.Printf("(in #%d) DestPort's are not equal (orig: %d, unmsh: %d)", commandIndex, originalConnection.DestPort, remoteConnection.DestPort)
		}

		if originalConnection.ClientIP != remoteConnection.ClientIP {
			t.Fail()
			log.Printf("(in #%d) ClientIP's are not equal (orig: %s, unmsh: %s)", commandIndex, originalConnection.ClientIP, remoteConnection.ClientIP)
		}

		if originalConnection.ClientPort != remoteConnection.ClientPort {
			t.Fail()
			log.Printf("(in #%d) ClientPort's are not equal (orig: %d, unmsh: %d)", commandIndex, originalConnection.ClientPort, remoteConnection.ClientPort)
		}
	}
}

func TestCheckClientParametersMarshalSupport(t *testing.T) {
	commandInput := &CheckClientParameters{
		Type:       "checkClientParameters",
		SourceIP:   "192.168.0.139",
		SourcePort: 19132,
		DestPort:   19132,
		Protocol:   "tcp",
	}

	commandMarshalled, err := Marshal(commandInput.Type, commandInput)

	if err != nil {
		t.Fatal(err.Error())
	}

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandType, commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	if commandType != commandInput.Type {
		t.Fail()
		log.Printf("command type does not match up! (orig: %s, unmsh: %s)", commandType, commandInput.Type)
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*CheckClientParameters)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.Type != commandUnmarshalled.Type {
		t.Fail()
		log.Printf("Types are not equal (orig: %s, unmsh: %s)", commandInput.Type, commandUnmarshalled.Type)
	}

	if commandInput.SourceIP != commandUnmarshalled.SourceIP {
		t.Fail()
		log.Printf("SourceIP's are not equal (orig: %s, unmsh: %s)", commandInput.SourceIP, commandUnmarshalled.SourceIP)
	}

	if commandInput.SourcePort != commandUnmarshalled.SourcePort {
		t.Fail()
		log.Printf("SourcePort's are not equal (orig: %d, unmsh: %d)", commandInput.SourcePort, commandUnmarshalled.SourcePort)
	}

	if commandInput.DestPort != commandUnmarshalled.DestPort {
		t.Fail()
		log.Printf("DestPort's are not equal (orig: %d, unmsh: %d)", commandInput.DestPort, commandUnmarshalled.DestPort)
	}

	if commandInput.Protocol != commandUnmarshalled.Protocol {
		t.Fail()
		log.Printf("Protocols are not equal (orig: %s, unmsh: %s)", commandInput.Protocol, commandUnmarshalled.Protocol)
	}
}

func TestCheckServerParametersMarshalSupport(t *testing.T) {
	commandInput := &CheckServerParameters{
		Type:      "checkServerParameters",
		Arguments: []byte("Hello from automated testing"),
	}

	commandMarshalled, err := Marshal(commandInput.Type, commandInput)

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	if err != nil {
		t.Fatal(err.Error())
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandType, commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	if commandType != commandInput.Type {
		t.Fail()
		log.Print("command type does not match up!")
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*CheckServerParameters)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.Type != commandUnmarshalled.Type {
		t.Fail()
		log.Printf("Types are not equal (orig: %s, unmsh: %s)", commandInput.Type, commandUnmarshalled.Type)
	}

	if !bytes.Equal(commandInput.Arguments, commandUnmarshalled.Arguments) {
		log.Fatalf("Arguments are not equal (orig: '%s', unmsh: '%s')", string(commandInput.Arguments), string(commandUnmarshalled.Arguments))
	}
}

func TestCheckParametersResponseMarshalSupport(t *testing.T) {
	commandInput := &CheckParametersResponse{
		Type:         "checkParametersResponse",
		InResponseTo: "checkClientParameters",
		IsValid:      true,
		Message:      "Hello from automated testing",
	}

	commandMarshalled, err := Marshal(commandInput.Type, commandInput)

	if err != nil {
		t.Fatal(err.Error())
	}

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandType, commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	if commandType != commandInput.Type {
		t.Fail()
		log.Printf("command type does not match up! (orig: %s, unmsh: %s)", commandType, commandInput.Type)
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*CheckParametersResponse)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.Type != commandUnmarshalled.Type {
		t.Fail()
		log.Printf("Types are not equal (orig: %s, unmsh: %s)", commandInput.Type, commandUnmarshalled.Type)
	}

	if commandInput.InResponseTo != commandUnmarshalled.InResponseTo {
		t.Fail()
		log.Printf("InResponseTo's are not equal (orig: %s, unmsh: %s)", commandInput.InResponseTo, commandUnmarshalled.InResponseTo)
	}

	if commandInput.IsValid != commandUnmarshalled.IsValid {
		t.Fail()
		log.Printf("IsValid's are not equal (orig: %t, unmsh: %t)", commandInput.IsValid, commandUnmarshalled.IsValid)
	}

	if commandInput.Message != commandUnmarshalled.Message {
		t.Fail()
		log.Printf("Messages are not equal (orig: %s, unmsh: %s)", commandInput.Message, commandUnmarshalled.Message)
	}
}

func TestBackendStatusRequestMarshalSupport(t *testing.T) {
	commandInput := &BackendStatusRequest{
		Type: "backendStatusRequest",
	}

	commandMarshalled, err := Marshal(commandInput.Type, commandInput)

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	if err != nil {
		t.Fatal(err.Error())
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandType, commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	if commandType != commandInput.Type {
		t.Fail()
		log.Print("command type does not match up!")
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*BackendStatusRequest)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.Type != commandUnmarshalled.Type {
		t.Fail()
		log.Printf("Types are not equal (orig: %s, unmsh: %s)", commandInput.Type, commandUnmarshalled.Type)
	}
}

func TestBackendStatusResponseMarshalSupport(t *testing.T) {
	commandInput := &BackendStatusResponse{
		Type:       "backendStatusResponse",
		IsRunning:  true,
		StatusCode: StatusFailure,
		Message:    "Hello from automated testing",
	}

	commandMarshalled, err := Marshal(commandInput.Type, commandInput)

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	if err != nil {
		t.Fatal(err.Error())
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandType, commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	if commandType != commandInput.Type {
		t.Fail()
		log.Print("command type does not match up!")
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*BackendStatusResponse)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.Type != commandUnmarshalled.Type {
		t.Fail()
		log.Printf("Types are not equal (orig: %s, unmsh: %s)", commandInput.Type, commandUnmarshalled.Type)
	}

	if commandInput.IsRunning != commandUnmarshalled.IsRunning {
		t.Fail()
		log.Printf("IsRunning's are not equal (orig: %t, unmsh: %t)", commandInput.IsRunning, commandUnmarshalled.IsRunning)
	}

	if commandInput.StatusCode != commandUnmarshalled.StatusCode {
		t.Fail()
		log.Printf("StatusCodes are not equal (orig: %d, unmsh: %d)", commandInput.StatusCode, commandUnmarshalled.StatusCode)
	}

	if commandInput.Message != commandUnmarshalled.Message {
		t.Fail()
		log.Printf("Messages are not equal (orig: %s, unmsh: %s)", commandInput.Message, commandUnmarshalled.Message)
	}
}

func TestProxyStatusRequestMarshalSupport(t *testing.T) {
	commandInput := &ProxyStatusRequest{
		Type:       "proxyStatusRequest",
		SourceIP:   "192.168.0.139",
		SourcePort: 19132,
		DestPort:   19132,
		Protocol:   "tcp",
	}

	commandMarshalled, err := Marshal(commandInput.Type, commandInput)

	if err != nil {
		t.Fatal(err.Error())
	}

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandType, commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	if commandType != commandInput.Type {
		t.Fail()
		log.Print("command type does not match up!")
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*ProxyStatusRequest)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.Type != commandUnmarshalled.Type {
		t.Fail()
		log.Printf("Types are not equal (orig: %s, unmsh: %s)", commandInput.Type, commandUnmarshalled.Type)
	}

	if commandInput.SourceIP != commandUnmarshalled.SourceIP {
		t.Fail()
		log.Printf("SourceIP's are not equal (orig: %s, unmsh: %s)", commandInput.SourceIP, commandUnmarshalled.SourceIP)
	}

	if commandInput.SourcePort != commandUnmarshalled.SourcePort {
		t.Fail()
		log.Printf("SourcePort's are not equal (orig: %d, unmsh: %d)", commandInput.SourcePort, commandUnmarshalled.SourcePort)
	}

	if commandInput.DestPort != commandUnmarshalled.DestPort {
		t.Fail()
		log.Printf("DestPort's are not equal (orig: %d, unmsh: %d)", commandInput.DestPort, commandUnmarshalled.DestPort)
	}

	if commandInput.Protocol != commandUnmarshalled.Protocol {
		t.Fail()
		log.Printf("Protocols are not equal (orig: %s, unmsh: %s)", commandInput.Protocol, commandUnmarshalled.Protocol)
	}
}

func TestProxyStatusResponseMarshalSupport(t *testing.T) {
	commandInput := &ProxyStatusResponse{
		Type:       "proxyStatusResponse",
		SourceIP:   "192.168.0.139",
		SourcePort: 19132,
		DestPort:   19132,
		Protocol:   "tcp",
		IsActive:   true,
	}

	commandMarshalled, err := Marshal(commandInput.Type, commandInput)

	if err != nil {
		t.Fatal(err.Error())
	}

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandType, commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	if commandType != commandInput.Type {
		t.Fail()
		log.Print("command type does not match up!")
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*ProxyStatusResponse)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.Type != commandUnmarshalled.Type {
		t.Fail()
		log.Printf("Types are not equal (orig: %s, unmsh: %s)", commandInput.Type, commandUnmarshalled.Type)
	}

	if commandInput.SourceIP != commandUnmarshalled.SourceIP {
		t.Fail()
		log.Printf("SourceIP's are not equal (orig: %s, unmsh: %s)", commandInput.SourceIP, commandUnmarshalled.SourceIP)
	}

	if commandInput.SourcePort != commandUnmarshalled.SourcePort {
		t.Fail()
		log.Printf("SourcePort's are not equal (orig: %d, unmsh: %d)", commandInput.SourcePort, commandUnmarshalled.SourcePort)
	}

	if commandInput.DestPort != commandUnmarshalled.DestPort {
		t.Fail()
		log.Printf("DestPort's are not equal (orig: %d, unmsh: %d)", commandInput.DestPort, commandUnmarshalled.DestPort)
	}

	if commandInput.Protocol != commandUnmarshalled.Protocol {
		t.Fail()
		log.Printf("Protocols are not equal (orig: %s, unmsh: %s)", commandInput.Protocol, commandUnmarshalled.Protocol)
	}

	if commandInput.IsActive != commandUnmarshalled.IsActive {
		t.Fail()
		log.Printf("IsActive's are not equal (orig: %t, unmsh: %t)", commandInput.IsActive, commandUnmarshalled.IsActive)
	}
}

func TestProxyConnectionRequestMarshalSupport(t *testing.T) {
	commandInput := &ProxyInstanceRequest{
		Type: "proxyInstanceRequest",
	}

	commandMarshalled, err := Marshal(commandInput.Type, commandInput)

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	if err != nil {
		t.Fatal(err.Error())
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandType, commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	if commandType != commandInput.Type {
		t.Fail()
		log.Print("command type does not match up!")
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*ProxyInstanceRequest)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.Type != commandUnmarshalled.Type {
		t.Fail()
		log.Printf("Types are not equal (orig: %s, unmsh: %s)", commandInput.Type, commandUnmarshalled.Type)
	}
}

func TestProxyConnectionResponseMarshalSupport(t *testing.T) {
	commandInput := &ProxyInstanceResponse{
		Type: "proxyInstanceResponse",
		Proxies: []*ProxyInstance{
			{
				SourceIP:   "192.168.0.168",
				SourcePort: 25565,
				DestPort:   25565,
				Protocol:   "tcp",
			},
			{
				SourceIP:   "127.0.0.1",
				SourcePort: 19132,
				DestPort:   19132,
				Protocol:   "udp",
			},
			{
				SourceIP:   "68.42.203.47",
				SourcePort: 22,
				DestPort:   2222,
				Protocol:   "tcp",
			},
		},
	}

	commandMarshalled, err := Marshal(commandInput.Type, commandInput)

	if err != nil {
		t.Fatal(err.Error())
	}

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	buf := bytes.NewBuffer(commandMarshalled)
	commandType, commandUnmarshalledRaw, err := Unmarshal(buf)

	if err != nil {
		t.Fatal(err.Error())
	}

	if commandType != commandInput.Type {
		t.Fail()
		log.Print("command type does not match up!")
	}

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*ProxyInstanceResponse)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.Type != commandUnmarshalled.Type {
		t.Fail()
		log.Printf("Types are not equal (orig: %s, unmsh: %s)", commandInput.Type, commandUnmarshalled.Type)
	}

	for proxyIndex, originalProxy := range commandInput.Proxies {
		remoteProxy := commandUnmarshalled.Proxies[proxyIndex]

		if originalProxy.SourceIP != remoteProxy.SourceIP {
			t.Fail()
			log.Printf("(in #%d) SourceIP's are not equal (orig: %s, unmsh: %s)", proxyIndex, originalProxy.SourceIP, remoteProxy.SourceIP)
		}

		if originalProxy.SourcePort != remoteProxy.SourcePort {
			t.Fail()
			log.Printf("(in #%d) SourcePort's are not equal (orig: %d, unmsh: %d)", proxyIndex, originalProxy.SourcePort, remoteProxy.SourcePort)
		}

		if originalProxy.DestPort != remoteProxy.DestPort {
			t.Fail()
			log.Printf("(in #%d) DestPort's are not equal (orig: %d, unmsh: %d)", proxyIndex, originalProxy.DestPort, remoteProxy.DestPort)
		}

		if originalProxy.Protocol != remoteProxy.Protocol {
			t.Fail()
			log.Printf("(in #%d) ClientIP's are not equal (orig: %s, unmsh: %s)", proxyIndex, originalProxy.Protocol, remoteProxy.Protocol)
		}
	}
}
