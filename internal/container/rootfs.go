package container

import (
	"os"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"
	"github.com/shipengqi/container/pkg/log"
	"github.com/shipengqi/container/pkg/utils"
)

func pivotRoot(root string) error {
	/*
	  为了使当前 root 的老 root 和新 root 不在同一个文件系统下，我们把 root 重新 mount 了一次
	  bind mount 是把相同的内容换了一个挂载点的挂载方法
	*/
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return errors.Wrap(err, "mount rootfs to itself")
	}
	// 创建 rootfs/.pivot_root 用来存储 old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	if utils.IsNotExist(pivotDir) {
		if err := os.Mkdir(pivotDir, 0777); err != nil {
			return err
		}
	}

	// https://github.com/xianlubird/mydocker/issues/62
	// pivot_root: invalid argument
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return errors.Wrap(err, "mount /")
	}

	// pivot_root 到新的 rootfs, 现在老的 old_root 是挂载在 rootfs/.pivot_root
	// 挂载点现在依然可以在 mount 命令中看到
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return errors.Wrap(err, "syscall.PivotRoot")
	}
	// 修改当前的工作目录到根目录
	if err := syscall.Chdir("/"); err != nil {
		return errors.Wrap(err, "syscall.Chdir")
	}
	pivotDir = filepath.Join("/", ".pivot_root")
	// umount rootfs/.pivot_root
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return errors.Wrap(err, "unmount pivot_root")
	}
	// 删除临时文件夹
	return os.Remove(pivotDir)
}

func setUpMount() error {
	// 获取当前路径
	pwd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "os.Getwd")
	}
	log.Debugf("pwd: %s", pwd)
	err = pivotRoot(pwd)
	if err != nil {
		return err
	}

	// mount proc
	mflags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	err = syscall.Mount("proc", "/proc", "proc", uintptr(mflags), "")
	if err != nil {
		return errors.Wrap(err, "mount proc")
	}

	// tmpfs 是一种基于内存的文件系统，可以使用 RAM 或 swap 分区来存储。
	err = syscall.Mount("tmpfs", "/dev", "tmpfs",
		syscall.MS_NOSUID | syscall.MS_STRICTATIME, "mode=755")
	if err != nil {
		return errors.Wrap(err, "mount tmpfs")
	}

	return nil
}

// [root@shcCDFrh75vm7 container]# ./container run -it -m 100m --cpus 1 /bin/sh
// 2021-08-20T17:20:40.019+0800    INFO    running: /bin/sh
// 2021-08-20T17:20:40.020+0800    INFO    running: [/bin/sh]
// 2021-08-20T17:20:40.020+0800    DEBUG   ***** [RUN] PreRun *****
// 2021-08-20T17:20:40.020+0800    DEBUG   ***** RUN Run *****
// 2021-08-20T17:20:40.029+0800    INFO    send cmd: /bin/sh
// 2021-08-20T17:20:40.032+0800    INFO    initializing
// 2021-08-20T17:20:40.032+0800    DEBUG   pwd: /root/busybox
// 2021-08-20T17:20:40.136+0800    DEBUG   find cmd path: /bin/sh
// 2021-08-20T17:20:40.136+0800    DEBUG   syscall.Exec cmd path: /bin/sh
// / # /bin/ls -l
// total 12
// drwxr-xr-x    2 root     root          4096 Sep 22  2020 bin
// drwxr-xr-x    2 root     root            40 Aug 20 09:20 dev
// drwxr-xr-x    4 root     root           171 Aug 20 09:02 etc
// drwxr-xr-x    2 nobody   nogroup          6 Sep 22  2020 home
// drwxr-xr-x    2 root     root          4096 Sep 22  2020 lib
// lrwxrwxrwx    1 root     root             3 Sep 22  2020 lib64 -> lib
// dr-xr-xr-x  252 root     root             0 Aug 20 09:20 proc
// drwx------    2 root     root            26 Aug 20 09:19 root
// drwxr-xr-x    2 root     root             6 Aug 20 09:02 sys
// drwxrwxrwt    2 root     root             6 Sep 22  2020 tmp
// drwxr-xr-x    3 root     root            18 Sep 22  2020 usr
// drwxr-xr-x    4 root     root            30 Sep 22  2020 var
// -rw-r--r--    1 root     root            12 Sep 22  2020 version.txt
// / #
