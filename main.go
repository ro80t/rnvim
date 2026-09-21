// Command rnvim connects a local nvim session to a devcontainer, a
// docker/podman container, or an ssh host with as little setup as possible.
package main

import (
	"os"

	"github.com/ro80t/rnvim/internal/cmd"
)

// version is overridable at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if err := cmd.NewRoot(version).Execute(); err != nil {
		os.Exit(1)
	}
}
