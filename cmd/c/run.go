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
	f.StringVar(&o.Network, "network", "default", "Connect a container to a network")
	f.StringSliceVarP(&o.Publish, "publish", "p", nil, "Publish a container's port(s) to the host")
	return c
}

// [root@shcCDFrh75vm7 container]# ./container network create --driver bridge --subnet 192.168.99.0/24 testbridge
// 2021-08-29T09:57:31.296+0800	DEBUG	subnet: 192.168.99.1/24
// [root@shcCDFrh75vm7 container]# ./container run -it -m 100m --cpus 1  --network testbridge  busybox /bin/sh
// 2021-08-29T09:57:44.391+0800	INFO	image name: busybox
// 2021-08-29T09:57:44.391+0800	INFO	command: [/bin/sh]
// 2021-08-29T09:57:44.391+0800	DEBUG	***** RUN Run *****
// 2021-08-29T09:57:44.435+0800	DEBUG	container id: 0790379253, name: 0790379253
// 2021-08-29T09:57:44.440+0800	DEBUG	load network: /var/run/q.container/network/network/testbridge
// 2021-08-29T09:57:44.441+0800	DEBUG	&{testbridge bridge 192.168.99.1/24}
// 2021-08-29T09:57:44.441+0800	DEBUG	ip: 192.168.99.1
// 2021-08-29T09:57:44.441+0800	DEBUG	mask: ffffff00
// 2021-08-29T09:57:44.441+0800	DEBUG	alloc ip: 192.168.99.2
// 2021-08-29T09:57:44.442+0800	INFO	initializing
// 2021-08-29T09:57:44.499+0800	DEBUG	ep ip: 192.168.99.2
// 2021-08-29T09:57:44.499+0800	DEBUG	ep ip range: {192.168.99.1 ffffff00}
// 2021-08-29T09:57:44.499+0800	INFO	send cmd: /bin/sh
// 2021-08-29T09:57:44.499+0800	INFO	send cmd: /bin/sh success
// 2021-08-29T09:57:44.499+0800	INFO	tty true
// 2021-08-29T09:57:44.500+0800	DEBUG	setting mount
// 2021-08-29T09:57:44.500+0800	DEBUG	pwd: /root/mnt/0790379253
// 2021-08-29T09:57:44.520+0800	DEBUG	find cmd path: /bin/sh
// 2021-08-29T09:57:44.520+0800	DEBUG	syscall.Exec cmd path: /bin/sh
// / # /bin/ifconfig
// cif-07903 Link encap:Ethernet  HWaddr 0E:35:2E:66:F1:B7
//          inet addr:192.168.99.2  Bcast:192.168.99.255  Mask:255.255.255.0
//          inet6 addr: fe80::c35:2eff:fe66:f1b7/64 Scope:Link
//          UP BROADCAST RUNNING MULTICAST  MTU:1500  Metric:1
//          RX packets:8 errors:0 dropped:0 overruns:0 frame:0
//          TX packets:8 errors:0 dropped:0 overruns:0 carrier:0
//          collisions:0 txqueuelen:1000
//          RX bytes:648 (648.0 B)  TX bytes:648 (648.0 B)
//
// lo        Link encap:Local Loopback
//          inet addr:127.0.0.1  Mask:255.0.0.0
//          inet6 addr: ::1/128 Scope:Host
//          UP LOOPBACK RUNNING  MTU:65536  Metric:1
//          RX packets:0 errors:0 dropped:0 overruns:0 frame:0
//          TX packets:0 errors:0 dropped:0 overruns:0 carrier:0
//          collisions:0 txqueuelen:1000
//          RX bytes:0 (0.0 B)  TX bytes:0 (0.0 B)
// / # /bin/route
// Kernel IP routing table
// Destination     Gateway         Genmask         Flags Metric Ref    Use Iface
// default         192.168.99.1    0.0.0.0         UG    0      0        0 cif-07903
// 192.168.99.0    *               255.255.255.0   U     0      0        0 cif-07903
// / #
// 可以看到这个容器有两个网卡设备，IP 地址是 192.168.99.2
// 启动另一个容器，并在另外一个容器中连接这个容器
// [root@shcCDFrh75vm7 container]# ./container run -it -m 100m --cpus 1  --network testbridge  busybox /bin/sh
// 2021-08-29T10:01:31.107+0800	INFO	image name: busybox
// 2021-08-29T10:01:31.107+0800	INFO	command: [/bin/sh]
// 2021-08-29T10:01:31.107+0800	DEBUG	***** RUN Run *****
// 2021-08-29T10:01:31.163+0800	DEBUG	container id: 1534611495, name: 1534611495
// 2021-08-29T10:01:31.166+0800	INFO	initializing
// 2021-08-29T10:01:31.168+0800	DEBUG	load network: /var/run/q.container/network/network/testbridge
// 2021-08-29T10:01:31.168+0800	DEBUG	&{testbridge bridge 192.168.99.1/24}
// 2021-08-29T10:01:31.168+0800	DEBUG	ip: 192.168.99.1
// 2021-08-29T10:01:31.168+0800	DEBUG	mask: ffffff00
// 2021-08-29T10:01:31.168+0800	DEBUG	alloc ip: 192.168.99.3
// 2021-08-29T10:01:31.204+0800	DEBUG	ep ip: 192.168.99.3
// 2021-08-29T10:01:31.205+0800	DEBUG	ep ip range: {192.168.99.1 ffffff00}
// 2021-08-29T10:01:31.205+0800	INFO	send cmd: /bin/sh
// 2021-08-29T10:01:31.206+0800	INFO	send cmd: /bin/sh success
// 2021-08-29T10:01:31.206+0800	INFO	tty true
// 2021-08-29T10:01:31.207+0800	DEBUG	setting mount
// 2021-08-29T10:01:31.208+0800	DEBUG	pwd: /root/mnt/1534611495
// 2021-08-29T10:01:31.224+0800	DEBUG	find cmd path: /bin/sh
// 2021-08-29T10:01:31.225+0800	DEBUG	syscall.Exec cmd path: /bin/sh
// / # /bin/ifconfig
// cif-15346 Link encap:Ethernet  HWaddr 06:83:87:1F:24:72
//          inet addr:192.168.99.3  Bcast:192.168.99.255  Mask:255.255.255.0
//          inet6 addr: fe80::483:87ff:fe1f:2472/64 Scope:Link
//          UP BROADCAST RUNNING MULTICAST  MTU:1500  Metric:1
//          RX packets:8 errors:0 dropped:0 overruns:0 frame:0
//          TX packets:8 errors:0 dropped:0 overruns:0 carrier:0
//          collisions:0 txqueuelen:1000
//          RX bytes:648 (648.0 B)  TX bytes:648 (648.0 B)
//
// lo        Link encap:Local Loopback
//          inet addr:127.0.0.1  Mask:255.0.0.0
//          inet6 addr: ::1/128 Scope:Host
//          UP LOOPBACK RUNNING  MTU:65536  Metric:1
//          RX packets:0 errors:0 dropped:0 overruns:0 frame:0
//          TX packets:0 errors:0 dropped:0 overruns:0 carrier:0
//          collisions:0 txqueuelen:1000
//          RX bytes:0 (0.0 B)  TX bytes:0 (0.0 B)
//
// / # /bin/ping 192.168.99.2
// PING 192.168.99.2 (192.168.99.2): 56 data bytes
// 64 bytes from 192.168.99.2: seq=0 ttl=64 time=0.299 ms
// 64 bytes from 192.168.99.2: seq=1 ttl=64 time=0.159 ms
// 64 bytes from 192.168.99.2: seq=2 ttl=64 time=0.176 ms
// 64 bytes from 192.168.99.2: seq=3 ttl=64 time=0.163 ms
// ^C
// --- 192.168.99.2 ping statistics ---
// 4 packets transmitted, 4 packets received, 0% packet loss
// round-trip min/avg/max = 0.159/0.199/0.299 ms
// / #
// 两个容器的网络是通的
// 在容器内访问外部网络，例如 www.baidu.com，但是由于容器内于默认没有配置 DNS 服务器，所以先来配置一下 DNS 服务器
// / # echo "nameserver 16.187.185.201" > /etc/resolv.conf
// / # /bin/ping ping www.baidu.com
//
// 正在 Ping www.wshifen.com [119.63.197.139] 具有 32 字节的数据:
// 来自 119.63.197.139 的回复: 字节=32 时间=79ms TTL=51
// 来自 119.63.197.139 的回复: 字节=32 时间=74ms TTL=51
// 来自 119.63.197.139 的回复: 字节=32 时间=81ms TTL=51
// 来自 119.63.197.139 的回复: 字节=32 时间=79ms TTL=51
//
// 119.63.197.139 的 Ping 统计信息:
//    数据包: 已发送 = 4，已接收 = 4，丢失 = 0 (0% 丢失)，
// 往返行程的估计时间(以毫秒为单位):
//    最短 = 74ms，最长 = 81ms，平均 = 78ms
// 容器映射端口到宿主机上供外部访问
// [root@shcCDFrh75vm7 container]# ./container run -it -m 100m --cpus 1 -p 80:80 --network testbridge  busybox /bin/sh
// 2021-08-29T10:23:05.909+0800	INFO	image name: busybox
// 2021-08-29T10:23:05.910+0800	INFO	command: [/bin/sh]
// 2021-08-29T10:23:05.910+0800	DEBUG	***** RUN Run *****
// 2021-08-29T10:23:05.957+0800	DEBUG	container id: 4674326516, name: 4674326516
// 2021-08-29T10:23:05.962+0800	DEBUG	load network: /var/run/q.container/network/network/testbridge
// 2021-08-29T10:23:05.962+0800	DEBUG	&{testbridge bridge 192.168.99.1/24}
// 2021-08-29T10:23:05.962+0800	DEBUG	ip: 192.168.99.1
// 2021-08-29T10:23:05.962+0800	DEBUG	mask: ffffff00
// 2021-08-29T10:23:05.962+0800	DEBUG	alloc ip: 192.168.99.6
// 2021-08-29T10:23:05.964+0800	INFO	initializing
// 2021-08-29T10:23:05.997+0800	DEBUG	ep ip: 192.168.99.6
// 2021-08-29T10:23:05.997+0800	DEBUG	ep ip range: {192.168.99.1 ffffff00}
// 2021-08-29T10:23:06.037+0800	INFO	send cmd: /bin/sh
// 2021-08-29T10:23:06.037+0800	INFO	send cmd: /bin/sh success
// 2021-08-29T10:23:06.037+0800	INFO	tty true
// 2021-08-29T10:23:06.037+0800	DEBUG	setting mount
// 2021-08-29T10:23:06.037+0800	DEBUG	pwd: /root/mnt/4674326516
// 2021-08-29T10:23:06.048+0800	DEBUG	find cmd path: /bin/sh
// 2021-08-29T10:23:06.048+0800	DEBUG	syscall.Exec cmd path: /bin/sh
// / # /bin/nc -lp 80
// hello
// [root@SGDLITVM0905 ~]# telnet shcCDFrh75vm7.hpeswlab.net 80
// Trying 16.155.197.5...
// Connected to shcCDFrh75vm7.hpeswlab.net.
// Escape character is '^]'.
// hello
