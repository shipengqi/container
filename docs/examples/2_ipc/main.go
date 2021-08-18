package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	cmd := exec.Command("sh")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC, // UTS 和 IPC Namespace
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

// 在宿主机上打开一个 terminal
// [root@shcCDFrh75vm7 ~]# ipcs -q // 查看当前的 ipc message queue
//
// ------ Message Queues --------
// key        msqid      owner      perms      used-bytes   messages
//
// [root@shcCDFrh75vm7 ~]# ipcmk -Q // 创建一个 message queue
// Message queue id: 0
// [root@shcCDFrh75vm7 ~]# ipcs -q // 查看当前的 ipc message queue
//
// ------ Message Queues --------
// key        msqid      owner      perms      used-bytes   messages
// 0xd8d34740 0          root       644        0            0
// 打开一个 terminal 运行 go run main.go
// [root@shcCDFrh75vm7 2_ipc]# go run main.go
// sh-4.2# ipcs -q
//
// ------ Message Queues --------
// key        msqid      owner      perms      used-bytes   messages
// 看不到宿主机上己经创建的 message queue ，说明 IPC Namespace 创建成功， IPC 已经被隔离。
