package container

import (
	"github.com/shipengqi/container/pkg/log"
	"go.uber.org/zap"
	"os"
	"syscall"
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
func InitProcess(command string, args []string) error {
	mflags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	err := syscall.Mount("proc", "/proc", "proc", uintptr(mflags), "")
	if err != nil {
		log.Errort("mount", zap.Error(err))
		return err
	}
	// 这个版本会出现下面的问题
	// https://github.com/xianlubird/mydocker/issues/8
	// 必须执行 mount -t proc proc /proc

	// 这里命令要用绝路径，否则找不到


	// 下面的代码可以直接运行命令，例如： ./container run -it top
	// argv := []string{"/bin/sh", "-c", command}
	// log.Debugf("command: %s, args: %v", command, argv)
	// err = syscall.Exec("/bin/sh", argv, os.Environ())

	// 下面的代码可以执行 ./container run -it /bin/sh
	argv := []string{command}
	log.Debugf("command: %s, args: %v", command, argv)
	err = syscall.Exec(command, argv, os.Environ())

	if err != nil {
		log.Errort("syscall.Exec", zap.Error(err))
		return err
	}
	return nil
}
