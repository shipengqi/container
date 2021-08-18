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
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}


// [root@shcCDFrh75vm7 3_pid]# go run main.go
// sh-4.2# echo $$
// 1
// 在宿主机上再打开一个 terminal
// [root@shcCDFrh75vm7 ~]# pstree -pl
//   ├─sshd(1873)─┬─sshd(5749)───bash(5768)───go(5880)─┬─main(5962)─┬─sh(5967)
//   │            │                                    │            ├─{main}(5963)
// 找到真实的 pid
// 可以看到 PID namespace 隔离了 pid，5962 这个 pid 被映射到了 namespace 中的 pid 1
