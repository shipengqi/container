package action

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	"github.com/shipengqi/container/pkg/log"
)

const (
	EnvExecPid = "container_pid"
	EnvExecCmd = "container_cmd"
)

type execA struct {
	*action

	cmdArray    []string
	containerId string
}

func NewExecAction(containerId string, cmdArray []string) Interface {
	return &execA{
		action: &action{
			name: "exec",
		},
		containerId: containerId,
		cmdArray:    cmdArray,
	}
}

func (a *execA) Run() error {
	pid, err := getContainerPid(a.containerId)
	if err != nil {
		return errors.Errorf("get container pid: %s, %v", a.containerId, err)
	}
	cmdStr := strings.Join(a.cmdArray, " ")
	log.Infof("container pid %s", pid)
	log.Infof("command %s", cmdStr)

	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = os.Setenv(EnvExecPid, pid)
	err = os.Setenv(EnvExecCmd, cmdStr)

	containerEnvs := getEnvsByPid(pid)
	cmd.Env = append(os.Environ(), containerEnvs...)

	if err := cmd.Run(); err != nil {
		log.Errorf("Exec container %s error %v", a.containerId, err)
		return errors.Errorf("exec container: %s, %v", a.containerId, err)
	}
	return nil
}

func getContainerPid(containerId string) (string, error) {
	info, err := getContainerInfoById(containerId)
	if err != nil {
		return "", err
	}
	return info.Pid, nil
}

func getEnvsByPid(pid string) []string {
	path := fmt.Sprintf("/proc/%s/environ", pid)
	contentBytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Warnf("Read file %s error %v", path, err)
		return nil
	}
	envs := strings.Split(string(contentBytes), "\u0000")
	return envs
}
