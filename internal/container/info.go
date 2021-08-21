package container

const (
	StatusRunning string = "running"
	StatusStop    string = "stopped"
	StatusExit    string = "exited"
)

var (
	DefaultInfoLocation = "/var/run/q.container/%s/"
	ConfigName          = "config.json"
)

type Information struct {
	Pid         string `json:"pid"`        // 容器的 init 进程在宿主机上的 PID
	Id          string `json:"id"`         // 容器 Id
	Name        string `json:"name"`       // 容器名
	Command     string `json:"command"`    // 容器内 init 运行命令
	CreatedTime string `json:"createTime"` // 创建时间
	Status      string `json:"status"`     // 容器的状态
}
