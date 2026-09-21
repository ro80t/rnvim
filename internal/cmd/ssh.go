package cmd

import (
	"github.com/spf13/cobra"

	"rnvim/internal/nvimsetup"
	"rnvim/internal/transport"
)

func newSSHCmd() *cobra.Command {
	var pf provisionFlags
	var port, identity string
	cmd := &cobra.Command{
		Use:   "ssh <[user@]host>",
		Short: "Connect nvim over ssh",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			t := &transport.SSHTransport{Host: args[0], Port: port, Identity: identity}
			res, err := nvimsetup.Ensure(t, pf.options())
			if err != nil {
				return err
			}
			return t.RunInteractive(res.Env, res.Exec)
		},
	}
	pf.register(cmd)
	cmd.Flags().StringVarP(&port, "port", "p", "", "ssh port")
	cmd.Flags().StringVarP(&identity, "identity", "i", "", "ssh identity file")
	return cmd
}
