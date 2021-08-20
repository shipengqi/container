package manager

import (
	"github.com/shipengqi/container/internal/cgroups/subsystems"
	"github.com/shipengqi/container/pkg/log"
)

// Manager 把不同的 subsystem 中的 cgroup 管理起来，并与容器建立关系。
type Manager struct {
	// cgroup 在 hierarchy 中的路径相当于创建的 cgroup 目录相对于 root cgroup 目录的路径
	Path     string
	// 资源配置
	Resource *subsystems.Resources
}

func New(path string) *Manager {
	return &Manager{
		Path:     path,
	}
}

// Apply 将进程 PID 加入到每个 cgroup 中
func (m *Manager) Apply(pid int) error {
	for _, subSysIns := range subsystems.SubsystemsIns {
		err := subSysIns.Apply(m.Path, pid)
		if err != nil {
			return err
		}
	}
	return nil
}

// Set 设置 cgroup 资源限制
func (m *Manager) Set(res *subsystems.Resources) error {
	for _, subSysIns := range subsystems.SubsystemsIns {
		err := subSysIns.Set(m.Path, res)
		if err != nil {
			return err
		}
	}
	return nil
}

// Destroy 释放 cgroup
func (m *Manager) Destroy() error {
	for _, subSysIns := range subsystems.SubsystemsIns {
		if err := subSysIns.Remove(m.Path); err != nil {
			log.Warnf("remove cgroup fail %v", err)
		}
	}
	return nil
}
