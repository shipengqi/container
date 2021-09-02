package c

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/shipengqi/container/internal/action"
)


func newRunCmd() *cobra.Command {
	o := &action.RunActionOptions{}
	c := &cobra.Command{
		Use:   "run [options]",
		Short: "Create a container with namespace and cgroups limit.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return errors.New("missing image or container command")
			}
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
	f.BoolVarP(&o.Detach, "detach", "d",false, "Run container in background and print container ID")
	f.StringVar(&o.Name, "name", "", "Assign a name to the container")
	f.StringSliceVarP(&o.Envs, "env", "e", nil, "Set environment variables")
	f.StringVar(&o.Network, "network", "qcontainer0", "Connect a container to a network")
	f.StringSliceVarP(&o.Publish, "publish", "p", nil, "Publish a container's port(s) to the host")
	return c
}
