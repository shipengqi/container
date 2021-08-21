package c

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/shipengqi/container/internal/container"
	"github.com/shipengqi/container/pkg/log"
)


func newLogsCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "logs [options]",
		Short:   "Fetch the logs of a container.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("Please input your container id")
			}
			logContainer(args[0])
			return container.InitProcess()
		},
	}
	c.DisableFlagsInUseLine = true
	return c
}

func logContainer(containerName string) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	logFileLocation := dirURL + container.LogFileName
	file, err := os.Open(logFileLocation)
	defer file.Close()
	if err != nil {
		log.Errorf("Log container open file %s error %v", logFileLocation, err)
		return
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Errorf("Log container read file %s error %v", logFileLocation, err)
		return
	}
	_, _ = fmt.Fprint(os.Stdout, string(content))
}