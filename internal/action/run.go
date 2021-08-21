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
	p, wp, err := container.NewParentProcess(r.options.TTY, r.options.Volume)
	if err := p.Start(); err != nil {
		log.Errort("parent run", zap.Error(err))
		return err
	}
	mntUrl := "/root/mnt/"
	rootUrl := "/root/"
	defer container.DeleteWorkSpace(rootUrl, mntUrl, r.options.Volume)
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

// [root@shcCDFrh75vm7 container]# ./container run -it -m 100m --cpus 1 -v /root/q.container.volume:containervolume /bin/sh
// 2021-08-21T15:55:29.303+0800	INFO	running: /bin/sh
// 2021-08-21T15:55:29.303+0800	INFO	running: [/bin/sh]
// 2021-08-21T15:55:29.303+0800	DEBUG	***** [RUN] PreRun *****
// 2021-08-21T15:55:29.304+0800	DEBUG	***** RUN Run *****
// 2021-08-21T15:55:29.351+0800	DEBUG	volume container url: /root/mnt/containervolume
// 2021-08-21T15:55:29.351+0800	DEBUG	volume dirs: lowerdir=/root/q.container.volume.ro,upperdir=/root/q.container.volume,workdir=/root/q.container.work
// 2021-08-21T15:55:29.403+0800	INFO	send cmd: /bin/sh
// 2021-08-21T15:55:29.403+0800	INFO	send cmd: /bin/sh success
// 2021-08-21T15:55:29.406+0800	INFO	initializing
// 2021-08-21T15:55:29.406+0800	DEBUG	setting mount
// 2021-08-21T15:55:29.406+0800	DEBUG	pwd: /root/mnt
// 2021-08-21T15:55:29.453+0800	DEBUG	find cmd path: /bin/sh
// 2021-08-21T15:55:29.454+0800	DEBUG	syscall.Exec cmd path: /bin/sh
// / # /bin/ls
// bin              dev              home             lib64            root             tmp              var
// containervolume  etc              lib              proc             sys              usr              version.txt
// / # cd containervolume/
// /containervolume # /bin/ls
// ro
// /containervolume # cp /version.txt ./
// /bin/sh: cp: not found
// /containervolume # /bin/cp /version.txt ./
// /containervolume # /bin/ls
// ro           version.txt
// /containervolume # /bin/ls
// version.txt
// /containervolume # /bin/rm -rf version.txt