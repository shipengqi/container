package container

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/pkg/errors"
	"github.com/shipengqi/container/pkg/log"
	"github.com/shipengqi/container/pkg/utils"
)

const stdioFdCount = 3


// NewParentProcess create a parent init process
func NewParentProcess(tty bool, volume, containerId, imgName string, envs []string) (*exec.Cmd, *os.File, error) {
	childInitRp, parentInitWp, err := NewPipe()
	if err != nil {
		log.Errorf("new pipe: %v", err)
		return nil, nil, errors.Wrap(err, "new init pipe")
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
		// mount stdout of container into /var/run/q.container/<containerId>/container.log
		dirURL := fmt.Sprintf(DefaultInfoLocation, containerId)
		if utils.IsNotExist(dirURL) {
			if err := os.MkdirAll(dirURL, 0622); err != nil {
				return nil, nil, errors.Errorf("mkdir: %s: %v", dirURL, err)
			}
		}
		containerStdLogFilePath := dirURL + LogFileName
		containerStdLogFile, err := os.Create(containerStdLogFilePath)
		if err != nil {
			return nil, nil, errors.Errorf("create: %s: %v", containerStdLogFilePath, err)
		}
		cmd.Stdout = containerStdLogFile
	}

	cmd.Env = append(os.Environ(), envs...)

	cmd.ExtraFiles = []*os.File{childInitRp}
	cmd.Env = append(cmd.Env,
		"_QCONTAINER_INITPIPE="+strconv.Itoa(stdioFdCount+len(cmd.ExtraFiles)-1),
	)

	mntUrl, err := NewWorkSpace(volume, imgName, containerId)
	if err != nil {
		return nil, nil, errors.Wrap(err, "new workspace")
	}
	cmd.Dir = mntUrl
	return cmd, parentInitWp, nil
}
