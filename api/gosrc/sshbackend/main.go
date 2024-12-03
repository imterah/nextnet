package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"git.greysoh.dev/imterah/nextnet/backendutil"
	"git.greysoh.dev/imterah/nextnet/commonbackend"
	"github.com/charmbracelet/log"
	"golang.org/x/crypto/ssh"
)

type SSHListener struct {
	SourceIP   string
	SourcePort uint16
	DestPort   uint16
	Protocol   string // Will be either 'tcp' or 'udp'
	Listeners  []net.Listener
}

type SSHBackend struct {
	config         SSHBackendData
	conn           *ssh.Client
	clients        []*commonbackend.ProxyClientConnection
	proxies        []*SSHListener
	arrayPropMutex sync.Mutex
}

type SSHBackendData struct {
	IP          string   `json:"ip"`
	Port        uint16   `json:"port"`
	Username    string   `json:"username"`
	PrivateKey  string   `json:"privateKey"`
	ListenOnIPs []string `json:"listenOnIPs"`
}

func (backend *SSHBackend) StartBackend(bytes []byte) (bool, error) {
	log.Info("SSHBackend is initializing...")
	var backendData SSHBackendData

	err := json.Unmarshal(bytes, &backendData)

	if err != nil {
		return false, err
	}

	backend.config = backendData

	if len(backend.config.ListenOnIPs) == 0 {
		backend.config.ListenOnIPs = []string{"0.0.0.0"}
	}

	signer, err := ssh.ParsePrivateKey([]byte(backendData.PrivateKey))

	if err != nil {
		return false, err
	}

	auth := ssh.PublicKeys(signer)

	config := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		User:            backendData.Username,
		Auth: []ssh.AuthMethod{
			auth,
		},
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", backendData.IP, backendData.Port), config)

	if err != nil {
		return false, err
	}

	backend.conn = conn

	log.Info("SSHBackend has initialized successfully.")
	return true, nil
}

func (backend *SSHBackend) StopBackend() (bool, error) {
	err := backend.conn.Close()

	if err != nil {
		return false, err
	}

	return true, nil
}

func (backend *SSHBackend) StartProxy(command *commonbackend.AddProxy) (bool, error) {
	listenerObject := &SSHListener{
		SourceIP:   command.SourceIP,
		SourcePort: command.SourcePort,
		DestPort:   command.DestPort,
		Protocol:   command.Protocol,
		Listeners:  []net.Listener{},
	}

	for _, ipListener := range backend.config.ListenOnIPs {
		ip := net.TCPAddr{
			IP:   net.ParseIP(ipListener),
			Port: int(command.DestPort),
		}

		listener, err := backend.conn.ListenTCP(&ip)

		if err != nil {
			// Incase we error out, we clean up all the other listeners
			for _, listener := range listenerObject.Listeners {
				err = listener.Close()

				if err != nil {
					log.Warnf("failed to close listener upon failure cleanup: %s", err.Error())
				}
			}

			return false, err
		}

		listenerObject.Listeners = append(listenerObject.Listeners, listener)

		go func() {
			for {
				forwardedConn, err := listener.Accept()

				if err != nil {
					log.Warnf("failed to accept listener connection: %s", err.Error())
					continue
				}

				sourceConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", command.SourceIP, command.SourcePort))

				if err != nil {
					log.Warnf("failed to dial source connection: %s", err.Error())
					continue
				}

				clientIPAndPort := forwardedConn.RemoteAddr().String()
				clientIP := clientIPAndPort[:strings.LastIndex(clientIPAndPort, ":")]
				clientPort, err := strconv.Atoi(clientIPAndPort[strings.LastIndex(clientIPAndPort, ":")+1:])

				if err != nil {
					log.Warnf("failed to parse client port: %s", err.Error())
					continue
				}

				advertisedConn := &commonbackend.ProxyClientConnection{
					SourceIP:   command.SourceIP,
					SourcePort: command.SourcePort,
					DestPort:   command.DestPort,
					ClientIP:   clientIP,
					ClientPort: uint16(clientPort),

					// FIXME (imterah): shouldn't protocol be in here?
					// Protocol:   command.Protocol,
				}

				backend.arrayPropMutex.Lock()
				backend.clients = append(backend.clients, advertisedConn)
				backend.arrayPropMutex.Unlock()

				cleanupJob := func() {
					defer backend.arrayPropMutex.Unlock()
					err := sourceConn.Close()

					if err != nil {
						log.Warnf("failed to close source connection: %s", err.Error())
					}

					err = forwardedConn.Close()

					if err != nil {
						log.Warnf("failed to close forwarded/proxied connection: %s", err.Error())
					}

					backend.arrayPropMutex.Lock()

					for clientIndex, clientInstance := range backend.clients {
						// Check if memory addresses are equal for the pointer
						if clientInstance == advertisedConn {
							// Splice out the clientInstance by clientIndex

							// TODO: change approach. It works but it's a bit wonky imho
							// I asked AI to do this as it's a relatively simple task and I forgot how to do this effectively
							backend.clients = append(backend.clients[:clientIndex], backend.clients[clientIndex+1:]...)
							return
						}
					}

					log.Warn("failed to delete client from clients metadata: couldn't find client in the array")
				}

				sourceBuffer := make([]byte, 65535)
				forwardedBuffer := make([]byte, 65535)

				go func() {
					defer cleanupJob()

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
					defer cleanupJob()

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

	backend.arrayPropMutex.Lock()
	backend.proxies = append(backend.proxies, listenerObject)
	backend.arrayPropMutex.Unlock()

	return true, nil
}

func (backend *SSHBackend) StopProxy(command *commonbackend.RemoveProxy) (bool, error) {
	defer backend.arrayPropMutex.Unlock()
	backend.arrayPropMutex.Lock()

	for proxyIndex, proxy := range backend.proxies {
		// Check if memory addresses are equal for the pointer
		if command.SourceIP == proxy.SourceIP && command.SourcePort == proxy.SourcePort && command.DestPort == proxy.DestPort && command.Protocol == proxy.Protocol {
			log.Debug("found proxy in StopProxy. shutting down listeners")

			for _, listener := range proxy.Listeners {
				err := listener.Close()

				if err != nil {
					log.Warnf("failed to stop listener in StopProxy: %s", err.Error())
				}
			}

			// Splice out the proxy instance by proxyIndex

			// TODO: change approach. It works but it's a bit wonky imho
			// I asked AI to do this as it's a relatively simple task and I forgot how to do this effectively
			backend.proxies = append(backend.proxies[:proxyIndex], backend.proxies[proxyIndex+1:]...)
			return true, nil
		}
	}

	return false, fmt.Errorf("could not find the proxy")
}

func (backend *SSHBackend) GetAllClientConnections() []*commonbackend.ProxyClientConnection {
	defer backend.arrayPropMutex.Unlock()
	backend.arrayPropMutex.Lock()

	return backend.clients
}

func (backend *SSHBackend) CheckParametersForConnections(clientParameters *commonbackend.CheckClientParameters) *commonbackend.CheckParametersResponse {
	if clientParameters.Protocol != "tcp" {
		return &commonbackend.CheckParametersResponse{
			IsValid: false,
			Message: "Only TCP is supported for SSH",
		}
	}

	return &commonbackend.CheckParametersResponse{
		IsValid: true,
	}
}

func (backend *SSHBackend) CheckParametersForBackend(arguments []byte) *commonbackend.CheckParametersResponse {
	var backendData SSHBackendData

	if err := json.Unmarshal(arguments, &backendData); err != nil {
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
