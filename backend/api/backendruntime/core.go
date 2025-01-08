package backendruntime

import "os"

var (
	AvailableBackends []*Backend
	RunningBackends   map[uint]*Runtime
	TempDir           string
	isDevelopmentMode bool
)

func init() {
	RunningBackends = make(map[uint]*Runtime)
	isDevelopmentMode = os.Getenv("HERMES_DEVELOPMENT_MODE") != ""
}
