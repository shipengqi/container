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
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWUSER | syscall.CLONE_NEWNET,
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

// 查看当前宿主机的网络设备
// [root@shcCDFrh75vm7 ~]# ifconfig
// ens32: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
//        inet 16.155.197.5  netmask 255.255.248.0  broadcast 16.155.199.255
//        inet6 fe80::6833:a88a:7b0f:285c  prefixlen 64  scopeid 0x20<link>
//        ether 00:50:56:b0:31:14  txqueuelen 1000  (Ethernet)
//        RX packets 535941  bytes 102351502 (97.6 MiB)
//        RX errors 0  dropped 0  overruns 0  frame 0
//        TX packets 776  bytes 96118 (93.8 KiB)
//        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0
//
// lo: flags=73<UP,LOOPBACK,RUNNING>  mtu 65536
//        inet 127.0.0.1  netmask 255.0.0.0
//        inet6 ::1  prefixlen 128  scopeid 0x10<host>
//        loop  txqueuelen 1000  (Local Loopback)
//        RX packets 102  bytes 8996 (8.7 KiB)
//        RX errors 0  dropped 0  overruns 0  frame 0
//        TX packets 102  bytes 8996 (8.7 KiB)
//        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0
//
// virbr0: flags=4099<UP,BROADCAST,MULTICAST>  mtu 1500
//        inet 192.168.122.1  netmask 255.255.255.0  broadcast 192.168.122.255
//        ether 52:54:00:b3:a0:49  txqueuelen 1000  (Ethernet)
//        RX packets 0  bytes 0 (0.0 B)
//        RX errors 0  dropped 0  overruns 0  frame 0
//        TX packets 0  bytes 0 (0.0 B)
//        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0
// [root@shcCDFrh75vm7 6_network]# go run main.go
// sh-4.2$ ifconfig
// sh-4.2$
// 没有网络设备，说明 network namespace 生效了
