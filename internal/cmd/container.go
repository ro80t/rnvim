package cmd

import (
	"github.com/spf13/cobra"

	"github.com/ro80t/rnvim/internal/nvimsetup"
	"github.com/ro80t/rnvim/internal/transport"
)

// provisionFlags are the --strategy/--config/--nvim-bin flags shared by
// docker/podman/ssh, which all provision nvim via internal/nvimsetup.
type provisionFlags struct {
	strategy string
	config   string
	nvimBin  string
}

func (f *provisionFlags) register(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.strategy, "strategy", "auto",
		"nvim provisioning strategy: auto (try network install, fall back to push) | install | push")
	cmd.Flags().StringVar(&f.config, "config", defaultLocalConfigDir(), "local nvim config directory to copy")
	cmd.Flags().StringVar(&f.nvimBin, "nvim-bin", "", "local nvim binary to push (default: found in PATH)")
}

func (f *provisionFlags) options() nvimsetup.Options {
	return nvimsetup.Options{Strategy: f.strategy, LocalConfig: f.config, LocalNvimBin: f.nvimBin}
}

func newContainerCmd(bin string) *cobra.Command {
	var pf provisionFlags
	cmd := &cobra.Command{
		Use:   bin + " <container>",
		Short: "Connect nvim to a " + bin + " container",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			t := transport.NewContainerTransport(bin, args[0])
			res, err := nvimsetup.Ensure(t, pf.options())
			if err != nil {
				return err
			}
			return t.RunInteractive(res.Env, res.Exec)
		},
	}
	pf.register(cmd)
	return cmd
}
