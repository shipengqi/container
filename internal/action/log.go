package action

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"

	"github.com/shipengqi/container/internal/container"
)

type logA struct {
	*action

	containerId string
}

func NewLogAction(containerId string) Interface {
	return &logA{
		action: &action{
			name: "logs",
		},
		containerId: containerId,
	}
}

func (a *logA) Run() error {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, a.containerId)
	logFileLocation := dirURL + container.LogFileName
	file, err := os.Open(logFileLocation)
	defer file.Close()
	if err != nil {
		return errors.Errorf("open: %s: %v", logFileLocation, err)
	}

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return errors.Errorf("read: %s: %v", logFileLocation, err)
	}
	_, _ = fmt.Fprint(os.Stdout, string(content))
	return nil
}
