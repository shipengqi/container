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
		Cloneflags: syscall.CLONE_NEWUTS, // 创建一个 UTS Namespace
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

// [root@shcCDFrh75vm7 1_uts]# go run main.go
// sh-4.2#
// 打开一个新的 terminal, pstree 查看系统中进程之间的关系
// ├─sshd(1873)─┬─sshd(5749)───bash(5768)───go(5880)─┬─main(5962)─┬─sh(5967)
// 返回
// sh-4.2# echo $$ // 当前 pid
// 5967
// sh-4.2# readlink /proc/5967/ns/uts
// uts:[4026532566]
// sh-4.2# readlink /proc/5962/ns/uts
// uts:[4026531838]
// 两个进程确实不在同一个 UTS Namespace 中
// sh-4.2# hostname
// test.pooky.net
// sh-4.2# hostname -b pooky.test // 修改 hostname，UTS namespace 对 hostname 做了隔离
// sh-4.2# hostname
// pooky.test
// 切换到另一个 terminal
// [root@shcCDFrh75vm7 ~]# hostname
// shcCDFrh75vm7.hpeswlab.net // hostname 没有受到影响

