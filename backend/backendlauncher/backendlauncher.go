package backendlauncher

import (
	"fmt"
	"math/rand/v2"
	"net"
)

func GetUnixSocket(folder string) (string, net.Listener, error) {
	socketPath := fmt.Sprintf("%s/sock-%d.sock", folder, rand.Uint())
	listener, err := net.Listen("unix", socketPath)

	if err != nil {
		return "", nil, err
	}

	return socketPath, listener, nil
}
