package transport

import (
	"testing"

	"rnvim/internal/dockertest"
)

// TestContainerTransport runs the shared transport suite against real
// docker and podman containers (each skipped individually if its daemon
// isn't available).
func TestContainerTransport(t *testing.T) {
	for _, bin := range []string{"docker", "podman"} {
		t.Run(bin, func(t *testing.T) {
			dockertest.RequireDaemon(t, bin)
			id := dockertest.RunContainer(t, bin, nil, "alpine:3.20", "sleep", "3600")
			testTransportSuite(t, NewContainerTransport(bin, id))
		})
	}
}
