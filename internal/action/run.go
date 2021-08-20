package action

import (
	"os"
	"strings"

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
	p, wp, err := container.NewParentProcess(r.options.TTY)
	if err := p.Start(); err != nil {
		log.Errort("parent run", zap.Error(err))
		return err
	}
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
		log.Errort("parent wait", zap.Error(err))
		return err
	}
	err = p.Wait()
	if err != nil {
		log.Errort("parent wait", zap.Error(err))
		return err
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
	return nil
}

// [root@shcCDFrh75vm7 container]# ./container run -it -m 100m --cpus 1 sh
// 2021-08-20T14:46:26.729+0800    INFO    running: sh
// 2021-08-20T14:46:26.732+0800    INFO    running: [sh]
// 2021-08-20T14:46:26.732+0800    DEBUG   ***** [RUN] PreRun *****
// 2021-08-20T14:46:26.732+0800    DEBUG   ***** RUN Run *****
// 2021-08-20T14:46:26.740+0800    INFO    send cmd: sh
// 2021-08-20T14:46:26.742+0800    INFO    initializing
// 2021-08-20T14:46:26.743+0800    DEBUG   find cmd path: sh
// 2021-08-20T14:46:26.743+0800    DEBUG   syscall.Exec cmd path: /usr/bin/sh
// sh-4.2# stress --vm-bytes 200m --vm-keep -m 1
// stress: info: [7] dispatching hogs: 0 cpu, 0 io, 1 vm, 0 hdd
// stress: FAIL: [7] (415) <-- worker 8 got signal 9
// stress: WARN: [7] (417) now reaping child worker processes
// stress: FAIL: [7] (451) failed run completed in 0s
// cobra 不支持
// [root@shcCDFrh75vm7 container]# ./container run -it -m 100m --cpus 1 - stress --vm-bytes 200m --vm-keep -m 1
// unknown flag: --vm-bytes
// 需要禁用 flags 解析
