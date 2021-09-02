package action

import (
	"fmt"
	"io/ioutil"
	"os"
	"text/tabwriter"

	"github.com/pkg/errors"

	"github.com/shipengqi/container/internal/container"
	"github.com/shipengqi/container/pkg/log"
)

type psA struct {
	*action
}

func NewPSAction() Interface {
	return &psA{
		action: &action{
			name: "ps",
		},
	}
}

func (a *psA) Run() error {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, "")
	dirURL = dirURL[:len(dirURL)-1]
	files, err := ioutil.ReadDir(dirURL)
	if err != nil {
		return errors.Errorf("read dir: %s: %v", dirURL, err)
	}
	var containers []*container.Information
	for _, file := range files {
		tmpContainer, err := getContainerInfo(file)
		if err != nil {
			log.Warnf("get container: %v", err)
			continue
		}
		containers = append(containers, tmpContainer)
	}

	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, _ = fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, item := range containers {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreatedTime)
	}
	if err := w.Flush(); err != nil {
		return errors.Wrap(err, "Flush")
	}
	return nil
}

func getContainerInfo(file os.FileInfo) (*container.Information, error) {
	containerId := file.Name()
	configFileDir := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	configFileDir = configFileDir + container.ConfigName
	content, err := ioutil.ReadFile(configFileDir)
	if err != nil {
		return nil, errors.Errorf("read: %s: %v", configFileDir, err)
	}
	var containerInfo container.Information
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		return nil, errors.Wrap(err, "container info unmarshal")
	}

	return &containerInfo, nil
}
