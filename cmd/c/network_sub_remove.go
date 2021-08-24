package c

import (
	"github.com/spf13/cobra"
)


func newNetworkSubRemoveCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "rm [options]",
		Short:   "Remove one networks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	c.DisableFlagsInUseLine = true
	return c
}
