package transport

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// SSHTransport shells out to the system ssh/scp binaries, so it inherits
// the user's ~/.ssh/config, agent, and known_hosts for free.
type SSHTransport struct {
	Host      string   // [user@]host or an ssh config alias
	Port      string   // "" = default
	Identity  string   // "" = default
	ExtraArgs []string // extra flags forwarded to both ssh and scp, e.g. []string{"-o", "StrictHostKeyChecking=no"}
}

func (s *SSHTransport) baseArgs() []string {
	var args []string
	if s.Port != "" {
		args = append(args, "-p", s.Port)
	}
	if s.Identity != "" {
		args = append(args, "-i", s.Identity)
	}
	return append(args, s.ExtraArgs...)
}

func (s *SSHTransport) RunInteractive(env map[string]string, args ...string) error {
	remoteCmd := buildRemoteCommand(env, args)
	cmdArgs := append(append([]string{}, s.baseArgs()...), "-t", s.Host, remoteCmd)
	cmd := exec.Command("ssh", cmdArgs...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

func (s *SSHTransport) Run(script string) (string, error) {
	cmdArgs := append(append([]string{}, s.baseArgs()...), s.Host, script)
	return runCapture(exec.Command("ssh", cmdArgs...))
}

func (s *SSHTransport) CopyTo(localPath, remotePath string) error {
	if err := s.MkdirAll(parentDir(remotePath)); err != nil {
		return fmt.Errorf("mkdir -p %s: %w", parentDir(remotePath), err)
	}
	var scpArgs []string
	if s.Port != "" {
		scpArgs = append(scpArgs, "-P", s.Port)
	}
	if s.Identity != "" {
		scpArgs = append(scpArgs, "-i", s.Identity)
	}
	scpArgs = append(scpArgs, s.ExtraArgs...)
	scpArgs = append(scpArgs, "-r", localPath, s.Host+":"+remotePath)
	cmd := exec.Command("scp", scpArgs...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func (s *SSHTransport) MkdirAll(remotePath string) error {
	_, err := s.Run("mkdir -p " + shQuote(remotePath))
	return err
}

func (s *SSHTransport) Home() (string, error) {
	out, err := s.Run("echo $HOME")
	return strings.TrimSpace(out), err
}

func (s *SSHTransport) Uname() (string, string, error) {
	out, err := s.Run("uname -s && uname -m")
	if err != nil {
		return "", "", err
	}
	f := strings.Fields(out)
	if len(f) < 2 {
		return "", "", fmt.Errorf("unexpected uname output: %q", out)
	}
	return f[0], f[1], nil
}
