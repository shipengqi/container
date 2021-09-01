package action

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"syscall"

	"github.com/pkg/errors"

	"github.com/shipengqi/container/internal/container"
	"github.com/shipengqi/container/pkg/log"
)

type stopA struct {
	*action

	containerId string
}

func NewStopAction(containerId string) Interface {
	return &stopA{
		action: &action{
			name: "stop",
		},
		containerId: containerId,
	}
}

func (a *stopA) Run() error {
	info, err := getContainerInfoById(a.containerId)
	if err != nil {
		return errors.Errorf("get container: %s, %v", a.containerId, err)
	}
	pidInt, err := strconv.Atoi(info.Pid)
	if err != nil {
		return errors.Wrap(err, "strconv.Atoi")
	}

	if err := syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		return errors.Errorf("stop container: %s, %v", a.containerId, err)
	}

	info.Status = container.StatusStop
	info.Pid = " "
	newContentBytes, err := json.Marshal(info)
	if err != nil {
		return errors.Errorf("marshal container: %s, %v", a.containerId, err)
	}
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, a.containerId)
	configFilePath := dirURL + container.ConfigName
	if err := ioutil.WriteFile(configFilePath, newContentBytes, 0622); err != nil {
		log.Warnf("write: %s, %v", configFilePath, err)
	}
	log.Info(a.containerId)
	return nil
}
