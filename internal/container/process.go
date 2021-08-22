package container

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/shipengqi/container/pkg/log"
	"github.com/shipengqi/container/pkg/utils"
)

/*
NewParentProcess create a parent process
1. exec /proc/self/exe 就是自己调用自己，/proc/self 就是本进程的运行环境
2. init 是传给本进程的第一个参数，意味着要调用 init command 去初始化环境和资源
3. clone 参数就是 fork 出来一个 namespace 隔离的进程环境
4. tty enabled，就把当前进程的输入输出导入到标准输入输出
*/
func NewParentProcess(tty bool, volume, containerId, imgName string, envs []string) (*exec.Cmd, *os.File, error) {
	rp, wp, err := NewPipe()
	if err != nil {
		log.Errorf("new pipe: %v", err)
		return nil, nil, err
	}

	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		// 将容器进程的标准输出挂载到 /var/run/q.container/<containerId>/container.log 文件中
		dirURL := fmt.Sprintf(DefaultInfoLocation, containerId)
		if utils.IsNotExist(dirURL) {
			if err := os.MkdirAll(dirURL, 0622); err != nil {
				log.Errorf("NewParentProcess mkdir %s error %v", dirURL, err)
				return nil, nil, err
			}
		}
		stdLogFilePath := dirURL + LogFileName
		stdLogFile, err := os.Create(stdLogFilePath)
		if err != nil {
			log.Errorf("NewParentProcess create file %s error %v", stdLogFilePath, err)
			return nil, nil, err
		}
		cmd.Stdout = stdLogFile
	}

	cmd.ExtraFiles = []*os.File{rp}
	// os.Environ 环境变量就是系统默认的配置
	// 默认情况下，新启动进程的环境变量都是继承于原来父进程的环境变量，但是如果手动指定了环境变量，那么这里就会覆盖掉原来继承自父进程的变量。
	// 由于在容器的进程中，有时候还需要使用原来父进程的环境变量，比如 PATH 等，因此这里会使用 os.Environ 来获
	// 取宿主机的环境变量，然后把自定义的变量加进去
	// Environ returns a copy of strings representing the environment,
	// in the form "key=value".
	cmd.Env = append(os.Environ(), envs...)
	mntUrl, err := NewWorkSpace(volume, imgName, containerId)
	if err != nil {
		log.Errorf("workspace: %v", err)
		return nil, nil, err
	}
	cmd.Dir = mntUrl
	return cmd, wp, nil
}
