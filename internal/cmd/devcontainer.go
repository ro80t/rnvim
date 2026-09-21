package cmd

import (
	"github.com/spf13/cobra"

	"github.com/ro80t/rnvim/internal/devcontainer"
)

func newDevcontainerCmd() *cobra.Command {
	var feature, config string
	cmd := &cobra.Command{
		Use:   "devcontainer [workspace]",
		Short: "Bring up a devcontainer (with a neovim feature) and connect nvim",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace := "."
			if len(args) > 0 {
				workspace = args[0]
			}
			return devcontainer.Connect(workspace, feature, config)
		},
	}
	cmd.Flags().StringVar(&feature, "feature", devcontainer.DefaultNeovimFeature, "devcontainer feature ref that installs neovim")
	cmd.Flags().StringVar(&config, "config", defaultLocalConfigDir(), "local nvim config directory to copy into the container")
	return cmd
}
