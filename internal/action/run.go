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
	"github.com/shipengqi/container/pkg/log"
)

type run struct {
	*action

	options *RunActionOptions
}

func NewRunAction(cmdArgs []string, options *RunActionOptions) Interface {
	return &run{
		action: &action{
			name:    "run",
			cmdArgs: cmdArgs,
		},
		options: options,
	}
}

func (r *run) Name() string {
	return r.name
}

func (r *run) Run() error {
	log.Debugf("***** %s Run *****", strings.ToUpper(r.name))
	p, wp, err := container.NewParentProcess(r.options.TTY, r.options.Volume)
	if err := p.Start(); err != nil {
		log.Errort("parent run", zap.Error(err))
		return err
	}

	// save container info
	containerId, err := saveContainerInfo(p.Process.Pid, r.cmdArgs, r.options.Name)
	if err != nil {
		log.Errorf("save container info error %v", err)
		return err
	}

	// 由于要实现后台运行，所以这里暂时去掉 delete workspace
	mntUrl := "/root/mnt/"
	rootUrl := "/root/"

	// use q.container.cgroup as group name
	// create cgroup manager
	cgroupManager := manager.New("q.container.cgroup")
	defer cgroupManager.Destroy()
	// set resource limitations
	res := &subsystems.Resources{
		MemoryLimit: r.options.MemoryLimit,
		CpuSet:      r.options.CpuSet,
		CpuShare:    r.options.CpuShare,
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
	err = notifyInitProcess(r.cmdArgs, wp)
	if err != nil {
		log.Errort("notify", zap.Error(err))
		return err
	}
	log.Infof("tty %v", r.options.TTY)
	if r.options.TTY {
		err = p.Wait()
		// tty 方式创建的容器，在容器退出后，需要删除容器的相关信息
		deleteContainerInfo(containerId)
		container.DeleteWorkSpace(rootUrl, mntUrl, r.options.Volume)
		if err != nil {
			log.Errort("parent wait", zap.Error(err))
			return err
		}
	}
	return nil
}

func (r *run) PreRun() error {
	if r.options.TTY && r.options.Detach {
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

func saveContainerInfo(containerPID int, commandArray []string, containerName string) (string, error) {
	id := containerId(10)
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArray, "")
	if containerName == "" {
		containerName = id
	}
	info := &container.Information{
		Id:          id,
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

	return id, nil
}

func deleteContainerInfo(containerId string) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	if err := os.RemoveAll(dirURL); err != nil {
		log.Errorf("remove dir %s error %v", dirURL, err)
	}
}

// [root@shcCDFrh75vm7 container]# ./container run -it -m 100m --cpus 1 /bin/sh
// 2021-08-21T20:34:47.917+0800	INFO	running: /bin/sh
// 2021-08-21T20:34:47.917+0800	INFO	running: [/bin/sh]
// 2021-08-21T20:34:47.917+0800	DEBUG	***** RUN Run *****
// 2021-08-21T20:34:47.976+0800	INFO	send cmd: /bin/sh
// 2021-08-21T20:34:47.976+0800	INFO	send cmd: /bin/sh success
// 2021-08-21T20:34:47.976+0800	INFO	tty true
// 2021-08-21T20:34:47.980+0800	INFO	initializing
// 2021-08-21T20:34:47.980+0800	DEBUG	setting mount
// 2021-08-21T20:34:47.980+0800	DEBUG	pwd: /root/mnt
// 2021-08-21T20:34:48.035+0800	DEBUG	find cmd path: /bin/sh
// 2021-08-21T20:34:48.035+0800	DEBUG	syscall.Exec cmd path: /bin/sh
// / #
// 打开一个新的 terminal
// [root@shcCDFrh75vm7 container]# ./container ps
// ID           NAME         PID         STATUS      COMMAND     CREATED
// 6573986770   6573986770   8781        running     /bin/sh     2021-08-21 20:34:47