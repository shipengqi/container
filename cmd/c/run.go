package c

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/shipengqi/container/internal/action"
	"github.com/shipengqi/container/pkg/log"
)


func newRunCmd() *cobra.Command {
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
			a := action.NewRunAction(args, o)
			return action.Execute(a)
		},
	}

	c.Flags().SortFlags = false
	c.DisableFlagsInUseLine = true
	f := c.Flags()
	f.BoolVarP(&o.Interactive, "interactive", "i", false, "Keep STDIN open even if not attached")
	f.BoolVarP(&o.TTY, "tty", "t",false, "Allocate a pseudo-TTY")
	f.StringVarP(&o.MemoryLimit, "memory", "m","", "Memory limit")
	f.StringVar(&o.CpuSet, "cpus","", "Number of CPUs")
	f.StringVarP(&o.CpuShare, "cpu-shares", "c","", "CPU shares (relative weight)")
	f.StringVarP(&o.Volume, "volume", "v","", "Bind mount a volume")
	return c
}
