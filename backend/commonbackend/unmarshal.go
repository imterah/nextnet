package commonbackend

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

func unmarshalIndividualConnectionStruct(conn io.Reader) (*ProxyClientConnection, error) {
	serverIPVersion := make([]byte, 1)

	if _, err := conn.Read(serverIPVersion); err != nil {
		return nil, fmt.Errorf("couldn't read server IP version")
	}

	var serverIPSize uint8

	if serverIPVersion[0] == 4 {
		serverIPSize = IPv4Size
	} else if serverIPVersion[0] == 6 {
		serverIPSize = IPv6Size
	} else {
		return nil, fmt.Errorf("invalid server IP version recieved")
	}

	serverIP := make(net.IP, serverIPSize)

	if _, err := conn.Read(serverIP); err != nil {
		return nil, fmt.Errorf("couldn't read server IP")
	}

	sourcePort := make([]byte, 2)

	if _, err := conn.Read(sourcePort); err != nil {
		return nil, fmt.Errorf("couldn't read source port")
	}

	destinationPort := make([]byte, 2)

	if _, err := conn.Read(destinationPort); err != nil {
		return nil, fmt.Errorf("couldn't read source port")
	}

	clientIPVersion := make([]byte, 1)

	if _, err := conn.Read(clientIPVersion); err != nil {
		return nil, fmt.Errorf("couldn't read server IP version")
	}

	var clientIPSize uint8

	if clientIPVersion[0] == 4 {
		clientIPSize = IPv4Size
	} else if clientIPVersion[0] == 6 {
		clientIPSize = IPv6Size
	} else {
		return nil, fmt.Errorf("invalid server IP version recieved")
	}

	clientIP := make(net.IP, clientIPSize)

	if _, err := conn.Read(clientIP); err != nil {
		return nil, fmt.Errorf("couldn't read server IP")
	}

	clientPort := make([]byte, 2)

	if _, err := conn.Read(clientPort); err != nil {
		return nil, fmt.Errorf("couldn't read source port")
	}

	return &ProxyClientConnection{
		SourceIP:   serverIP.String(),
		SourcePort: binary.BigEndian.Uint16(sourcePort),
		DestPort:   binary.BigEndian.Uint16(destinationPort),
		ClientIP:   clientIP.String(),
		ClientPort: binary.BigEndian.Uint16(clientPort),
	}, nil
}

func unmarshalIndividualProxyStruct(conn io.Reader) (*ProxyInstance, error) {
	ipVersion := make([]byte, 1)

	if _, err := conn.Read(ipVersion); err != nil {
		return nil, fmt.Errorf("couldn't read ip version")
	}

	var ipSize uint8

	if ipVersion[0] == 4 {
		ipSize = IPv4Size
	} else if ipVersion[0] == 6 {
		ipSize = IPv6Size
	} else {
		return nil, fmt.Errorf("invalid IP version recieved")
	}

	ip := make(net.IP, ipSize)

	if _, err := conn.Read(ip); err != nil {
		return nil, fmt.Errorf("couldn't read source IP")
	}

	sourcePort := make([]byte, 2)

	if _, err := conn.Read(sourcePort); err != nil {
		return nil, fmt.Errorf("couldn't read source port")
	}

	destPort := make([]byte, 2)

	if _, err := conn.Read(destPort); err != nil {
		return nil, fmt.Errorf("couldn't read destination port")
	}

	protocolBytes := make([]byte, 1)

	if _, err := conn.Read(protocolBytes); err != nil {
		return nil, fmt.Errorf("couldn't read protocol")
	}

	var protocol string

	if protocolBytes[0] == TCP {
		protocol = "tcp"
	} else if protocolBytes[0] == UDP {
		protocol = "udp"
	} else {
		return nil, fmt.Errorf("invalid protocol")
	}

	return &ProxyInstance{
		SourceIP:   ip.String(),
		SourcePort: binary.BigEndian.Uint16(sourcePort),
		DestPort:   binary.BigEndian.Uint16(destPort),
		Protocol:   protocol,
	}, nil
}

