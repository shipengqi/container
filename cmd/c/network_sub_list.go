package c

import (
	"github.com/spf13/cobra"

	"github.com/shipengqi/container/internal/action"
)


func newNetworkSubListCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "ls [options]",
		Short:   "List networks",
		RunE: func(cmd *cobra.Command, args []string) error {
			a := action.NewNetworkListAction()
			return action.Execute(a)
		},
	}
	c.DisableFlagsInUseLine = true
	return c
}
