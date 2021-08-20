package subsystems

import (
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
)

type CpuSubSystem struct {}

func (s *CpuSubSystem) Name() string {
	return "cpu"
}

func (s *CpuSubSystem) Set(cgrouppath string, res *Resources) error {
	subsysCgroupPath, err := GetCgroupPath(s.Name(), cgrouppath, true)
	if err != nil {
		return err
	}
	if len(res.CpuShare) > 0 {
		if err := ioutil.WriteFile(
			filepath.Join(subsysCgroupPath, FileNameCpuShares),
			[]byte(res.CpuShare), 0644); err != nil {
			return errors.Errorf("set %s: %v", FileNameCpuShares, err)
		}
	}
	return nil
}

func (s *CpuSubSystem) Remove(cgrouppath string) error {
	return removeCgroup(s.Name(), cgrouppath)
}

func (s *CpuSubSystem)Apply(cgrouppath string, pid int) error {
	return applyCgroup(s.Name(), cgrouppath, pid)
}

