package c

import (
	"github.com/spf13/cobra"
)

func newNetworkCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "network [options]",
		Short: "Manage networks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	c.DisableFlagsInUseLine = true

	c.AddCommand(
		newNetworkSubCreateCmd(),
		newNetworkSubRemoveCmd(),
		newNetworkSubListCmd(),
	)
	return c
}
