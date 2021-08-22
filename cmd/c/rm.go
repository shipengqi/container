package c

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/shipengqi/container/pkg/log"
	"github.com/spf13/cobra"

	"github.com/shipengqi/container/internal/container"
)


func newRemoveContainerCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "rm [options]",
		Short:   "Remove one container.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("missing container id")
			}
			removeContainer(args[0])
			return nil
		},
	}
	c.DisableFlagsInUseLine = true
	return c
}

func removeContainer(containerId string) {
	containerInfo, err := getContainerInfoById(containerId)
	if err != nil {
		log.Errorf("Get container %s info error %v", containerId, err)
		return
	}
	if containerInfo.Status != container.StatusStop {
		log.Errorf("Couldn't remove running container")
		return
	}
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	if err := os.RemoveAll(dirURL); err != nil {
		log.Errorf("Remove file %s error %v", dirURL, err)
		return
	}
}