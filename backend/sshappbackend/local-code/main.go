package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"git.terah.dev/imterah/hermes/backend/backendutil"
	"git.terah.dev/imterah/hermes/backend/commonbackend"
	"github.com/charmbracelet/log"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SSHAppBackendData struct {
	IP          string   `json:"ip" validate:"required"`
	Port        uint16   `json:"port" validate:"required"`
	Username    string   `json:"username" validate:"required"`
	PrivateKey  string   `json:"privateKey" validate:"required"`
	ListenOnIPs []string `json:"listenOnIPs"`
}

type SSHAppBackend struct {
	config         *SSHAppBackendData
	conn           *ssh.Client
	clients        []*commonbackend.ProxyClientConnection
	arrayPropMutex sync.Mutex
}

func (backend *SSHAppBackend) StartBackend(configBytes []byte) (bool, error) {
	log.Info("SSHAppBackend is initializing...")
	var backendData SSHAppBackendData

	if err := json.Unmarshal(configBytes, &backendData); err != nil {
		return false, err
	}

	if err := validator.New().Struct(&backendData); err != nil {
		return false, err
	}

	backend.config = &backendData

	if len(backend.config.ListenOnIPs) == 0 {
		backend.config.ListenOnIPs = []string{"0.0.0.0"}
	}

	signer, err := ssh.ParsePrivateKey([]byte(backendData.PrivateKey))

	if err != nil {
		log.Warnf("Failed to initialize: %s", err.Error())
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
		log.Warnf("Failed to initialize: %s", err.Error())
		return false, err
	}

	backend.conn = conn

	log.Info("SSHAppBackend has connected successfully.")
	log.Info("Getting CPU architecture...")

	session, err := backend.conn.NewSession()

	if err != nil {
		log.Warnf("Failed to create session: %s", err.Error())
		conn.Close()
		backend.conn = nil
		return false, err
	}

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf

	err = session.Run("uname -m")

	if err != nil {
		log.Warnf("Failed to run uname command: %s", err.Error())
		conn.Close()
		backend.conn = nil
		return false, err
	}

	cpuArchBytes := make([]byte, stdoutBuf.Len())
	stdoutBuf.Read(cpuArchBytes)

	cpuArch := string(cpuArchBytes)
	cpuArch = cpuArch[:len(cpuArch)-1]

	var backendBinary string

	// Ordered in (subjective) popularity
	if cpuArch == "x86_64" {
		backendBinary = "remote-bin/rt-amd64"
	} else if cpuArch == "aarch64" {
		backendBinary = "remote-bin/rt-arm64"
	} else if cpuArch == "arm" {
		backendBinary = "remote-bin/rt-arm"
	} else if len(cpuArch) == 4 && string(cpuArch[0]) == "i" && strings.HasSuffix(cpuArch, "86") {
		backendBinary = "remote-bin/rt-386"
	} else {
		log.Warn("Failed to determine executable to use: CPU architecture not compiled/supported currently")
		conn.Close()
		backend.conn = nil
		return false, fmt.Errorf("CPU architecture not compiled/supported currently")
	}

	log.Info("Checking if we need to copy the application...")

	var binary []byte
	needsToCopyBinary := true

	session, err = backend.conn.NewSession()

	if err != nil {
		log.Warnf("Failed to create session: %s", err.Error())
		conn.Close()
		backend.conn = nil
		return false, err
	}

	session.Stdout = &stdoutBuf

	err = session.Start("[ -f /tmp/sshappbackend.runtime ] && md5sum /tmp/sshappbackend.runtime | cut -d \" \" -f 1")

	if err != nil {
		log.Warnf("Failed to calculate hash of possibly existing backend: %s", err.Error())
		conn.Close()
		backend.conn = nil
		return false, err
	}

	fileExists := stdoutBuf.Len() != 0

	if fileExists {
		remoteMD5HashStringBuf := make([]byte, stdoutBuf.Len())
		stdoutBuf.Read(remoteMD5HashStringBuf)

		remoteMD5HashString := string(remoteMD5HashStringBuf)
		remoteMD5HashString = remoteMD5HashString[:len(remoteMD5HashString)-1]

		remoteMD5Hash, err := hex.DecodeString(remoteMD5HashString)

		if err != nil {
			log.Warnf("Failed to decode hex: %s", err.Error())
			conn.Close()
			backend.conn = nil
			return false, err
		}

		binary, err = binFiles.ReadFile(backendBinary)

		if err != nil {
			log.Warnf("Failed to read file in the embedded FS: %s", err.Error())
			conn.Close()
			backend.conn = nil
			return false, fmt.Errorf("(embedded FS): %s", err.Error())
		}

		localMD5Hash := md5.Sum(binary)

		log.Infof("remote: %s, local: %s", remoteMD5HashString, hex.EncodeToString(localMD5Hash[:]))

		if bytes.Compare(localMD5Hash[:], remoteMD5Hash) == 0 {
			needsToCopyBinary = false
		}
	}

	if needsToCopyBinary {
		log.Info("Copying binary...")
		sftpInstance, err := sftp.NewClient(conn)

		if err != nil {
			log.Warnf("Failed to initialize SFTP: %s", err.Error())
			conn.Close()
			backend.conn = nil
			return false, err
		}

		defer sftpInstance.Close()

		if len(binary) == 0 {
			binary, err = binFiles.ReadFile(backendBinary)

			if err != nil {
				log.Warnf("Failed to read file in the embedded FS: %s", err.Error())
				conn.Close()
				backend.conn = nil
				return false, fmt.Errorf("(embedded FS): %s", err.Error())
			}
		}

		var file *sftp.File

		if fileExists {
			file, err = sftpInstance.Create("/tmp/sshappbackend.runtime")
		} else {
			file, err = sftpInstance.OpenFile("/tmp/sshappbackend.runtime", os.O_WRONLY)
		}

		if err != nil {
			log.Warnf("Failed to create file: %s", err.Error())
			conn.Close()
			backend.conn = nil
			return false, err
		}

		_, err = file.Write(binary)

		if err != nil {
			log.Warnf("Failed to write file: %s", err.Error())
			conn.Close()
			backend.conn = nil
			return false, err
		}

		err = file.Chmod(775)

		if err != nil {
			log.Warnf("Failed to change permissions on file: %s", err.Error())
			conn.Close()
			backend.conn = nil
			return false, err
		}

		log.Info("Done copying file.")
	} else {
		log.Info("Skipping copying as there's a copy on disk already.")
	}

	log.Info("Starting process...")

	session, err = backend.conn.NewSession()

	if err != nil {
		log.Warnf("Failed to create session: %s", err.Error())
		conn.Close()
		backend.conn = nil
		return false, err
	}

	session.Stdout = WriteLogger{}
	session.Stderr = WriteLogger{}

	go session.Run("/tmp/sshappbackend.runtime")
	log.Info("SSHAppBackend has initialized successfully.")

	return true, nil
}

