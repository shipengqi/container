package c

import (
	"github.com/pkg/errors"
	"github.com/shipengqi/container/internal/action"
	"github.com/spf13/cobra"
)


func newCommitCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "commit [options] CONTAINER IMAGE",
		Short:   "Create a new image from a container's changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return errors.New("missing container id or image name")
			}
			a := action.NewCommitAction(args[0], args[1])
			return action.Execute(a)
		},
	}
	c.DisableFlagsInUseLine = true
	return c
}
