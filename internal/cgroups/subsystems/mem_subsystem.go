package subsystems

import (
	"github.com/pkg/errors"
	"io/ioutil"
	"path/filepath"
)

type MemorySubSystem struct {
}

func (s *MemorySubSystem) Name() string {
	return "memory"
}

func (s *MemorySubSystem) Set(cgrouppath string, res *Resources) error {
	subsysCgroupPath, err := GetCgroupPath(s.Name(), cgrouppath, true)
	if err != nil {
		return err
	}
	if len(res.MemoryLimit) > 0 {
		if err := ioutil.WriteFile(
			filepath.Join(subsysCgroupPath, FileNameMemoryLimitInBytes),
			[]byte(res.MemoryLimit), 0644); err != nil {
			return errors.Errorf("set %s: %v", FileNameMemoryLimitInBytes, err)
		}
	}
	return nil
}

func (s *MemorySubSystem) Remove(cgrouppath string) error {
	return removeCgroup(s.Name(), cgrouppath)
}

func (s *MemorySubSystem) Apply(cgrouppath string, pid int) error {
	return applyCgroup(s.Name(), cgrouppath, pid)
}