func Unmarshal(conn io.Reader) (string, interface{}, error) {
	commandType := make([]byte, 1)

	if _, err := conn.Read(commandType); err != nil {
		return "", nil, fmt.Errorf("couldn't read command")
	}

	switch commandType[0] {
	case StartID:
		argumentsLength := make([]byte, 2)

		if _, err := conn.Read(argumentsLength); err != nil {
			return "", nil, fmt.Errorf("couldn't read argument length")
		}

		arguments := make([]byte, binary.BigEndian.Uint16(argumentsLength))

		if _, err := conn.Read(arguments); err != nil {
			return "", nil, fmt.Errorf("couldn't read arguments")
		}

		return "start", &Start{
			Type:      "start",
			Arguments: arguments,
		}, nil
	case StopID:
		return "stop", &Stop{
			Type: "stop",
		}, nil
	case AddProxyID:
		ipVersion := make([]byte, 1)

		if _, err := conn.Read(ipVersion); err != nil {
			return "", nil, fmt.Errorf("couldn't read ip version")
		}

		var ipSize uint8

		if ipVersion[0] == 4 {
			ipSize = IPv4Size
		} else if ipVersion[0] == 6 {
			ipSize = IPv6Size
		} else {
			return "", nil, fmt.Errorf("invalid IP version recieved")
		}

		ip := make(net.IP, ipSize)

		if _, err := conn.Read(ip); err != nil {
			return "", nil, fmt.Errorf("couldn't read source IP")
		}

		sourcePort := make([]byte, 2)

		if _, err := conn.Read(sourcePort); err != nil {
			return "", nil, fmt.Errorf("couldn't read source port")
		}

		destPort := make([]byte, 2)

		if _, err := conn.Read(destPort); err != nil {
			return "", nil, fmt.Errorf("couldn't read destination port")
		}

		protocolBytes := make([]byte, 1)

		if _, err := conn.Read(protocolBytes); err != nil {
			return "", nil, fmt.Errorf("couldn't read protocol")
		}

		var protocol string

		if protocolBytes[0] == TCP {
			protocol = "tcp"
		} else if protocolBytes[1] == UDP {
			protocol = "udp"
		} else {
			return "", nil, fmt.Errorf("invalid protocol")
		}

		return "addProxy", &AddProxy{
			Type:       "addProxy",
			SourceIP:   ip.String(),
			SourcePort: binary.BigEndian.Uint16(sourcePort),
			DestPort:   binary.BigEndian.Uint16(destPort),
			Protocol:   protocol,
		}, nil
	case RemoveProxyID:
		ipVersion := make([]byte, 1)

		if _, err := conn.Read(ipVersion); err != nil {
			return "", nil, fmt.Errorf("couldn't read ip version")
		}

		var ipSize uint8

		if ipVersion[0] == 4 {
			ipSize = IPv4Size
		} else if ipVersion[0] == 6 {
			ipSize = IPv6Size
		} else {
			return "", nil, fmt.Errorf("invalid IP version recieved")
		}

		ip := make(net.IP, ipSize)

		if _, err := conn.Read(ip); err != nil {
			return "", nil, fmt.Errorf("couldn't read source IP")
		}

		sourcePort := make([]byte, 2)

		if _, err := conn.Read(sourcePort); err != nil {
			return "", nil, fmt.Errorf("couldn't read source port")
		}

		destPort := make([]byte, 2)

		if _, err := conn.Read(destPort); err != nil {
			return "", nil, fmt.Errorf("couldn't read destination port")
		}

		protocolBytes := make([]byte, 1)

		if _, err := conn.Read(protocolBytes); err != nil {
			return "", nil, fmt.Errorf("couldn't read protocol")
		}

		var protocol string

		if protocolBytes[0] == TCP {
			protocol = "tcp"
		} else if protocolBytes[1] == UDP {
			protocol = "udp"
		} else {
			return "", nil, fmt.Errorf("invalid protocol")
		}

		return "removeProxy", &RemoveProxy{
			Type:       "removeProxy",
			SourceIP:   ip.String(),
			SourcePort: binary.BigEndian.Uint16(sourcePort),
			DestPort:   binary.BigEndian.Uint16(destPort),
			Protocol:   protocol,
		}, nil
	case ProxyConnectionsResponseID:
		connections := []*ProxyClientConnection{}
		delimiter := make([]byte, 1)
		var errorReturn error

		// Infinite loop because we don't know the length
		for {
			connection, err := unmarshalIndividualConnectionStruct(conn)

			if err != nil {
				return "", nil, err
			}

			connections = append(connections, connection)

			if _, err := conn.Read(delimiter); err != nil {
				return "", nil, fmt.Errorf("couldn't read delimiter")
			}

			if delimiter[0] == '\r' {
				continue
			} else if delimiter[0] == '\n' {
				break
			} else {
				// WTF? This shouldn't happen. Break out and return, but give an error
				errorReturn = fmt.Errorf("invalid delimiter recieved while processing stream")
				break
			}
		}

		return "proxyConnectionsResponse", &ProxyConnectionsResponse{
			Type:        "proxyConnectionsResponse",
			Connections: connections,
		}, errorReturn
	case CheckClientParametersID:
		ipVersion := make([]byte, 1)

		if _, err := conn.Read(ipVersion); err != nil {
			return "", nil, fmt.Errorf("couldn't read ip version")
		}

		var ipSize uint8

		if ipVersion[0] == 4 {
			ipSize = IPv4Size
		} else if ipVersion[0] == 6 {
			ipSize = IPv6Size
		} else {
			return "", nil, fmt.Errorf("invalid IP version recieved")
		}

		ip := make(net.IP, ipSize)

		if _, err := conn.Read(ip); err != nil {
			return "", nil, fmt.Errorf("couldn't read source IP")
		}

		sourcePort := make([]byte, 2)

		if _, err := conn.Read(sourcePort); err != nil {
			return "", nil, fmt.Errorf("couldn't read source port")
		}

		destPort := make([]byte, 2)

		if _, err := conn.Read(destPort); err != nil {
			return "", nil, fmt.Errorf("couldn't read destination port")
		}

		protocolBytes := make([]byte, 1)

		if _, err := conn.Read(protocolBytes); err != nil {
			return "", nil, fmt.Errorf("couldn't read protocol")
		}

		var protocol string

		if protocolBytes[0] == TCP {
			protocol = "tcp"
		} else if protocolBytes[1] == UDP {
			protocol = "udp"
		} else {
			return "", nil, fmt.Errorf("invalid protocol")
		}

		return "checkClientParameters", &CheckClientParameters{
			Type:       "checkClientParameters",
			SourceIP:   ip.String(),
			SourcePort: binary.BigEndian.Uint16(sourcePort),
			DestPort:   binary.BigEndian.Uint16(destPort),
			Protocol:   protocol,
		}, nil
	case CheckServerParametersID:
		argumentsLength := make([]byte, 2)

		if _, err := conn.Read(argumentsLength); err != nil {
			return "", nil, fmt.Errorf("couldn't read argument length")
		}

		arguments := make([]byte, binary.BigEndian.Uint16(argumentsLength))

		if _, err := conn.Read(arguments); err != nil {
			return "", nil, fmt.Errorf("couldn't read arguments")
		}

		return "checkServerParameters", &CheckServerParameters{
			Type:      "checkServerParameters",
			Arguments: arguments,
		}, nil
	case CheckParametersResponseID:
		checkMethodByte := make([]byte, 1)

		if _, err := conn.Read(checkMethodByte); err != nil {
			return "", nil, fmt.Errorf("couldn't read check method byte")
		}

		var checkMethod string

		if checkMethodByte[0] == CheckClientParametersID {
			checkMethod = "checkClientParameters"
		} else if checkMethodByte[0] == CheckServerParametersID {
			checkMethod = "checkServerParameters"
		} else {
			return "", nil, fmt.Errorf("invalid check method recieved")
		}

		isValid := make([]byte, 1)

		if _, err := conn.Read(isValid); err != nil {
			return "", nil, fmt.Errorf("couldn't read isValid byte")
		}

		messageLengthBytes := make([]byte, 2)

		if _, err := conn.Read(messageLengthBytes); err != nil {
			return "", nil, fmt.Errorf("couldn't read message length")
		}

		messageLength := binary.BigEndian.Uint16(messageLengthBytes)
		var message string

		if messageLength != 0 {
			messageBytes := make([]byte, messageLength)

			if _, err := conn.Read(messageBytes); err != nil {
				return "", nil, fmt.Errorf("couldn't read message")
			}

			message = string(messageBytes)
		}

		return "checkParametersResponse", &CheckParametersResponse{
			Type:         "checkParametersResponse",
			InResponseTo: checkMethod,
			IsValid:      isValid[0] == 1,
			Message:      message,
		}, nil
	case BackendStatusResponseID:
		isRunning := make([]byte, 1)

		if _, err := conn.Read(isRunning); err != nil {
			return "", nil, fmt.Errorf("couldn't read isRunning field")
		}

		statusCode := make([]byte, 1)

		if _, err := conn.Read(statusCode); err != nil {
			return "", nil, fmt.Errorf("couldn't read status code field")
		}

		messageLengthBytes := make([]byte, 2)

		if _, err := conn.Read(messageLengthBytes); err != nil {
			return "", nil, fmt.Errorf("couldn't read message length")
		}

		messageLength := binary.BigEndian.Uint16(messageLengthBytes)
		var message string

		if messageLength != 0 {
			messageBytes := make([]byte, messageLength)

			if _, err := conn.Read(messageBytes); err != nil {
				return "", nil, fmt.Errorf("couldn't read message")
			}

			message = string(messageBytes)
		}

		return "backendStatusResponse", &BackendStatusResponse{
			Type:       "backendStatusResponse",
			IsRunning:  isRunning[0] == 1,
			StatusCode: int(statusCode[0]),
			Message:    message,
		}, nil
	case BackendStatusRequestID:
		return "backendStatusRequest", &BackendStatusRequest{
			Type: "backendStatusRequest",
		}, nil
	case ProxyStatusRequestID:
		ipVersion := make([]byte, 1)

		if _, err := conn.Read(ipVersion); err != nil {
			return "", nil, fmt.Errorf("couldn't read ip version")
		}

		var ipSize uint8

		if ipVersion[0] == 4 {
			ipSize = IPv4Size
		} else if ipVersion[0] == 6 {
			ipSize = IPv6Size
		} else {
			return "", nil, fmt.Errorf("invalid IP version recieved")
		}

		ip := make(net.IP, ipSize)

		if _, err := conn.Read(ip); err != nil {
			return "", nil, fmt.Errorf("couldn't read source IP")
		}

		sourcePort := make([]byte, 2)

		if _, err := conn.Read(sourcePort); err != nil {
			return "", nil, fmt.Errorf("couldn't read source port")
		}

		destPort := make([]byte, 2)

		if _, err := conn.Read(destPort); err != nil {
			return "", nil, fmt.Errorf("couldn't read destination port")
		}

		protocolBytes := make([]byte, 1)

		if _, err := conn.Read(protocolBytes); err != nil {
			return "", nil, fmt.Errorf("couldn't read protocol")
		}

		var protocol string

		if protocolBytes[0] == TCP {
			protocol = "tcp"
		} else if protocolBytes[1] == UDP {
			protocol = "udp"
		} else {
			return "", nil, fmt.Errorf("invalid protocol")
		}

		return "proxyStatusRequest", &ProxyStatusRequest{
			Type:       "proxyStatusRequest",
			SourceIP:   ip.String(),
			SourcePort: binary.BigEndian.Uint16(sourcePort),
			DestPort:   binary.BigEndian.Uint16(destPort),
			Protocol:   protocol,
		}, nil
	case ProxyStatusResponseID:
		ipVersion := make([]byte, 1)

		if _, err := conn.Read(ipVersion); err != nil {
			return "", nil, fmt.Errorf("couldn't read ip version")
		}

		var ipSize uint8

		if ipVersion[0] == 4 {
			ipSize = IPv4Size
		} else if ipVersion[0] == 6 {
			ipSize = IPv6Size
		} else {
			return "", nil, fmt.Errorf("invalid IP version recieved")
		}

		ip := make(net.IP, ipSize)

		if _, err := conn.Read(ip); err != nil {
			return "", nil, fmt.Errorf("couldn't read source IP")
		}

		sourcePort := make([]byte, 2)

		if _, err := conn.Read(sourcePort); err != nil {
			return "", nil, fmt.Errorf("couldn't read source port")
		}

		destPort := make([]byte, 2)

		if _, err := conn.Read(destPort); err != nil {
			return "", nil, fmt.Errorf("couldn't read destination port")
		}

		protocolBytes := make([]byte, 1)

		if _, err := conn.Read(protocolBytes); err != nil {
			return "", nil, fmt.Errorf("couldn't read protocol")
		}

		var protocol string

		if protocolBytes[0] == TCP {
			protocol = "tcp"
		} else if protocolBytes[1] == UDP {
			protocol = "udp"
		} else {
			return "", nil, fmt.Errorf("invalid protocol")
		}

		isActive := make([]byte, 1)

		if _, err := conn.Read(isActive); err != nil {
			return "", nil, fmt.Errorf("couldn't read isActive field")
		}

		return "proxyStatusResponse", &ProxyStatusResponse{
			Type:       "proxyStatusResponse",
			SourceIP:   ip.String(),
			SourcePort: binary.BigEndian.Uint16(sourcePort),
			DestPort:   binary.BigEndian.Uint16(destPort),
			Protocol:   protocol,
			IsActive:   isActive[0] == 1,
		}, nil
	case ProxyInstanceRequestID:
		return "proxyInstanceRequest", &ProxyInstanceRequest{
			Type: "proxyInstanceRequest",
		}, nil
	case ProxyInstanceResponseID:
		proxies := []*ProxyInstance{}
		delimiter := make([]byte, 1)
		var errorReturn error

		// Infinite loop because we don't know the length
		for {
			proxy, err := unmarshalIndividualProxyStruct(conn)

			if err != nil {
				return "", nil, err
			}

			proxies = append(proxies, proxy)

			if _, err := conn.Read(delimiter); err != nil {
				return "", nil, fmt.Errorf("couldn't read delimiter")
			}

			if delimiter[0] == '\r' {
				continue
			} else if delimiter[0] == '\n' {
				break
			} else {
				// WTF? This shouldn't happen. Break out and return, but give an error
				errorReturn = fmt.Errorf("invalid delimiter recieved while processing stream")
				break
			}
		}

		return "proxyInstanceResponse", &ProxyInstanceResponse{
			Type:    "proxyInstanceResponse",
			Proxies: proxies,
		}, errorReturn
	case ProxyConnectionsRequestID:
		return "proxyConnectionsRequest", &ProxyConnectionsRequest{
			Type: "proxyConnectionsRequest",
		}, nil
	}

	return "", nil, fmt.Errorf("couldn't match command ID")
}
