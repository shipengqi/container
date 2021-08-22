package c

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"syscall"

	"github.com/pkg/errors"
	"github.com/shipengqi/container/pkg/log"
	"github.com/spf13/cobra"

	"github.com/shipengqi/container/internal/container"
)


func newStopCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "stop [options]",
		Short:   "Stop one running containers.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("missing container id")
			}
			stopContainer(args[0])
			return nil
		},
	}
	c.DisableFlagsInUseLine = true
	return c
}

func stopContainer(containerId string) {
	info, err := getContainerInfoById(containerId)
	if err != nil {
		log.Errorf("Get getContainerInfoById pid by name %s error %v", containerId, err)
		return
	}
	pidInt, err := strconv.Atoi(info.Pid)
	if err != nil {
		log.Errorf("strconv.Atoi pid: %v", err)
		return
	}

	if err := syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		log.Errorf("Stop container %s error %v", containerId, err)
		return
	}

	// 容器进程已经被 kill ，所以下面需要修改容器状态，PIO 可以直为空
	info.Status = container.StatusStop
	info.Pid = " "
	newContentBytes, err := json.Marshal(info)
	if err != nil {
		log.Errorf("Json marshal %s error %v", containerId, err)
		return
	}
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	configFilePath := dirURL + container.ConfigName
	if err := ioutil.WriteFile(configFilePath, newContentBytes, 0622); err != nil {
		log.Errorf("Write file %s error", configFilePath, err)
	}
	log.Info(containerId)
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
		log.Errorf("getContainerInfoById unmarshal error %v", err)
		return nil, err
	}
	return &containerInfo, nil
}

// 系统调用 kill 可以发送信号给迸程，通过传递 syscall.SIGTERM 信号，去杀掉容搭主进程
// SIGTERM 比较友好，进程能捕捉这个信号，根据需要来关闭程序。在关闭程序之前，可以结束打开的记录文件和完成正在做的任务。
// 在某些情况下，假如进程正在进行作业而且不能中断，那么进程可以忽略这个 SIGTERM 信号。
// SIGKILL 信号，进程是不能忽略的。发送 SIGKILL 信号给进程，Linux 就将进程终止。
// 所以下面的命令，在我运行 ./container stop 3890769609 之后，ps -ef 查看，进程还在
// 可以把代码里面的信号 SIGTERM 改成 SIGKILL，就可以看到进程会立刻被终止
// [root@shcCDFrh75vm7 container]# ./container run -d -m 100m --cpus 1 /bin/sleep 500s
// 2021-08-22T08:18:42.279+0800	INFO	running: /bin/sleep
// 2021-08-22T08:18:42.279+0800	INFO	running: [/bin/sleep 500s]
// 2021-08-22T08:18:42.279+0800	DEBUG	***** RUN Run *****
// 2021-08-22T08:18:42.397+0800	DEBUG	container id: 3890769609, name: 3890769609
// 2021-08-22T08:18:42.401+0800	INFO	send cmd: /bin/sleep 500s
// 2021-08-22T08:18:42.401+0800	INFO	send cmd: /bin/sleep 500s success
// 2021-08-22T08:18:42.401+0800	INFO	tty false
// 2021-08-22T08:18:42.402+0800	WARN	remove cgroup fail unlinkat /sys/fs/cgroup/cpuset/q.container.cgroup/cpuset.memory_spread_slab: operation not permitted
// 2021-08-22T08:18:42.402+0800	WARN	remove cgroup fail unlinkat /sys/fs/cgroup/memory/q.container.cgroup/memory.kmem.tcp.max_usage_in_bytes: operation not permitted
// 2021-08-22T08:18:42.403+0800	WARN	remove cgroup fail unlinkat /sys/fs/cgroup/cpu,cpuacct/q.container.cgroup/cpu.rt_period_us: operation not permitted
// 2021-08-22T08:18:42.403+0800	DEBUG	***** [RUN] PostRun *****
// [root@shcCDFrh75vm7 container]# ./container ps
// ID           NAME         PID         STATUS      COMMAND             CREATED
// 3431872348   3431872348               stopped     /bin/sleep500s      2021-08-22 08:14:37
// 3890769609   3890769609               stopped     /bin/sleep500s      2021-08-22 08:18:42
// 7618998790   7618998790               stopped     /bin/echotest log   2021-08-21 20:57:49