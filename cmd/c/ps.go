package c

import (
	"github.com/spf13/cobra"

	"github.com/shipengqi/container/internal/action"
)


func newListCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "ps [options]",
		Short:   "List containers",
		RunE: func(cmd *cobra.Command, args []string) error {
			a := action.NewPSAction()
			return action.Execute(a)
		},
	}
	c.DisableFlagsInUseLine = true
	return c
}
