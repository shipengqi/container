package action

import (
	"fmt"
	"github.com/pkg/errors"
	"os/exec"

	"github.com/shipengqi/container/internal/container"
	"github.com/shipengqi/container/pkg/log"
)

type commitA struct {
	*action

	imageName   string
	containerId string
}

func NewCommitAction(containerId, imageName string) Interface {
	return &commitA{
		action: &action{
			name: "commit",
		},
		imageName:   imageName,
		containerId: containerId,
	}
}

func (a *commitA) Run() error {
	rootfs := fmt.Sprintf(container.MntUrl, a.containerId)
	imageTar := "/root/" + a.imageName + ".tar"
	log.Debugf("commit: %s", imageTar)
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", rootfs, ".").CombinedOutput(); err != nil {
		return errors.Errorf("tar: %s: %v", rootfs, err)
	}
	return nil
}
