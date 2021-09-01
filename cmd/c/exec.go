package c

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/shipengqi/container/internal/action"
	_ "github.com/shipengqi/container/internal/nsenter"
	"github.com/shipengqi/container/pkg/log"
)

func newExecCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "exec [options]",
		Short: "Run a command in a running container",
		RunE: func(cmd *cobra.Command, args []string) error {
			// This is for callback
			if os.Getenv(action.EnvExecPid) != "" {
				log.Infof("pid callback pid %s", os.Getgid())
				return nil
			}

			if len(args) < 2 {
				return errors.New("missing container id or command")
			}
			containerId := args[0]

			a := action.NewExecAction(containerId, args[1:])
			return action.Execute(a)
		},
	}
	c.DisableFlagsInUseLine = true
	return c
}
