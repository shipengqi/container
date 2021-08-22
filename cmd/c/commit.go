package c

import (
	"fmt"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/shipengqi/container/internal/container"
	"github.com/spf13/cobra"

	"github.com/shipengqi/container/pkg/log"
)


func newCommitCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "commit [options]",
		Short:   "Create a new image from a container's changes.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return errors.New("missing container id os image name")
			}
			log.Info("committing")
			CommitContainer(args[0], args[1])
			return nil
		},
	}
	c.DisableFlagsInUseLine = true
	return c
}

func CommitContainer(containerId, imgName string) {
	rootfs := fmt.Sprintf(container.MntUrl, containerId)
	imageTar := "/root/" + imgName + ".tar"
	log.Debugf("commit: %s",imageTar)
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", rootfs, ".").CombinedOutput(); err != nil {
		log.Errorf("Tar folder %s error %v", rootfs, err)
	}
}

// [root@shcCDFrh75vm7 container]# ./container run -d -m 100m --cpus 1 -v /root/q.container.volume:/testvolume busybox /bin/sh
// 2021-08-22T10:25:32.728+0800	INFO	image name: busybox
// 2021-08-22T10:25:32.729+0800	INFO	command: [/bin/sh]
// 2021-08-22T10:25:32.729+0800	DEBUG	***** RUN Run *****
// 2021-08-22T10:25:32.786+0800	DEBUG	volume container url: /root/mnt/8099746273/testvolume
// 2021-08-22T10:25:32.786+0800	DEBUG	volume dirs: lowerdir=/root/q.container.volume.ro,upperdir=/root/q.container.volume,workdir=/root/q.container.work
// 2021-08-22T10:25:33.049+0800	DEBUG	container id: 8099746273, name: 8099746273
// 2021-08-22T10:25:33.056+0800	INFO	send cmd: /bin/sh
// 2021-08-22T10:25:33.056+0800	INFO	send cmd: /bin/sh success
// 2021-08-22T10:25:33.056+0800	INFO	tty false
// 2021-08-22T10:25:33.059+0800	WARN	remove cgroup fail unlinkat /sys/fs/cgroup/cpuset/q.container.cgroup/cpuset.memory_spread_slab: operation not permitted
// 2021-08-22T10:25:33.059+0800	WARN	remove cgroup fail unlinkat /sys/fs/cgroup/memory/q.container.cgroup/memory.kmem.tcp.max_usage_in_bytes: operation not permitted
// 2021-08-22T10:25:33.060+0800	WARN	remove cgroup fail unlinka
// 打开一个新的 terminal
// [root@shcCDFrh75vm7 container]# ./container commit 8099746273 busytest1
// 2021-08-22T10:26:36.124+0800	INFO	committing
// 2021-08-22T10:26:36.124+0800	DEBUG	commit: /root/busytest1.tar
// [root@shcCDFrh75vm7 container]# cd /root/untar
// [root@shcCDFrh75vm7 untar]# tar -xvf /root/busytest1.tar
// ...
// [root@shcCDFrh75vm7 untar]# ls
// bin  dev  etc  home  lib  lib64  proc  root  sys  testvolume  tmp  usr  var  version.txt
// [root@shcCDFrh75vm7 untar]# cd testvolume/
// [root@shcCDFrh75vm7 testvolume]# ls
// 1.txt
// [root@shcCDFrh75vm7 testvolume]# cd ~
// [root@shcCDFrh75vm7 ~]# ls /var/run/q.container/
// 6730420388  8099746273
// [root@shcCDFrh75vm7 ~]# ls /root/mnt
// 6730420388  8099746273  bin  dev  etc  home  lib  lib64  proc  root  sys  tmp  usr  var  version.txt
// [root@shcCDFrh75vm7 ~]# ls /root/mnt/6730420388
// bin  dev  etc  home  lib  lib64  proc  root  sys  tmp  usr  var  version.txt
// [root@shcCDFrh75vm7 ~]# ls /root/writeLayer/
// 6730420388  8099746273