//go:build !debug

package backendutil

var endProfileFunc func()

func configureAndLaunchBackgroundProfilingTasks() error {
	return nil
}