func (backend *SSHAppBackend) StopBackend() (bool, error) {
	err := backend.conn.Close()

	if err != nil {
		return false, err
	}

	return true, nil
}

func (backend *SSHAppBackend) GetBackendStatus() (bool, error) {
	return backend.conn != nil, nil
}

func (backend *SSHAppBackend) StartProxy(command *commonbackend.AddProxy) (bool, error) {
	return true, nil
}

func (backend *SSHAppBackend) StopProxy(command *commonbackend.RemoveProxy) (bool, error) {
	return false, fmt.Errorf("could not find the proxy")
}

func (backend *SSHAppBackend) GetAllClientConnections() []*commonbackend.ProxyClientConnection {
	return backend.clients
}

func (backend *SSHAppBackend) CheckParametersForConnections(clientParameters *commonbackend.CheckClientParameters) *commonbackend.CheckParametersResponse {
	return &commonbackend.CheckParametersResponse{
		IsValid: true,
	}
}

func (backend *SSHAppBackend) CheckParametersForBackend(arguments []byte) *commonbackend.CheckParametersResponse {
	var backendData SSHAppBackendData

	if err := json.Unmarshal(arguments, &backendData); err != nil {
		return &commonbackend.CheckParametersResponse{
			IsValid: false,
			Message: fmt.Sprintf("could not read json: %s", err.Error()),
		}
	}

	if err := validator.New().Struct(&backendData); err != nil {
		return &commonbackend.CheckParametersResponse{
			IsValid: false,
			Message: fmt.Sprintf("failed validation of parameters: %s", err.Error()),
		}
	}

	return &commonbackend.CheckParametersResponse{
		IsValid: true,
	}
}

func main() {
	logLevel := os.Getenv("HERMES_LOG_LEVEL")

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

	backend := &SSHAppBackend{}

	application := backendutil.NewHelper(backend)
	err := application.Start()

	if err != nil {
		log.Fatalf("failed execution in application: %s", err.Error())
	}
}
