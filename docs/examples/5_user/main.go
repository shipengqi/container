package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

// https://github.com/xianlubird/mydocker/issues/3
// https://github.com/golang/go/issues/10626
// https://github.com/golang/go/issues/16283
// centos 默认的没有开启 user namespace, 开启命令
// grubby --args="user_namespace.enable=1" --update-kernel="$(grubby --default-kernel)"
// reboot
// 如果已经内核开启 user_namespace 依旧 invalid argument 的话，执行
// echo 640 > /proc/sys/user/max_user_namespaces
// https://unix.stackexchange.com/questions/479635/unable-to-create-user-namespace-in-rhel?rq=1
func main() {
	cmd := exec.Command("sh")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER,
		// Credential: &syscall.Credential{
		// 	Uid: uint32(1999),
		// 	Gid: uint32(1999),
		// },
		// Fix fork/exec /usr/bin/sh: operation not permitted
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 1999,
				HostID:      syscall.Getuid(),
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 1999,
				HostID:      syscall.Getgid(),
				Size:        1,
			},
		},
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

// 查看当前宿主机的用户和用户组
// [root@shcCDFrh75vm7 ~]# id
// uid=0(root) gid=0(root) groups=0(root)
// [root@shcCDFrh75vm7 5_user]# go run main.go
// sh-4.2$ id
// uid=1999 gid=1999 groups=1999
// uid 是不同的，说明 User Namespace 生效了
