package transport

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ContainerTransport talks to a docker or podman container. The two CLIs
// share the same exec/cp surface, so one implementation covers both.
type ContainerTransport struct {
	Bin       string // "docker" or "podman"
	Container string
}

func NewContainerTransport(bin, container string) *ContainerTransport {
	return &ContainerTransport{Bin: bin, Container: container}
}

func (c *ContainerTransport) RunInteractive(env map[string]string, args ...string) error {
	cmdArgs := []string{"exec", "-it"}
	for k, v := range env {
		cmdArgs = append(cmdArgs, "-e", fmt.Sprintf("%s=%s", k, v))
	}
	cmdArgs = append(cmdArgs, c.Container)
	cmdArgs = append(cmdArgs, args...)
	cmd := exec.Command(c.Bin, cmdArgs...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

func (c *ContainerTransport) Run(script string) (string, error) {
	return runCapture(exec.Command(c.Bin, "exec", c.Container, "sh", "-c", script))
}

func (c *ContainerTransport) CopyTo(localPath, remotePath string) error {
	if err := c.MkdirAll(parentDir(remotePath)); err != nil {
		return fmt.Errorf("mkdir -p %s: %w", parentDir(remotePath), err)
	}
	cmd := exec.Command(c.Bin, "cp", localPath, c.Container+":"+remotePath)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func (c *ContainerTransport) MkdirAll(remotePath string) error {
	_, err := c.Run("mkdir -p " + shQuote(remotePath))
	return err
}

func (c *ContainerTransport) Home() (string, error) {
	out, err := c.Run("echo $HOME")
	return strings.TrimSpace(out), err
}

func (c *ContainerTransport) Uname() (string, string, error) {
	out, err := c.Run("uname -s && uname -m")
	if err != nil {
		return "", "", err
	}
	f := strings.Fields(out)
	if len(f) < 2 {
		return "", "", fmt.Errorf("unexpected uname output: %q", out)
	}
	return f[0], f[1], nil
}
