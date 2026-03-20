package docker

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CheckDocker verifies that the docker binary is available and the daemon is running.
func CheckDocker() error {
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker is not installed. install from https://docs.docker.com/get-docker/")
	}
	if err := exec.Command("docker", "info").Run(); err != nil {
		return fmt.Errorf("docker daemon is not running. start Docker and try again")
	}
	return nil
}

// Client shells out to `docker compose` for the given compose file.
type Client struct {
	composePath string
}

// NewClient creates a Client bound to the given compose file path.
func NewClient(composePath string) *Client {
	return &Client{composePath: composePath}
}

// Up runs `docker compose up -d`, streaming output to the user.
func (c *Client) Up() error {
	return c.stream("up", "-d")
}

// Down runs `docker compose down`, streaming output to the user.
func (c *Client) Down() error {
	return c.stream("down")
}

// DownVolumes runs `docker compose down --volumes`, streaming output to the user.
func (c *Client) DownVolumes() error {
	return c.stream("down", "--volumes")
}

// Pull runs `docker compose pull`, streaming output to the user.
func (c *Client) Pull() error {
	return c.stream("pull")
}

// Logs runs `docker compose logs`, streaming output.
// If follow is true, passes --follow. If service is non-empty, filters to that service.
func (c *Client) Logs(follow bool, service string) error {
	args := []string{"logs"}
	if follow {
		args = append(args, "--follow")
	}
	if service != "" {
		args = append(args, service)
	}
	return c.stream(args...)
}

// IsRunning returns true if any containers are currently running for this compose project.
func (c *Client) IsRunning() (bool, error) {
	cmd := exec.Command("docker", "compose", "-f", c.composePath, "ps", "-q")
	out, err := cmd.Output()
	if err != nil {
		return false, nil
	}
	return len(strings.TrimSpace(string(out))) > 0, nil
}

// IsServiceRunning returns true if the named service has a running container.
func (c *Client) IsServiceRunning(service string) bool {
	cmd := exec.Command("docker", "compose", "-f", c.composePath, "ps", "-q", service)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(out))) > 0
}

// DownSilent runs `docker compose down` capturing output (not streaming to user).
func (c *Client) DownSilent() error {
	cmd := exec.Command("docker", "compose", "-f", c.composePath, "down")
	cmd.CombinedOutput() //nolint:errcheck
	return nil
}

// stream executes a `docker compose -f <path> <args...>` command, wiring
// stdout and stderr directly to the process so output streams to the user.
func (c *Client) stream(args ...string) error {
	fullArgs := append([]string{"compose", "-f", c.composePath}, args...)
	cmd := exec.Command("docker", fullArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
