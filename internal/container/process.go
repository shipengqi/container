package container

import (
	"os"
	"os/exec"
	"syscall"
)

/*
NewParentProcess create a parent process
1. exec /proc/self/exe 就是自己调用自己，/proc/self 就是本进程的运行环境
2. init 是传给本进程的第一个参数，意味着要调用 init command 去初始化环境和资源
3. clone 参数就是 fork 出来一个 namespace 隔离的进程环境
4. tty enabled，就把当前进程的输入输出导入到标准输入输出
*/
func NewParentProcess(tty bool, command string) *exec.Cmd {
	args := []string{"init", command}
	cmd := exec.Command("/proc/self/exe", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd
}
