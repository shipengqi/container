package action

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"

	"github.com/shipengqi/container/internal/container"
	"github.com/shipengqi/container/pkg/log"
)

type rmA struct {
	*action

	containerId string
}

func NewRMAction(containerId string) Interface {
	return &rmA{
		action: &action{
			name: "rm",
		},
		containerId: containerId,
	}
}

func (a *rmA) Run() error {
	containerInfo, err := getContainerInfoById(a.containerId)
	if err != nil {
		return errors.Errorf("get container: %s, %v", a.containerId, err)
	}
	if containerInfo.Status != container.StatusStop {
		return errors.New("cannot remove running container")
	}
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, a.containerId)
	if err := os.RemoveAll(dirURL); err != nil {
		return errors.Errorf("remove %s, %v", dirURL, err)
	}

	return nil
}

func getContainerInfoById(containerId string) (*container.Information, error) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	configFilePath := dirURL + container.ConfigName
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Errorf("Read file %s error %v", configFilePath, err)
		return nil, err
	}
	var containerInfo container.Information
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return nil, errors.Wrap(err, "container info unmarshal")
	}
	return &containerInfo, nil
}
