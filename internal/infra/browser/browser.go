package browser

import (
	"os/exec"
	"runtime"
)

// Open opens the given URL in the default system browser.
// It returns an error only if the open command itself fails to start;
// callers should always print the URL regardless.
func Open(url string) error {
	var cmd string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "linux":
		cmd = "xdg-open"
	case "windows":
		cmd = "start"
	default:
		return nil
	}
	return exec.Command(cmd, url).Start()
}
