package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

const cgroupMemoryHierarchMount = "/sys/fs/cgroup/memory"

func main() {
	if os.Args[0] == "/proc/self/exe" {
		// container process
		log.Printf("container pid: %d\n", syscall.Getpid())
		cmd := exec.Command("sh", "-c", "stress --vm-bytes 200m --vm-keep -m 1")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatalln(err)
		}
	}
	log.Print("main\n")
	cmd := exec.Command("/proc/self/exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Fatalln(err)
	} else {
		log.Printf("parrent pid: %d\n", cmd.Process.Pid)
		// 在系统默认创建挂载了 memory subsystem 的 Hierarchy 上创建 cgroup
		_ = os.Mkdir(filepath.Join(cgroupMemoryHierarchMount, "testmemorylimit"), 0755)
		// 将容器进程 pid 加入到 cgroup
		_ = ioutil.WriteFile(filepath.Join(cgroupMemoryHierarchMount, "testmemorylimit", "tasks"),
			[]byte(strconv.Itoa(cmd.Process.Pid)), 0644)
		// 限制该容器进程的内存使用
		_ = ioutil.WriteFile(filepath.Join(cgroupMemoryHierarchMount, "testmemorylimit", "memory.limit_in_bytes"),
			[]byte("300m"), 0644)

		_, _ = cmd.Process.Wait()
		// _ = cmd.Wait()
	}
}

// [root@shcCDFrh75vm7 7_cgroup]# go run main.go
// 2021/08/18 17:33:40 main
// 2021/08/18 17:33:40 parrent pid: 6649
// 2021/08/18 17:33:40 container pid: 1
// stress: info: [6] dispatching hogs: 0 cpu, 0 io, 1 vm, 0 hdd
//
// 如果将 300m 改为 100m


// https://github.com/xianlubird/mydocker/issues/69
// [root@shcCDFrh75vm7 memory]# cat /sys/fs/cgroup/memory/memory.oom_control
// oom_kill_disable 0
// under_oom 0
// oom_kill_disable 0 表示启动 OOM-killer。当内核无法给进程分配足够的内存时，将会直接 kill 掉该进程。oom_kill_disable 1，表示不启动 OOM-killer，
// 当内核无法给进程分配足够的内存时，将会暂停该进程直到有空余的内存之后再继续运行
// under_oom 字段，用来表示当前是否已经进入 oom 状态，也即是否有进程被暂停了。
// [root@shcCDFrh75vm7 memory]# cat /sys/fs/cgroup/memory/memory.swappiness
// 30
// [root@shcCDFrh75vm7 7_cgroup]# go run main.go
// 2021/08/18 18:01:05 main
// 2021/08/18 18:01:05 parrent pid: 7719
// 2021/08/18 18:01:05 container pid: 1
// stress: info: [6] dispatching hogs: 0 cpu, 0 io, 1 vm, 0 hdd
// stress: FAIL: [6] (415) <-- worker 7 got signal 9
// stress: WARN: [6] (417) now reaping child worker processes
// stress: FAIL: [6] (421) kill error: No such process
// stress: FAIL: [6] (451) failed run completed in 0s
// 由于内存限制，进程退出了
