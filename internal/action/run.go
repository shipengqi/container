package action

import (
	"os"
	"strings"

	"github.com/pkg/errors"
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
	// 由于要实现后台运行，所以这里暂时去掉 delete workspace
	// mntUrl := "/root/mnt/"
	// rootUrl := "/root/"
	// defer container.DeleteWorkSpace(rootUrl, mntUrl, r.options.Volume)

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

// 如果 -d 运行后子进程没有被 init 进程托管，可能是 top 命令出错了,可以先以 -ti 的形式进入程序，运行 top 命令，如果 top 运行失败，ps -ef 时
// 自然是看不到子进程的