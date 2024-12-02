package commonbackend

import (
	"encoding/binary"
	"fmt"
	"net"
)

func marshalIndividualConnectionStruct(conn *ClientConnection) []byte {
	sourceIPOriginal := net.ParseIP(conn.SourceIP)
	clientIPOriginal := net.ParseIP(conn.ClientIP)

	var serverIPVer uint8
	var sourceIP []byte

	if sourceIPOriginal.To4() == nil {
		serverIPVer = IPv6
		sourceIP = sourceIPOriginal.To16()
	} else {
		serverIPVer = IPv4
		sourceIP = sourceIPOriginal.To4()
	}

	var clientIPVer uint8
	var clientIP []byte

	if clientIPOriginal.To4() == nil {
		clientIPVer = IPv6
		clientIP = clientIPOriginal.To16()
	} else {
		clientIPVer = IPv4
		clientIP = clientIPOriginal.To4()
	}

	connectionBlock := make([]byte, 8+len(sourceIP)+len(clientIP))

	connectionBlock[0] = serverIPVer
	copy(connectionBlock[1:len(sourceIP)+1], sourceIP)
	binary.BigEndian.PutUint16(connectionBlock[1+len(sourceIP):3+len(sourceIP)], conn.SourcePort)
	binary.BigEndian.PutUint16(connectionBlock[3+len(sourceIP):5+len(sourceIP)], conn.DestPort)

	connectionBlock[5+len(sourceIP)] = clientIPVer
	copy(connectionBlock[6+len(sourceIP):6+len(sourceIP)+len(clientIP)], clientIP)
	binary.BigEndian.PutUint16(connectionBlock[6+len(sourceIP)+len(clientIP):8+len(sourceIP)+len(clientIP)], conn.ClientPort)

	return connectionBlock
}

