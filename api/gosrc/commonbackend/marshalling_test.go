package commonbackend

import (
	"bytes"
	"log"
	"os"
	"testing"
)

var logLevel = os.Getenv("NEXTNET_LOG_LEVEL")

func TestStartCommandMarshalSupport(t *testing.T) {
	commandInput := &StartCommand{
		Type:      "start",
		Arguments: []byte("Hello from automated testing"),
	}

	commandMarshalled, err := Marshal(commandInput.Type, commandInput)

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	if err != nil {
		t.Fatalf(err.Error())
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

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*StartCommand)

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
	commandInput := &StopCommand{
		Type: "stop",
	}

	commandMarshalled, err := Marshal(commandInput.Type, commandInput)

	if logLevel == "debug" {
		log.Printf("Generated array contents: %v", commandMarshalled)
	}

	if err != nil {
		t.Fatalf(err.Error())
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

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*StopCommand)

	if !ok {
		t.Fatal("failed typecast")
	}

	if commandInput.Type != commandUnmarshalled.Type {
		t.Fail()
		log.Printf("Types are not equal (orig: %s, unmsh: %s)", commandInput.Type, commandUnmarshalled.Type)
	}
}

func TestAddConnectionCommandMarshalSupport(t *testing.T) {
	commandInput := &AddConnectionCommand{
		Type:       "addConnection",
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
		t.Fatalf(err.Error())
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

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*AddConnectionCommand)

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
	commandInput := &RemoveConnectionCommand{
		Type:       "removeConnection",
		SourceIP:   "192.168.0.139",
		SourcePort: 19132,
		DestPort:   19132,
		Protocol:   "tcp",
	}

	commandMarshalled, err := Marshal(commandInput.Type, commandInput)

	if err != nil {
		t.Fatalf(err.Error())
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

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*RemoveConnectionCommand)

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
	commandInput := &GetAllConnections{
		Type: "getAllConnections",
		Connections: []*ClientConnection{
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
		t.Fatalf(err.Error())
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

	commandUnmarshalled, ok := commandUnmarshalledRaw.(*GetAllConnections)

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

		if originalConnection.SourcePort != remoteConnection.SourcePort {
			t.Fail()
			log.Printf("(in #%d) DestPort's are not equal (orig: %d, unmsh: %d)", commandIndex, originalConnection.DestPort, remoteConnection.DestPort)
		}

		if originalConnection.SourcePort != remoteConnection.SourcePort {
			t.Fail()
			log.Printf("(in #%d) ClientIP's are not equal (orig: %s, unmsh: %s)", commandIndex, originalConnection.ClientIP, remoteConnection.ClientIP)
		}

		if originalConnection.SourcePort != remoteConnection.SourcePort {
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
		t.Fatalf(err.Error())
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
		t.Fatalf(err.Error())
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
		Type:      "checkParametersResponse",
		InReplyTo: "checkClientParameters",
		IsValid:   true,
		Message:   "Hello from automated testing",
	}

	commandMarshalled, err := Marshal(commandInput.Type, commandInput)

	if err != nil {
		t.Fatalf(err.Error())
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

	if commandInput.InReplyTo != commandUnmarshalled.InReplyTo {
		t.Fail()
		log.Printf("InReplyTo's are not equal (orig: %s, unmsh: %s)", commandInput.InReplyTo, commandUnmarshalled.InReplyTo)
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
