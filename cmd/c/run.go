package c

import (
	"github.com/pkg/errors"
	"github.com/shipengqi/container/pkg/log"
	"github.com/spf13/cobra"

	"github.com/shipengqi/container/internal/action"
)


func newRun() *cobra.Command {
	o := &action.RunActionOptions{}
	c := &cobra.Command{
		Use:   "run [options]",
		Short: "Create a container with namespace and cgroups limit.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("missing container command")
			}
			log.Infof("running: %s", args[0])
			log.Infof("running: %v", args)
			a := action.NewRunAction(args[0])
			return action.Execute(a)
		},
	}

	c.Flags().SortFlags = false
	c.DisableFlagsInUseLine = true
	f := c.Flags()
	f.BoolVarP(&o.Interactive, "interactive", "i", false, "Keep STDIN open even if not attached")
	f.BoolVarP(&o.TTY, "tty", "t",false, "Allocate a pseudo-TTY")
	return c
}
