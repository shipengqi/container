package container

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/shipengqi/container/pkg/log"
)

/*
NewParentProcess create a parent process
1. exec /proc/self/exe 就是自己调用自己，/proc/self 就是本进程的运行环境
2. init 是传给本进程的第一个参数，意味着要调用 init command 去初始化环境和资源
3. clone 参数就是 fork 出来一个 namespace 隔离的进程环境
4. tty enabled，就把当前进程的输入输出导入到标准输入输出
*/
func NewParentProcess(tty bool) (*exec.Cmd, *os.File, error) {
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
	}

	// 方传入管道文件读取端的句柄
	// ExtraFiles 会带着这个文件句柄去创建子进程
	// 1 个进程默认会有 3 个文件描述符，分别是标准输入、标准输出、标准错误
	// 这 3 个是子进程一创建的时候就会默认带着的，那么外带的这个文件描述符理所当然地就成为了第 4 个
	// [root@shccdfrh75vm8 ~]# ll /proc/self/fd
	// total 0
	// lrwx------. 1 root root 64 Aug 20 13:48 0 -> /dev/pts/0
	// lrwx------. 1 root root 64 Aug 20 13:48 1 -> /dev/pts/0
	// lrwx------. 1 root root 64 Aug 20 13:48 2 -> /dev/pts/0
	// lr-x------. 1 root root 64 Aug 20 13:48 3 -> /proc/4887/fd
	cmd.ExtraFiles = []*os.File{rp}
	// 先把下载好的 busybox 放到宿主机的 /root/busybox 目录下，然后 cmd.Dir = "/root/busybox"
	// 会指定子进程初始化后的工作目录
	cmd.Dir = "/root/busybox"
	return cmd, wp, nil
}
