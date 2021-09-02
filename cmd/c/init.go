package c

import (
	"github.com/spf13/cobra"

	"github.com/shipengqi/container/internal/container"
)


func newInitCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "init [options]",
		Short:   "Init container process run user's process in container",
		RunE: func(cmd *cobra.Command, args []string) error {
			return container.InitProcess()
		},
	}
	c.DisableFlagsInUseLine = true
	c.Hidden = true
	return c
}
