package container

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"github.com/shipengqi/container/pkg/log"
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
	cmdArgs, err := readParentInitPipe()
	if err != nil {
		return errors.Wrap(err, "read parent pipe")
	}
	if len(cmdArgs) < 1 {
		return errors.New("user command not found")
	}

	log.Debugf("setting mount")
	err = setUpMount()
	if err != nil {
		return err
	}
	// exec.LookPath
	log.Debugf("find cmd path: %s", cmdArgs[0])
	cmdPath, err := exec.LookPath(cmdArgs[0])
	if err != nil {
		return errors.Wrap(err, "exec.LookPath")
	}

	log.Debugf("syscall.Exec cmd path: %s", cmdPath)
	err = syscall.Exec(cmdPath, cmdArgs[0:], os.Environ())

	if err != nil {
		return errors.Wrap(err, "syscall.Exec")
	}
	log.Debug("syscall.Exec done")
	return nil
}

func readParentInitPipe() ([]string, error) {
	initPipeFdStr, exists := os.LookupEnv("_QCONTAINER_INITPIPE")
	if !exists {
		panic("_QCONTAINER_INITPIPE not found")
	}
	initPipeFd, err := strconv.Atoi(initPipeFdStr)
	if err != nil {
		panic(fmt.Sprintf("_QCONTAINER_INITPIPE=%s to int: %s", initPipeFdStr, err))
	}

	pipe := os.NewFile(uintptr(initPipeFd), "initpipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		return nil, errors.Wrap(err, "read init pipe")
	}
	return strings.Split(string(msg), " "), nil
}

// closeParentLogPipe Close the log pipe fd so the parent's ForwardLogs can exit.
func closeParentLogPipe() {
	logPipeFdStr, exists := os.LookupEnv("_QCONTAINER_LOGPIPE")
	if !exists {
		panic("_QCONTAINER_INITPIPE not found")
	}
	logPipeFd, err := strconv.Atoi(logPipeFdStr)
	if err != nil {
		panic(fmt.Sprintf("_QCONTAINER_LOGPIPE=%s to int: %s", logPipeFdStr, err))
	}
	log.Debugf("_QCONTAINER_LOGPIPE=%s", logPipeFdStr)
	pipe := os.NewFile(uintptr(logPipeFd), "logpipe")
	err = pipe.Close()
	if err != nil {
		log.Debugf("Close %v", err)
	}
}
