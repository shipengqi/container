package c

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/shipengqi/container/internal/action"
)


func newNetworkSubRemoveCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "rm [options]",
		Short:   "Remove one networks",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("missing network name")
			}
			a := action.NewNetworkRemoveAction(args[0])
			return action.Execute(a)
		},
	}
	c.DisableFlagsInUseLine = true
	return c
}
