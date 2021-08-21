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
	mntUrl := "/root/mnt/"
	rootUrl := "/root/"
	defer container.DeleteWorkSpace(rootUrl, mntUrl)
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
	log.Infof("send cmd: %s success", command)
	return nil
}

// https://github.com/xianlubird/mydocker/issues/14
// exec.LookPath	{"error": "exec: \"sh\": executable file not found in $PATH"}
// 要用 /bin/sh
// [root@shcCDFrh75vm7 container]# ./container run -m 100m --cpus 1 -it /bin/sh
// 2021-08-21T11:34:30.123+0800	INFO	running: /bin/sh
// 2021-08-21T11:34:30.123+0800	INFO	running: [/bin/sh]
// 2021-08-21T11:34:30.123+0800	DEBUG	***** [RUN] PreRun *****
// 2021-08-21T11:34:30.123+0800	DEBUG	***** RUN Run *****
// 2021-08-21T11:34:30.208+0800	INFO	send cmd: /bin/sh
// 2021-08-21T11:34:30.208+0800	INFO	send cmd: /bin/sh success
// 2021-08-21T11:34:30.210+0800	INFO	initializing
// 2021-08-21T11:34:30.211+0800	DEBUG	setting mount
// 2021-08-21T11:34:30.211+0800	DEBUG	pwd: /root/mnt
// 2021-08-21T11:34:30.258+0800	DEBUG	find cmd path: /bin/sh
// 2021-08-21T11:34:30.258+0800	DEBUG	syscall.Exec cmd path: /bin/sh
// / # ls
// /bin/sh: ls: not found
// / # /bin/ls
// bin          dev          etc          home         lib          lib64        proc         root         sys          tmp          usr          var          version.txt
// / # /bin/ls
// bin          dev          etc          home         lib          lib64        proc         root         sys          tmp          usr          var          version.txt
// / # /bin/cp version.txt /tmp/
// / # /bin/ls /tmp
// version.txt
// 宿主机下的 /root/busybox/tmp 没有 version.txt，并没有受到影响