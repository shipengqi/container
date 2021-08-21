package c

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"text/tabwriter"

	"github.com/shipengqi/container/pkg/log"
	"github.com/spf13/cobra"

	"github.com/shipengqi/container/internal/container"
)


func newListCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "ps [options]",
		Short:   "List containers.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ListContainers()
			return nil
		},
	}
	c.DisableFlagsInUseLine = true
	return c
}

func ListContainers() {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, "")
	dirURL = dirURL[:len(dirURL)-1]
	files, err := ioutil.ReadDir(dirURL)
	if err != nil {
		log.Errorf("read dir %s error %v", dirURL, err)
		return
	}

	var containers []*container.Information
	for _, file := range files {
		tmpContainer, err := getContainerInfo(file)
		if err != nil {
			log.Errorf("Get container info error %v", err)
			continue
		}
		containers = append(containers, tmpContainer)
	}

	// 使用 tabwriter.NewWriter 在控制台打印出容器信息
	// tabwriter 是引用 text/tabwriter 类库，用于在控制台打印对齐的表格
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, item := range containers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreatedTime)
	}
	// 刷新标准输出流，打印信息
	if err := w.Flush(); err != nil {
		log.Errorf("Flush error %v", err)
		return
	}
}

func getContainerInfo(file os.FileInfo) (*container.Information, error) {
	containerId := file.Name()
	configFileDir := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	configFileDir = configFileDir + container.ConfigName
	content, err := ioutil.ReadFile(configFileDir)
	if err != nil {
		log.Errorf("read file %s error %v", configFileDir, err)
		return nil, err
	}
	var containerInfo container.Information
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		log.Errorf("Json unmarshal error %v", err)
		return nil, err
	}

	return &containerInfo, nil
}