func Marshal(commandType string, command interface{}) ([]byte, error) {
	switch commandType {
	case "start":
		startCommand, ok := command.(*StartCommand)

		if !ok {
			return nil, fmt.Errorf("failed to typecast")
		}

		startCommandBytes := make([]byte, 1+2+len(startCommand.Arguments))
		startCommandBytes[0] = StartCommandID
		binary.BigEndian.PutUint16(startCommandBytes[1:3], uint16(len(startCommand.Arguments)))
		copy(startCommandBytes[3:], startCommand.Arguments)

		return startCommandBytes, nil
	case "stop":
		_, ok := command.(*StopCommand)

		if !ok {
			return nil, fmt.Errorf("failed to typecast")
		}

		stopCommandBytes := make([]byte, 1)
		stopCommandBytes[0] = StopCommandID

		return stopCommandBytes, nil
	case "addConnection":
		addConnectionCommand, ok := command.(*AddConnectionCommand)

		if !ok {
			return nil, fmt.Errorf("failed to typecast")
		}

		sourceIP := net.ParseIP(addConnectionCommand.SourceIP)

		var ipVer uint8
		var ipBytes []byte

		if sourceIP.To4() == nil {
			ipBytes = sourceIP.To16()
			ipVer = IPv6
		} else {
			ipBytes = sourceIP.To4()
			ipVer = IPv4
		}

		addConnectionBytes := make([]byte, 1+1+len(ipBytes)+2+2+1)

		addConnectionBytes[0] = AddConnectionCommandID
		addConnectionBytes[1] = ipVer

		copy(addConnectionBytes[2:2+len(ipBytes)], ipBytes)

		binary.BigEndian.PutUint16(addConnectionBytes[2+len(ipBytes):4+len(ipBytes)], addConnectionCommand.SourcePort)
		binary.BigEndian.PutUint16(addConnectionBytes[4+len(ipBytes):6+len(ipBytes)], addConnectionCommand.DestPort)

		var protocol uint8

		if addConnectionCommand.Protocol == "tcp" {
			protocol = TCP
		} else if addConnectionCommand.Protocol == "udp" {
			protocol = UDP
		} else {
			return nil, fmt.Errorf("invalid protocol")
		}

		addConnectionBytes[6+len(ipBytes)] = protocol

		return addConnectionBytes, nil
	case "removeConnection":
		removeConnectionCommand, ok := command.(*RemoveConnectionCommand)

		if !ok {
			return nil, fmt.Errorf("failed to typecast")
		}

		sourceIP := net.ParseIP(removeConnectionCommand.SourceIP)

		var ipVer uint8
		var ipBytes []byte

		if sourceIP.To4() == nil {
			ipBytes = sourceIP.To16()
			ipVer = IPv6
		} else {
			ipBytes = sourceIP.To4()
			ipVer = IPv4
		}

		removeConnectionBytes := make([]byte, 1+1+len(ipBytes)+2+2+1)

		removeConnectionBytes[0] = RemoveConnectionCommandID
		removeConnectionBytes[1] = ipVer
		copy(removeConnectionBytes[2:2+len(ipBytes)], ipBytes)
		binary.BigEndian.PutUint16(removeConnectionBytes[2+len(ipBytes):4+len(ipBytes)], removeConnectionCommand.SourcePort)
		binary.BigEndian.PutUint16(removeConnectionBytes[4+len(ipBytes):6+len(ipBytes)], removeConnectionCommand.DestPort)

		var protocol uint8

		if removeConnectionCommand.Protocol == "tcp" {
			protocol = TCP
		} else if removeConnectionCommand.Protocol == "udp" {
			protocol = UDP
		} else {
			return nil, fmt.Errorf("invalid protocol")
		}

		removeConnectionBytes[6+len(ipBytes)] = protocol

		return removeConnectionBytes, nil
	case "getAllConnections":
		allConnectionsCommand, ok := command.(*GetAllConnections)

		if !ok {
			return nil, fmt.Errorf("failed to typecast")
		}

		connectionsArray := make([][]byte, len(allConnectionsCommand.Connections))
		totalSize := 0

		for connIndex, conn := range allConnectionsCommand.Connections {
			connectionsArray[connIndex] = marshalIndividualConnectionStruct(conn)
			totalSize += len(connectionsArray[connIndex]) + 1
		}

		connectionCommandArray := make([]byte, totalSize+1)
		connectionCommandArray[0] = GetAllConnectionsID

		currentPosition := 1

		for _, connection := range connectionsArray {
			copy(connectionCommandArray[currentPosition:currentPosition+len(connection)], connection)
			connectionCommandArray[currentPosition+len(connection)] = '\r'
			currentPosition += len(connection) + 1
		}

		connectionCommandArray[totalSize] = '\n'
		return connectionCommandArray, nil
	case "checkClientParameters":
		checkClientCommand, ok := command.(*CheckClientParameters)

		if !ok {
			return nil, fmt.Errorf("failed to typecast")
		}

		sourceIP := net.ParseIP(checkClientCommand.SourceIP)

		var ipVer uint8
		var ipBytes []byte

		if sourceIP.To4() == nil {
			ipBytes = sourceIP.To16()
			ipVer = IPv6
		} else {
			ipBytes = sourceIP.To4()
			ipVer = IPv4
		}

		checkClientBytes := make([]byte, 1+1+len(ipBytes)+2+2+1)

		checkClientBytes[0] = CheckClientParametersID
		checkClientBytes[1] = ipVer
		copy(checkClientBytes[2:2+len(ipBytes)], ipBytes)
		binary.BigEndian.PutUint16(checkClientBytes[2+len(ipBytes):4+len(ipBytes)], checkClientCommand.SourcePort)
		binary.BigEndian.PutUint16(checkClientBytes[4+len(ipBytes):6+len(ipBytes)], checkClientCommand.DestPort)

		var protocol uint8

		if checkClientCommand.Protocol == "tcp" {
			protocol = TCP
		} else if checkClientCommand.Protocol == "udp" {
			protocol = UDP
		} else {
			return nil, fmt.Errorf("invalid protocol")
		}

		checkClientBytes[6+len(ipBytes)] = protocol

		return checkClientBytes, nil
	case "checkServerParameters":
		checkServerCommand, ok := command.(*CheckServerParameters)

		if !ok {
			return nil, fmt.Errorf("failed to typecast")
		}

		serverCommandBytes := make([]byte, 1+2+len(checkServerCommand.Arguments))
		serverCommandBytes[0] = CheckServerParametersID
		binary.BigEndian.PutUint16(serverCommandBytes[1:3], uint16(len(checkServerCommand.Arguments)))
		copy(serverCommandBytes[3:], checkServerCommand.Arguments)

		return serverCommandBytes, nil
	case "checkParametersResponse":
		checkParametersCommand, ok := command.(*CheckParametersResponse)

		if !ok {
			return nil, fmt.Errorf("failed to typecast")
		}

		var checkMethod uint8

		if checkParametersCommand.InReplyTo == "checkClientParameters" {
			checkMethod = CheckClientParametersID
		} else if checkParametersCommand.InReplyTo == "checkServerParameters" {
			checkMethod = CheckServerParametersID
		} else {
			return nil, fmt.Errorf("invalid mode recieved (must be either checkClientParameters or checkServerParameters)")
		}

		var isValid uint8

		if checkParametersCommand.IsValid {
			isValid = 1
		}

		checkResponseBytes := make([]byte, 3+2+len(checkParametersCommand.Message))
		checkResponseBytes[0] = CheckParametersResponseID
		checkResponseBytes[1] = checkMethod
		checkResponseBytes[2] = isValid

		binary.BigEndian.PutUint16(checkResponseBytes[3:5], uint16(len(checkParametersCommand.Message)))

		if len(checkParametersCommand.Message) != 0 {
			copy(checkResponseBytes[5:], []byte(checkParametersCommand.Message))
		}

		return checkResponseBytes, nil
	}

	return nil, fmt.Errorf("couldn't match command")
}
