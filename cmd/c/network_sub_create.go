package c

import (
	"github.com/spf13/cobra"

	"github.com/shipengqi/container/internal/action"
)

func newNetworkSubCreateCmd() *cobra.Command {
	o := action.NetWorkCreateActionOptions{}
	c := &cobra.Command{
		Use:   "create [options]",
		Short: "Create a network",
		RunE: func(cmd *cobra.Command, args []string) error {
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
