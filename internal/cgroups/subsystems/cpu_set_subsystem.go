package subsystems

import (
	"github.com/pkg/errors"
	"io/ioutil"
	"path/filepath"
)

type CpusetSubSystem struct {}

func (s *CpusetSubSystem) Name() string {
	return "cpuset"
}

// Set https://github.com/xianlubird/mydocker/issues/66
// https://blog.csdn.net/xftony/article/details/80536562
// When using cpuset in NUMA architecture, cpuset.cpus and cpuset.mems need to
// be written at the same time or both are empty to write PID to tasks
func (s *CpusetSubSystem) Set(cgrouppath string, res *Resources) error {
	subsysCgroupPath, err := GetCgroupPath(s.Name(), cgrouppath, true)
	if err != nil {
		return err
	}
	// 这个版本下如果 --cpu 没有设置会报下面的错
	// set tasks: write /sys/fs/cgroup/cpuset/q.container.cgroup/tasks: no space left on device
	// 因为写入 pid 到 tasks 之前，cpuset.cpus 和 cpuset.mems 需要设置
	if len(res.CpuSet) > 0 {
		if err := ioutil.WriteFile(
			filepath.Join(subsysCgroupPath, FileNameCpuSetMems),
			[]byte("0"), 0644); err != nil {
			return errors.Errorf("set %s: %v", FileNameCpuSetMems, err)
		}
		if err := ioutil.WriteFile(
			filepath.Join(subsysCgroupPath, FileNameCpuSetCpus),
			[]byte(res.CpuSet), 0644); err != nil {
			return errors.Errorf("set %s: %v", FileNameCpuSetCpus, err)
		}
	}
	return nil
}

func (s *CpusetSubSystem) Remove(cgrouppath string) error {
	return removeCgroup(s.Name(), cgrouppath)
}


func (s *CpusetSubSystem)Apply(cgrouppath string, pid int) error {
	return applyCgroup(s.Name(), cgrouppath, pid)
}
