package c

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/shipengqi/container/internal/network"
)


func newNetworkSubRemoveCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "rm [options]",
		Short:   "Remove one networks",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("missing network name")
			}
			err := network.Init()
			if err != nil {
				return err
			}
			err = network.DeleteNetwork(args[0])
			if err != nil {
				return errors.Wrap(err, "remove network")
			}
			return nil
		},
	}
	c.DisableFlagsInUseLine = true
	return c
}
