package c

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/shipengqi/container/internal/action"
)


func newRemoveContainerCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "rm [options] CONTAINER",
		Short:   "Remove one container",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("missing container id")
			}
			a := action.NewRMAction(args[0])
			return action.Execute(a)
		},
	}
	c.DisableFlagsInUseLine = true
	return c
}
