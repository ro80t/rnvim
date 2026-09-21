// Package transport abstracts "run a command / copy a file onto a remote
// target" over docker/podman exec and ssh, so nvimsetup and devcontainer
// don't need to know which one they're talking to.
package transport

// Transport is the minimal set of operations rnvim needs against a
// docker/podman container or an ssh host.
type Transport interface {
	// RunInteractive execs args with stdio attached to the current
	// terminal (a PTY on the remote side), with env exported first.
	RunInteractive(env map[string]string, args ...string) error
	// Run executes a posix shell script remotely and returns its stdout.
	// On failure, the returned error's message includes stderr.
	Run(script string) (string, error)
	// CopyTo copies a local file or directory to remotePath, creating
	// parent directories as needed.
	CopyTo(localPath, remotePath string) error
	// MkdirAll creates remotePath (and parents) on the target.
	MkdirAll(remotePath string) error
	// Home returns $HOME on the target.
	Home() (string, error)
	// Uname returns `uname -s` and `uname -m` from the target.
	Uname() (osName, arch string, err error)
}
