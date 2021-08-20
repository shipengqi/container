package c

import (
	"github.com/shipengqi/container/internal/container"
	"github.com/shipengqi/container/pkg/log"
	"github.com/spf13/cobra"
)


func newInit() *cobra.Command {
	c := &cobra.Command{
		Use:     "init [options]",
		Short:   "Init container process run user's process in container.",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info("initializing")
			return container.InitProcess()
		},
	}
	c.DisableFlagsInUseLine = true
	c.Hidden = true
	return c
}
