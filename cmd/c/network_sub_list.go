package c

import (
	"github.com/spf13/cobra"

	"github.com/shipengqi/container/internal/network"
)


func newNetworkSubListCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "ls [options]",
		Short:   "List networks",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := network.Init()
			if err != nil {
				return err
			}
			network.ListNetwork()
			return nil
		},
	}
	c.DisableFlagsInUseLine = true
	return c
}
