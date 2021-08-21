package c

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/shipengqi/container/internal/container"
	_ "github.com/shipengqi/container/internal/nsenter"
	"github.com/shipengqi/container/pkg/log"
)

const (
	EnvExecPid = "container_pid"
	EnvExecCmd = "container_cmd"
)

func newExecCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "exec [options]",
		Short: "Run a command in a running container.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// This is for callback
			if os.Getenv(EnvExecPid) != "" {
				log.Infof("pid callback pid %s", os.Getgid())
				return nil
			}

			if len(args) < 2 {
				return errors.New("missing container id or command")
			}
			containerId := args[0]

			ExecContainer(containerId, args[1:])
			return nil
		},
	}
	c.DisableFlagsInUseLine = true
	return c
}

func ExecContainer(containerId string, cmdArray []string) {
	pid, err := getContainerPid(containerId)
	if err != nil {
		log.Errorf("getContainerPid %s error %v", containerId, err)
		return
	}
	cmdStr := strings.Join(cmdArray, " ")
	log.Infof("container pid %s", pid)
	log.Infof("command %s", cmdStr)

	// 这里目的就是为了那段 C 代码的执行。因为一旦程序启动，那段 C 代码就会自动运行，
	// 那么对于使用 exec 来说，当容器名和对应的命令传递进来以后，程序己经执行了，而且
	// 那段 C 代码也应该运行完毕。那么，怎么指定环境变量让它再执行一遍呢？这里就用到了这
	// 个 /proc/self/exe 。这里又创建了一个 command ，只不过这次只是简单地 fork 出来一个进程，
	// 然后把这个进程的标准输入输出都绑定到宿主机上。
	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = os.Setenv(EnvExecPid, pid)
	err = os.Setenv(EnvExecCmd, cmdStr)

	// 去 run 这里的进程时，实际上就是又运行了一遍自己的程序，但是这时有一点不同的就是，
	// 再一次运行的时候已经指定了环境变量，所以 C 代码执行的时候就能拿到对应的环境变量，
	// 便可以进入到指定的 Namespace 中进行操作了。
	if err := cmd.Run(); err != nil {
		log.Errorf("Exec container %s error %v", containerId, err)
	}
}

func getContainerPid(containerId string) (string, error) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	configFilePath := dirURL + container.ConfigName
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return "", err
	}
	var info container.Information
	if err := json.Unmarshal(contentBytes, &info); err != nil {
		return "", err
	}
	return info.Pid, nil
}

// [root@shcCDFrh75vm7 container]# ./container run -it -m 100m --cpus 1 /bin/sh
// 2021-08-22T07:53:31.425+0800	INFO	running: /bin/sh
// 2021-08-22T07:53:31.426+0800	INFO	running: [/bin/sh]
// 2021-08-22T07:53:31.426+0800	DEBUG	***** RUN Run *****
// 2021-08-22T07:53:31.558+0800	DEBUG	container id: 2665239328, name: 2665239328
// 2021-08-22T07:53:31.561+0800	INFO	send cmd: /bin/sh
// 2021-08-22T07:53:31.561+0800	INFO	send cmd: /bin/sh success
// 2021-08-22T07:53:31.561+0800	INFO	tty true
// 2021-08-22T07:53:31.563+0800	INFO	initializing
// 2021-08-22T07:53:31.564+0800	DEBUG	setting mount
// 2021-08-22T07:53:31.564+0800	DEBUG	pwd: /root/mnt
// 2021-08-22T07:53:31.622+0800	DEBUG	find cmd path: /bin/sh
// 2021-08-22T07:53:31.622+0800	DEBUG	syscall.Exec cmd path: /bin/sh
// / #
// 打开一个新的 terminal
// [root@shcCDFrh75vm7 container]# ./container exec 2665239328 /bin/sh
// 2021-08-22T07:53:47.860+0800	INFO	container pid 17202
// 2021-08-22T07:53:47.860+0800	INFO	command /bin/sh
// / # /bin/ls
// bin          dev          etc          home         lib          lib64        proc         root         sys          tmp          usr          var          version.txt
// / # exit