package backendutil

import (
	"fmt"
	"net"
	"os"

	"git.greysoh.dev/imterah/nextnet/commonbackend"
	"github.com/charmbracelet/log"
)

type BackendApplicationHelper struct {
	Backend    BackendInterface
	SocketPath string

	socket net.Conn
}

func (helper *BackendApplicationHelper) Start() error {
	var err error
	helper.socket, err = net.Dial("unix", helper.SocketPath)

	if err != nil {
		return err
	}

	for {
		commandType, commandRaw, err := commonbackend.Unmarshal(helper.socket)

		if err != nil {
			return err
		}

		switch commandType {
		case "start":
			// TODO: implement response logic
			command, ok := commandRaw.(*commonbackend.StartCommand)

			if !ok {
				return fmt.Errorf("failed to typecast")
			}

			// ok, err :=
			_, _ = helper.Backend.StartBackend(command.Arguments)
		case "stop":
			// TODO: implement response logic
			_, ok := commandRaw.(*commonbackend.StopCommand)

			if !ok {
				return fmt.Errorf("failed to typecast")
			}

			_, _ = helper.Backend.StopBackend()
		case "addConnection":
			// TODO: implement response logic
			command, ok := commandRaw.(*commonbackend.AddConnectionCommand)

			if !ok {
				return fmt.Errorf("failed to typecast")
			}

			_, _ = helper.Backend.StartProxy(command)
		case "removeConnection":
			// TODO: implement response logic
			command, ok := commandRaw.(*commonbackend.RemoveConnectionCommand)

			if !ok {
				return fmt.Errorf("failed to typecast")
			}

			_, _ = helper.Backend.StopProxy(command)
		case "getAllConnections":
			_, ok := commandRaw.(*commonbackend.AddConnectionCommand)

			if !ok {
				return fmt.Errorf("failed to typecast")
			}

			connections := helper.Backend.GetAllClientConnections()

			serverParams := &commonbackend.ConnectionsResponse{
				Type:        "connectionsResponse",
				Connections: connections,
			}

			byteData, err := commonbackend.Marshal(serverParams.Type, serverParams)

			if err != nil {
				return err
			}

			if _, err = helper.socket.Write(byteData); err != nil {
				return err
			}
		case "checkClientParameters":
			// TODO: implement response logic
			command, ok := commandRaw.(*commonbackend.CheckClientParameters)

			if !ok {
				return fmt.Errorf("failed to typecast")
			}

			resp := helper.Backend.CheckParametersForConnections(command)
			resp.Type = "checkParametersResponse"
			resp.InResponseTo = "checkClientParameters"

			byteData, err := commonbackend.Marshal(resp.Type, resp)

			if err != nil {
				return err
			}

			if _, err = helper.socket.Write(byteData); err != nil {
				return err
			}
		case "checkServerParameters":
			// TODO: implement response logic
			command, ok := commandRaw.(*commonbackend.CheckServerParameters)

			if !ok {
				return fmt.Errorf("failed to typecast")
			}

			resp := helper.Backend.CheckParametersForBackend(command.Arguments)
			resp.Type = "checkParametersResponse"
			resp.InResponseTo = "checkServerParameters"

			byteData, err := commonbackend.Marshal(resp.Type, resp)

			if err != nil {
				return err
			}

			if _, err = helper.socket.Write(byteData); err != nil {
				return err
			}
		}
	}
}

func NewHelper(backend BackendInterface) *BackendApplicationHelper {
	socketPath, ok := os.LookupEnv("NEXTNET_API_SOCK")

	if !ok {
		log.Warn("NEXTNET_API_SOCK is not defined! This will cause an issue unless the backend manually overwrites it")
	}

	helper := &BackendApplicationHelper{
		Backend:    backend,
		SocketPath: socketPath,
	}

	return helper
}
