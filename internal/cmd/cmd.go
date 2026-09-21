// Package cmd builds the rnvim cobra command tree.
package cmd

import "github.com/spf13/cobra"

// NewRoot builds the root "rnvim" command with all subcommands attached.
func NewRoot(version string) *cobra.Command {
	root := &cobra.Command{
		Use:          "rnvim",
		Short:        "Connect nvim to a devcontainer/docker/podman/ssh target",
		Version:      version,
		SilenceUsage: true,
	}
	root.AddCommand(newConnectCmd())
	return root
}
