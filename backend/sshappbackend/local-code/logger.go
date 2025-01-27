package main

import (
	"strings"

	"github.com/charmbracelet/log"
)

type WriteLogger struct{}

func (writer WriteLogger) Write(p []byte) (n int, err error) {
	logSplit := strings.Split(string(p), "\n")

	for _, line := range logSplit {
		if line == "" {
			continue
		}

		log.Infof("Process: %s", line)
	}

	return len(p), err
}
