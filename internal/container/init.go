package container

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"github.com/shipengqi/container/pkg/log"
	"go.uber.org/zap"
)

/*
InitProcess 这个函数在容器内部执行，这是容器执行的第一个进程
使用 mount 挂在 proc 文件系统
MS_NOEXEC 在本文件系统中不允许运行其他程序
MS_NOSUID 在本系统中运行程序的时候，不允许 set-user-ID 或 set-group-ID
MS_NODEV  Linux 2.4 以来，所有 mount 的系统都会默认设定的参数

容器创建之后，执行的第一个进程并不是用户的进程，而是 init 初始化的进程如果通过 ps 命令查看就会发现，容器内第一个进程变成了
自己的 init，这和预想的是不一样的。并且 pid 1 的进程不能 kill，kill 的话容器进程就退出了。
syscall.Exec 底层调用了 int execve(const char *filename, char *const argv[], char *const envp[])
这个函数的作用是执行当前 filename 对应的程序。它会覆盖当前进程的镜像、数据和堆械等信息，包括 PID ， 这些都会被将要运行的进程覆盖掉。
也就是说，调用这个方法，将用户指定的进程运行起来，把最初的 init 进程给替换掉，这样当进入到容器内部的时候，就会发现容器内的第一个程序就是我们
指定的进程了。这是 runc 的实现方式之一。
*/
func InitProcess() error {
	cmdArgs, err := readParentPipe()
	if err != nil {
		log.Errorf("read parent pipe: %s", err)
		return err
	}
	if len(cmdArgs) < 1 {
		return errors.New("user command not found")
	}

	log.Debugf("setting mount")
	err = setUpMount()
	if err != nil {
		return err
	}
	// exec.LookPath 寻找绝对路径
	log.Debugf("find cmd path: %s", cmdArgs[0])
	cmdPath, err := exec.LookPath(cmdArgs[0])
	if err != nil {
		log.Errort("exec.LookPath", zap.Error(err))
		return err
	}

	log.Debugf("syscall.Exec cmd path: %s", cmdPath)
	err = syscall.Exec(cmdPath, cmdArgs[0:], os.Environ())

	if err != nil {
		log.Errort("syscall.Exec", zap.Error(err))
		return err
	}
	return nil
}

func readParentPipe() ([]string, error) {
	// uintptr(3)，就是指 index 为 3 的文件描述符
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		log.Errorf("read pipe: %v", err)
		return nil, err
	}
	return strings.Split(string(msg), " "), nil
}
