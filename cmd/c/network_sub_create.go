package c

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/shipengqi/container/internal/action"
)

func newNetworkSubCreateCmd() *cobra.Command {
	o := action.NetworkCreateActionOptions{}
	c := &cobra.Command{
		Use:   "create [options]",
		Short: "Create a network",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("missing network name")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			a := action.NewNetworkCreateAction(args[0], &o)
			return action.Execute(a)
		},
	}
	c.Flags().SortFlags = false
	c.DisableFlagsInUseLine = true
	f := c.Flags()
	f.StringVar(&o.Subnet, "subnet", "", "Subnet in CIDR format that represents a network segment")
	f.StringVar(&o.Driver, "driver", "bridge", "Driver to manage the Network")

	return c
}
