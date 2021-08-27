package c

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/shipengqi/container/internal/action"
	"github.com/shipengqi/container/internal/network"
)

func newNetworkSubCreateCmd() *cobra.Command {
	o := action.NetWorkCreateActionOptions{}
	c := &cobra.Command{
		Use:   "create [options]",
		Short: "Create a network",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("missing network name")
			}
			err := network.Init()
			if err != nil {
				return err
			}
			err = network.CreateNetwork(o.Driver, args[0], o.Subnet)
			if err != nil {
				return errors.Wrap(err, "create network")
			}
			return nil
		},
	}
	c.Flags().SortFlags = false
	c.DisableFlagsInUseLine = true
	f := c.Flags()
	f.StringVar(&o.Subnet, "subnet", "", "Subnet in CIDR format that represents a network segment")
	f.StringVar(&o.Driver, "driver", "bridge", "Driver to manage the Network")

	return c
}
