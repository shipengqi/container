package subsystems

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/shipengqi/container/pkg/log"
	"github.com/shipengqi/container/pkg/utils"
)

const (
	FileNameTasks              = "tasks"
	FileNameMemoryLimitInBytes = "memory.limit_in_bytes"
	FileNameCpuSetCpus         = "cpuset.cpus"
	FileNameCpuSetMems         = "cpuset.mems"
	FileNameCpuShares          = "cpu.shares"
)

var (
	SubsystemsIns = []Interface{
		&CpusetSubSystem{},
		&MemorySubSystem{},
		&CpuSubSystem{},
	}
)

type Resources struct {
	MemoryLimit string
	CpuShare    string
	CpuSet      string
}

type Interface interface {
	// Name return name of subsystem, etc. cpu memory
	Name() string
	// Set setting a cgroup to subsystem
	Set(cgrouppath string, res *Resources) error
	// Apply applying a process to a cgroup
	Apply(cgrouppath string, pid int) error
	// Remove a cgroup
	Remove(cgrouppath string) error
}

func FindCgroupMountpoint(subsystem string) string {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return ""
	}
	defer f.Close()

	// 30 24 0:26 / /sys/fs/cgroup/memory rw,nosuid,nodev,noexec,relatime shared:13 - cgroup cgroup rw,seclabel,memory
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		txt := scanner.Text()
		fields := strings.Split(txt, " ")
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			if opt == subsystem {
				return fields[4] // /sys/fs/cgroup/memory
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return ""
	}

	return ""
}

// GetCgroupPath 获取当前 subsystem 在虚拟文件系统中的路径
func GetCgroupPath(subsystem string, cgrouppath string, create bool) (string, error) {
	cgroupRoot := FindCgroupMountpoint(subsystem)
	abPath := filepath.Join(cgroupRoot, cgrouppath)
	if !utils.IsDir(abPath) {
		if !create {
			return "", errors.New("cgroup path not found")
		}
		if err := os.Mkdir(abPath, 0755); err != nil {
			log.Errort("os.Mkdir", zap.Error(err))
			return "", err
		}
	}
	return abPath, nil
}


func applyCgroup(subsystem, cgrouppath string, pid int) error {
	subsysCgroupPath, err := GetCgroupPath(subsystem, cgrouppath, true)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(
		filepath.Join(subsysCgroupPath, FileNameTasks),
		[]byte(strconv.Itoa(pid)), 0644); err != nil {
		return errors.Errorf("set %s: %v", FileNameTasks, err)
	}

	return nil
}

func removeCgroup(subsystem, cgrouppath string) error {
	subsysCgroupPath, err := GetCgroupPath(subsystem, cgrouppath, true)
	if err != nil {
		return err
	}
	return os.RemoveAll(subsysCgroupPath)
}
