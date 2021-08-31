package action

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/shipengqi/container/pkg/utils"
	"go.uber.org/zap"

	"github.com/shipengqi/container/internal/cgroups/manager"
	"github.com/shipengqi/container/internal/cgroups/subsystems"
	"github.com/shipengqi/container/internal/container"
	"github.com/shipengqi/container/internal/network"
	"github.com/shipengqi/container/pkg/log"
)

type RunActionOptions struct {
	Interactive bool
	TTY         bool
	Detach      bool
	CpuSet      string
	MemoryLimit string
	CpuShare    string
	Volume      string
	Name        string
	Network     string
	Publish     []string
	Envs        []string
}

type runA struct {
	*action

	cmdArgs []string
	imgName string
	options *RunActionOptions
}

func NewRunAction(args []string, options *RunActionOptions) Interface {
	imageName := args[0]
	cmdArgs := args[1:]
	log.Infof("image name: %s", imageName)
	log.Infof("command: %v", cmdArgs)
	return &runA{
		action: &action{
			name: "run",
		},
		cmdArgs: cmdArgs,
		imgName: imageName,
		options: options,
	}
}

func (a *runA) Run() error {
	log.Debugf("***** %s Run *****", strings.ToUpper(a.name))
	containerId := containerId(10)
	p, wp, err := container.NewInitProcess(a.options.TTY, a.options.Volume, containerId, a.imgName, a.options.Envs)
	if err := p.Start(); err != nil {
		log.Errort("parent run", zap.Error(err))
		return err
	}

	// save container info
	containerName, err := saveContainerInfo(p.Process.Pid, a.cmdArgs, a.options.Name, containerId)
	if err != nil {
		log.Errorf("save container info error %v", err)
		return err
	}

	log.Debugf("container id: %s, name: %s", containerId, containerName)

	// use q.container.cgroup as group name
	// create cgroup manager
	cgroupManager := manager.New("q.container.cgroup")
	defer cgroupManager.Destroy()
	// set resource limitations
	res := &subsystems.Resources{
		MemoryLimit: a.options.MemoryLimit,
		CpuSet:      a.options.CpuSet,
		CpuShare:    a.options.CpuShare,
	}
	err = cgroupManager.Set(res)
	if err != nil {
		log.Errort("cgroup manager set", zap.Error(err))
		return err
	}
	// 将容器进程加入到各个 subsystem 挂载对应的 cgroup 中
	err = cgroupManager.Apply(p.Process.Pid)
	if err != nil {
		log.Errort("cgroup manager apply", zap.Error(err))
		return err
	}

	if len(a.options.Network) > 0 {
		// config container network
		err = network.Init()
		if err != nil {
			log.Errort("network init", zap.Error(err))
			return err
		}
		containerInfo := &container.Information{
			Id:          containerId,
			Pid:         strconv.Itoa(p.Process.Pid),
			Name:        containerName,
			PortMapping: a.options.Publish,
		}
		if err = network.Connect(a.options.Network, containerInfo); err != nil {
			log.Errort("connect network", zap.Error(err))
			return err
		}
	}
	err = notifyInitProcess(a.cmdArgs, wp)
	if err != nil {
		log.Errort("notify", zap.Error(err))
		return err
	}
	log.Infof("tty %v", a.options.TTY)
	if a.options.TTY {
		err = p.Wait()
		// tty 方式创建的容器，在容器退出后，需要删除容器的相关信息
		deleteContainerInfo(containerId)
		container.DeleteWorkSpace(a.options.Volume, containerId)
		if err != nil {
			log.Errort("parent wait", zap.Error(err))
			return err
		}
	}
	return nil
}

func (a *runA) PreRun() error {
	if a.options.TTY && a.options.Detach {
		return errors.New("--tty and --detach flags cannot both provided")
	}
	return nil
}

func notifyInitProcess(cmdArgs []string, wp *os.File) error {
	command := strings.Join(cmdArgs, " ")
	log.Infof("send cmd: %s", command)
	_, err := wp.WriteString(command)
	if err != nil {
		log.Errort("write pipe", zap.Error(err))
		return err
	}
	wp.Close()
	log.Infof("send cmd: %s success", command)
	return nil
}

// containerId 时间戳为种子，每次生成一个 10 以内的数字作为 letterBytes 数组的下标，最后拼
// 接生成整个容器的 ID
func containerId(n int) string {
	letterBytes := "1234567890"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func saveContainerInfo(containerPID int, commandArray []string, containerName, containerId string) (string, error) {
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArray, "")
	if containerName == "" {
		containerName = containerId
	}
	info := &container.Information{
		Id:          containerId,
		Pid:         strconv.Itoa(containerPID),
		Command:     command,
		CreatedTime: createTime,
		Status:      container.StatusRunning,
		Name:        containerName,
	}
	infoBytes, err := json.Marshal(info)
	if err != nil {
		log.Errorf("save container info error %v", err)
		return "", err
	}
	infoStr := string(infoBytes)
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if utils.IsNotExist(dirUrl) {
		if err := os.MkdirAll(dirUrl, 0622); err != nil {
			log.Errorf("Mkdir error %s error %v", dirUrl, err)
			return "", err
		}
	}
	fileName := dirUrl + "/" + container.ConfigName
	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		log.Errorf("create file %s error %v", fileName, err)
		return "", err
	}
	if _, err := file.WriteString(infoStr); err != nil {
		log.Errorf("write string error %v", err)
		return "", err
	}

	return containerName, nil
}

func deleteContainerInfo(containerId string) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	if err := os.RemoveAll(dirURL); err != nil {
		log.Errorf("remove dir %s error %v", dirURL, err)
	}
}
