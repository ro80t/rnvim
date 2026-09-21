package cmd

import "github.com/spf13/cobra"

func newConnectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Provision nvim on a target (if needed) and attach to it",
	}
	cmd.AddCommand(
		newDevcontainerCmd(),
		newContainerCmd("docker"),
		newContainerCmd("podman"),
		newSSHCmd(),
	)
	return cmd
}
