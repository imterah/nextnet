//go:build debug

package backendutil

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"golang.org/x/exp/rand"
)

func configureAndLaunchBackgroundProfilingTasks() error {
	profilingMode, err := os.ReadFile("/tmp/hermes.backendlauncher.profilebackends")

	if err != nil && errors.Is(err, os.ErrNotExist) {
		return nil
	}

	switch string(profilingMode) {
	case "cpu":
		log.Debug("Starting CPU profiling as a background task")
		go doCPUProfiling()
	case "mem":
		log.Debug("Starting memory profiling as a background task")
		go doMemoryProfiling()
	default:
		log.Warnf("Unknown profiling mode: %s", string(profilingMode))
		return nil
	}

	return nil
}

func doCPUProfiling() {
	// (imterah) WTF? why isn't this being seeded on its own? according to Go docs, this should be seeded automatically...
	rand.Seed(uint64(time.Now().UnixNano()))

	profileFileName := fmt.Sprintf("/tmp/hermes.backendlauncher.cpu.prof.%d", rand.Int())
	profileFile, err := os.Create(profileFileName)

	if err != nil {
		log.Fatalf("Failed to create CPU profiling file: %s", err.Error())
	}

	log.Debugf("Writing CPU usage profile to '%s'. Will capture when Ctrl+C/SIGTERM is recieved.", profileFileName)
	pprof.StartCPUProfile(profileFile)

	exitNotification := make(chan os.Signal, 1)
	signal.Notify(exitNotification, os.Interrupt, syscall.SIGTERM)
	<-exitNotification

	log.Debug("Recieved SIGTERM. Cleaning up and exiting...")

	pprof.StopCPUProfile()
	profileFile.Close()

	log.Debug("Exiting...")
	os.Exit(0)
}

func doMemoryProfiling() {
	// (imterah) WTF? why isn't this being seeded on its own? according to Go docs, this should be seeded automatically...
	rand.Seed(uint64(time.Now().UnixNano()))

	profileFileName := fmt.Sprintf("/tmp/hermes.backendlauncher.mem.prof.%d", rand.Int())
	profileFile, err := os.Create(profileFileName)

	if err != nil {
		log.Fatalf("Failed to create memory profiling file: %s", err.Error())
	}

	log.Debugf("Writing memory profile to '%s'. Will capture when Ctrl+C/SIGTERM is recieved.", profileFileName)

	exitNotification := make(chan os.Signal, 1)
	signal.Notify(exitNotification, os.Interrupt, syscall.SIGTERM)
	<-exitNotification

	log.Debug("Recieved SIGTERM. Cleaning up and exiting...")

	pprof.WriteHeapProfile(profileFile)
	profileFile.Close()

	log.Debug("Exiting...")
	os.Exit(0)
}
