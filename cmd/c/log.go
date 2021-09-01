package c

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/shipengqi/container/internal/action"
)


func newLogsCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "logs [options] CONTAINER",
		Short:   "Fetch the logs of a container",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("missing container id")
			}
			a := action.NewLogAction(args[0])
			return action.Execute(a)
		},
	}
	c.DisableFlagsInUseLine = true
	return c
}
