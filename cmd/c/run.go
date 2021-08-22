package c

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/shipengqi/container/internal/action"
)


func newRunCmd() *cobra.Command {
	o := &action.RunActionOptions{}
	c := &cobra.Command{
		Use:   "run [options]",
		Short: "Create a container with namespace and cgroups limit.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return errors.New("missing image or container command")
			}
			a := action.NewRunAction(args, o)
			return action.Execute(a)
		},
	}

	c.Flags().SortFlags = false
	c.DisableFlagsInUseLine = true
	f := c.Flags()
	f.BoolVarP(&o.Interactive, "interactive", "i", false, "Keep STDIN open even if not attached")
	f.BoolVarP(&o.TTY, "tty", "t",false, "Allocate a pseudo-TTY")
	f.StringVarP(&o.MemoryLimit, "memory", "m","", "Memory limit")
	f.StringVar(&o.CpuSet, "cpus","", "Number of CPUs")
	f.StringVarP(&o.CpuShare, "cpu-shares", "c","", "CPU shares (relative weight)")
	f.StringVarP(&o.Volume, "volume", "v","", "Bind mount a volume")
	f.BoolVarP(&o.Detach, "detach", "d",false, "Run container in background and print container ID")
	f.StringVar(&o.Name, "name", "", "Assign a name to the container")
	f.StringSliceVarP(&o.Envs, "env", "e", nil, "Set environment variables")
	return c
}

// [root@shcCDFrh75vm7 container]# ./container run -it -m 100m --cpus 1 -e "test1=111" -e "test2=222" busybox /bin/sh
// 2021-08-22T11:00:50.305+0800	INFO	image name: busybox
// 2021-08-22T11:00:50.306+0800	INFO	command: [/bin/sh]
// 2021-08-22T11:00:50.306+0800	DEBUG	***** RUN Run *****
// 2021-08-22T11:00:50.448+0800	DEBUG	container id: 9103882410, name: 9103882410
// 2021-08-22T11:00:50.451+0800	INFO	send cmd: /bin/sh
// 2021-08-22T11:00:50.451+0800	INFO	send cmd: /bin/sh success
// 2021-08-22T11:00:50.451+0800	INFO	tty true
// 2021-08-22T11:00:50.454+0800	INFO	initializing
// 2021-08-22T11:00:50.455+0800	DEBUG	setting mount
// 2021-08-22T11:00:50.455+0800	DEBUG	pwd: /root/mnt/9103882410
// 2021-08-22T11:00:50.632+0800	DEBUG	find cmd path: /bin/sh
// 2021-08-22T11:00:50.632+0800	DEBUG	syscall.Exec cmd path: /bin/sh
// / # echo test1
// test1
// / # echo $test1
// 111
// / # echo $test2
// 222
// 但是打开一个新的 terminal
// [root@shcCDFrh75vm7 container]# ./container exec 9103882410 /bin/sh
// 2021-08-22T11:02:21.889+0800	INFO	container pid 20855
// 2021-08-22T11:02:21.890+0800	INFO	command /bin/sh
// / # echo $test1
//
// / #
// exec 进入容器后没有拿到环境变量，因为 exec 命令其实是 ./container 发起的另外一个进程，
// 这个进程的父进程其实是宿主机的，并不是容器内的。因为在 Cgo 里面使用了 setns 系统调用，
// 才使得这个进程进入到了容器内的命名空间，但是由于环境变量是继承自父进程的，因此这个 exec 进程的环境变量其实是
// 继承自宿主机的，所以在 exec 进程内看到的环境变量其实是宿主机的环境变量。但是，只要是容器内 PID 为 1 的进程，创建出来的进程
// 都会继承它的环境变量。