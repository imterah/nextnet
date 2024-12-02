package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"git.greysoh.dev/imterah/nextnet/backendutil"
	"git.greysoh.dev/imterah/nextnet/commonbackend"
	"github.com/charmbracelet/log"
	"golang.org/x/crypto/ssh"
)

type SSHBackend struct {
	data    SSHBackendData
	conn    ssh.Client
	clients []*commonbackend.ClientConnection
}

type SSHBackendData struct {
	Ip          string   `json:"ip"`
	Port        uint16   `json:"port"`
	Username    string   `json:"username"`
	PrivateKey  string   `json:"privateKey"`
	ListenOnIPs []string `json:"listenOnIPs"`
}

func (backend *SSHBackend) StartBackend(bytes []byte) (bool, error) {
	var backendData SSHBackendData

	err := json.Unmarshal(bytes, &backendData) // ?????
	if err != nil {
		return false, err
	}
	backend.data = backendData

	if len(backend.data.ListenOnIPs) == 0 {
		backend.data.ListenOnIPs = []string{"0.0.0.0"}
	}

	// create signer for privateKey
	signer, err := ssh.ParsePrivateKey([]byte(backendData.PrivateKey))
	if err != nil {
		return false, err
	}

	auth := ssh.PublicKeys(signer)

	config := &ssh.ClientConfig{
		User: backendData.Username,
		Auth: []ssh.AuthMethod{
			auth,
		},
	}

	conn, err := ssh.Dial("tcp", backendData.Ip+":"+string(backendData.Port), config)
	if err != nil {
		return false, err
	}
	backend.conn = conn

	return true, nil
}

func (backend *SSHBackend) StopBackend() (bool, error) {
	err := backend.conn.Close()
	if err != nil {
		return false, err
	}

	return true, nil
}

func (backend *SSHBackend) AddConnection(command *commonbackend.AddConnectionCommand) (bool, error) {
	for _, ipListener := range backend.data.ListenOnIPs {
		ip := net.TCPAddr{
			IP:   net.ParseIP(ipListener),
			Port: int(command.DestPort),
		}
		listener, err := backend.conn.ListenTCP(&ip)
		if err != nil {
			return false, err
		}
		go func() {
			for {
				forwardedConn, err := listener.Accept()
				if err != nil {
					log.Warnf("failed to accept listener connection: %s", err.Error())
					continue
				}
				sourceConn, err := net.Dial("tcp", command.SourceIP+":"+string(command.SourcePort))
				if err != nil {
					log.Warnf("failed to dial source connection: %s", err.Error())
					continue
				}
				sourceBuffer := make([]byte, 65535)
				forwardedBuffer := make([]byte, 65535)
				go func() {
					defer sourceConn.Close()
					defer forwardedConn.Close()

					for {
						len, err := forwardedConn.Read(forwardedBuffer)
						if err != nil {
							log.Errorf("failed to read from forwarded connection: %s", err.Error())
							return
						}

						_, err = sourceConn.Write(forwardedBuffer[:len])
						if err != nil {
							log.Errorf("failed to write to source connection: %s", err.Error())
							return
						}
					}
				}()
				go func() {
					defer sourceConn.Close()
					defer forwardedConn.Close()

					for {
						len, err := sourceConn.Read(sourceBuffer)
						if err != nil && err.Error() != "EOF" && strings.HasSuffix(err.Error(), "use of closed network connection") {
							log.Errorf("failed to read from source connection: %s", err.Error())
							return
						}

						_, err = forwardedConn.Write(sourceBuffer[:len])
						if err != nil && err.Error() != "EOF" && strings.HasSuffix(err.Error(), "use of closed network connection") {
							log.Errorf("failed to write to forwarded connection: %s", err.Error())
							return
						}
					}
				}()
			}
		}()
	}

	return true, nil
}

func (backend *SSHBackend) RemoveConnection(command *commonbackend.RemoveConnectionCommand) (bool, error) {
	// FIXME: implement
	return true, nil
}

func (backend *SSHBackend) GetAllConnections() []*commonbackend.ClientConnection {
	// return []*commonbackend.ClientConnection{}
	return backend.clients
}

func (backend *SSHBackend) CheckParametersForConnections(clientParameters *commonbackend.CheckClientParameters) *commonbackend.CheckParametersResponse {
	if clientParameters.Protocol != "tcp" {
		return &commonbackend.CheckParametersResponse{
			IsValid: false,
			Message: "Only TCP is supported",
		}
	}

	return &commonbackend.CheckParametersResponse{
		IsValid: true,
	}
}

func (backend *SSHBackend) CheckParametersForBackend(arguments []byte) *commonbackend.CheckParametersResponse {
	var backendData SSHBackendData

	err := json.Unmarshal(arguments, &backendData) // ?????
	if err != nil {
		return &commonbackend.CheckParametersResponse{
			IsValid: false,
			Message: fmt.Sprintf("could not read json: %s", err.Error()),
		}
	}

	return &commonbackend.CheckParametersResponse{
		IsValid: true,
	}
}

func main() {
	// When using logging, you should use charmbracelet/log, because that's what everything else uses in this ecosystem of a project. - imterah
	logLevel := os.Getenv("NEXTNET_LOG_LEVEL")

	if logLevel != "" {
		switch logLevel {
		case "debug":
			log.SetLevel(log.DebugLevel)

		case "info":
			log.SetLevel(log.InfoLevel)

		case "warn":
			log.SetLevel(log.WarnLevel)

		case "error":
			log.SetLevel(log.ErrorLevel)

		case "fatal":
			log.SetLevel(log.FatalLevel)
		}
	}

	backend := &SSHBackend{}

	application := backendutil.NewHelper(backend)
	err := application.Start()

	if err != nil {
		log.Fatalf("failed execution in application: %s", err.Error())
	}
}
