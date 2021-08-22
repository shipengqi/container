package c

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

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
		Short: "Run a command in a running container",
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

	// 获取对应的 PID 环境变量，也就是容器的环境变量
	containerEnvs := getEnvsByPid(pid)
	cmd.Env = append(os.Environ(), containerEnvs...)

	// 去 run 这里的进程时，实际上就是又运行了一遍自己的程序，但是这时有一点不同的就是，
	// 再一次运行的时候已经指定了环境变量，所以 C 代码执行的时候就能拿到对应的环境变量，
	// 便可以进入到指定的 Namespace 中进行操作了。
	if err := cmd.Run(); err != nil {
		log.Errorf("Exec container %s error %v", containerId, err)
	}
}

func getContainerPid(containerId string) (string, error) {
	info, err := getContainerInfoById(containerId)
	if err != nil {
		return "", err
	}
	return info.Pid, nil
}

// 修改 exec 命令来直接使用 env 命令查看容器内环境变量的功能。
func getEnvsByPid(pid string) []string {
	path := fmt.Sprintf("/proc/%s/environ", pid)
	contentBytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Errorf("Read file %s error %v", path, err)
		return nil
	}
	// [root@shcCDFrh75vm7 container]# cat /proc/self/environ
	// XDG_SESSION_ID=732HOSTNAME=shcCDFrh75vm7.hpeswlab.netSELINUX_ROLE_REQUESTED=TERM=xtermSHELL=/bin/bashHISTSIZE=1000
	// SSH_CLIENT=15.122.67.231 59439 22SELINUX_USE_CURRENT_RANGE=SSH_TTY=/dev/pts/0NO_PROXY=127.0.0.1,localhost,.hpe.com,.hp.com,.hpeswlab.net
	// http_proxy=http://web-proxy.jp.softwaregrp.net:8080
	// \u 开头的是一个 Unicode 码的字符,每一个 '\u0000' 都代表了一个空格。
	// env split by \u0000
	envs := strings.Split(string(contentBytes), "\u0000")
	return envs
}

// [root@shcCDFrh75vm7 container]# ./container exec 5867115283 /bin/sh
// 2021-08-22T11:11:57.864+0800	INFO	container pid 21164
// 2021-08-22T11:11:57.864+0800	INFO	command /bin/sh
// / # echo $test1
// 111
// / # echo $test2
// 222