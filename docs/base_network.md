# 容器网络

Linux 是通过网络设备去操作和使用网卡的，系统装了一个网卡之后会为其生成一个网络设备实例，比如 ethO。而随着网络虚拟化技术的发展，Linux 支持创建出虚
拟化的设备，可以通过虚拟化设备的组合实现多种多样的功能和网络拓扑。常见的虚拟化设备有 Veth、Bridge、802.1.q VLAN device、TAP 等。

## Veth
Veth 是成对出现的虚拟网络设备，发送到 Veth 一端虚拟设备的请求会从另一端的虚拟设备中发出。在容器的虚拟化场景中，经常会使用 Veth 连接不同的网络 Namespace。

```bash
# 创建两个 network namespace
[root@shcCDFrh75vm7 ~]# ip netns add ns1
[root@shcCDFrh75vm7 ~]# ip netns add ns2
# 创建一对 veth 设备
[root@shcCDFrh75vm7 ~]# ip link add veth0 type veth peer name veth1
# 将两个 veth 设备移动到两个 network namespace 中
[root@shcCDFrh75vm7 ~]# ip link set veth0 netns ns1
[root@shcCDFrh75vm7 ~]# ip link set veth1 netns ns2
# 在 ns1 中查看网络设备
[root@shcCDFrh75vm7 ~]# ip netns exec ns1 ip link
1: lo: <LOOPBACK> mtu 65536 qdisc noop state DOWN mode DEFAULT group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
6: veth0@if5: <BROADCAST,MULTICAST> mtu 1500 qdisc noop state DOWN mode DEFAULT group default qlen 1000
    link/ether fa:af:3d:87:de:6d brd ff:ff:ff:ff:ff:ff link-netnsid 1
```

除了 lo 设备，就只有一个网络设备 veth0。当请求发送到 veth0 这个虚拟网络设备时，都会原封不动地从另外一个网络 Namespace ns2 的网络接口 veth1 中
出来。

```bash
# 配置每个 veth 设备的 ip 地址和 namespace 路由
[root@shcCDFrh75vm7 ~]# ip netns exec ns1 ifconfig veth0 172.18.0.2/24 up
[root@shcCDFrh75vm7 ~]# ip netns exec ns2 ifconfig veth1 172.18.0.3/24 up
[root@shcCDFrh75vm7 ~]# ip netns exec ns1 route add default dev veth0
[root@shcCDFrh75vm7 ~]# ip netns exec ns2 route add default dev veth1
# 通过 veth 一端出去的包，另外一端能直接接收到
[root@shcCDFrh75vm7 ~]# ip netns exec ns1 ping -c 3 172.18.0.3
PING 172.18.0.3 (172.18.0.3) 56(84) bytes of data.
64 bytes from 172.18.0.3: icmp_seq=1 ttl=64 time=0.114 ms
64 bytes from 172.18.0.3: icmp_seq=2 ttl=64 time=0.112 ms
64 bytes from 172.18.0.3: icmp_seq=3 ttl=64 time=0.083 ms

--- 172.18.0.3 ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 2000ms
rtt min/avg/max/mdev = 0.083/0.103/0.114/0.014 ms
```

## Bridge
Bridge 虚拟设备是用来桥接的网络设备，它相当于现实世界中的交换机，可以连接不同的网络设备，当请求到达 Bridge 设备时，可以通过报文中的 Mac 地址进行
广播或转发。

删除前面创建的 namespace：
```bash
[root@shcCDFrh75vm7 ~]# ip netns delete ns1
[root@shcCDFrh75vm7 ~]# ip netns delete ns2
```


```bash
# # 创建 1 个 network namespace
[root@shcCDFrh75vm7 ~]# ip netns add ns1
# 创建一对 veth 设备
[root@shcCDFrh75vm7 ~]# ip link add veth0 type veth peer name veth1
# 将 veth1 设备移动到 ns1 namespace 中
[root@shcCDFrh75vm7 ~]# ip link set veth1 netns ns1
# 创建网桥
[root@shcCDFrh75vm7 ~]# brctl addbr br0
# 挂载网络设备
[root@shcCDFrh75vm7 ~]# brctl addif br0 eth0
[root@shcCDFrh75vm7 ~]# brctl addif br0 veth0
```

## Linux 路由表
路由表是 Linux 内核的一个模块，通过定义路由表来决定在某个网络 Namespace 中包的流向，从而定义请求会到哪个网络设备上。

**如果是使用 ssh 远程连接的虚机，下面的操作会导致断开连接**。

```bash
# 启动虚拟网络设备，并设置 veth1 的 ip 地址
[root@shcCDFrh75vm7 ~]# ip link set veth0 up
[root@shcCDFrh75vm7 ~]# ip link set br0 up
[root@shcCDFrh75vm7 ~]# ip netns exec ns1 ifconfig veth1 172.18.0.2/24 up
# 分别设置 ns1 网络中的路由和宿主机的路由
# default 代表 0.0.0.0/0，即在 namespace 中所有流量都经过 veth1 流出
[root@shcCDFrh75vm7 ~]# ip netns exec ns1 route add default dev veth1
# 在宿主机上将 172.18.0.0/24 的网段请求路由到 br0 网桥
[root@shcCDFrh75vm7 ~]# route add -net 172.18.0.0/24 dev br0
# 查看宿主机 ip
[root@shcCDFrh75vm7 ~]# ifconfig eth0
ethO    Link encap:Ethernet HWaddr 08:00:27:0e:94:e9
        inet addr: 10.0.2.15 Bcast:10.0.2.255 Mask:255.255.255.0
        inet6 addr:fe80::a00:27ff:fe0e:94e9/64 Scope:Link
        UP BROADCAST RUNNING MULTICAST MTU:1500 Metric:1
        RX packets:4521 errors:O dropped:0 overruns:O frame:O
        TX packets:1028 errors:O dropped:O overruns:O carrier:O
        collisions:O txqueuelen:1OOO
        RX bytes:4959937 (4.9 MB) TX bytes:70270 (70.2 KB)
# 在 ns1 中访问宿主机地址
[root@shcCDFrh75vm7 ~]# ip netns exec ns1 ping -c 3 10.0.2.15
PING 10.0.2.15 (10.0.2.15) 56(84) bytes of data.
64 bytes from 10.0.2.15: icmp_seq=1 ttl=64 time=0.114 ms
64 bytes from 10.0.2.15: icmp_seq=2 ttl=64 time=0.112 ms
64 bytes from 10.0.2.15: icmp_seq=3 ttl=64 time=0.083 ms

--- 10.0.2.15 ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 2000ms
rtt min/avg/max/mdev = 0.083/0.103/0.114/0.014 ms
# 在宿主机访问 ns1 中的地址
[root@shcCDFrh75vm7 ~]# ping -c 3 172.18.0.2
[root@shcCDFrh75vm7 ~]# ip netns exec ns1 ping -c 3 10.0.2.15
PING 172.18.0.2 (172.18.0.2) 56(84) bytes of data.
64 bytes from 172.18.0.2: icmp_seq=1 ttl=64 time=0.114 ms
64 bytes from 172.18.0.2: icmp_seq=2 ttl=64 time=0.112 ms
64 bytes from 172.18.0.2: icmp_seq=3 ttl=64 time=0.083 ms

--- 172.18.0.2 ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 2000ms
rtt min/avg/max/mdev = 0.083/0.103/0.114/0.014 ms
```

## Linux iptables
iptables 是对 Linux 内核的 netfilter 模块进行操作的工具，用来管理网络包的流动和转发。

iptables 定义了一套链式处理结构，在网络报传输的各个阶段可以使用不同的策略对包进行加工，转发，丢弃。

在容器化技术中，经常会用到两种策略 MASQUERADE 和 DNAT，用户容器和宿主机外部的网络通信。

### MASQUERADE

MASQUERADE 策略可以将请求包中的元四肢转换成一个网络设备的地址。

比如 Namespace 中网络设备的地址是 `172.18.0.2`，这个地址虽然在宿主机上可以路由到 brO 的网桥，但是到
达宿主机外部之后，是不知道如何路由到这个 IP 地址的，所以如果请求外部地址的话，需要先通过 MASQUERADE 策略将这个 IP 转换成宿主机出口网
卡的 IP。

### DNAT

DNAT 策略也是做网络地址的转换，不过它是要更换目标地址，经常用于将内部网络地址的端口映射到外部去。比如，Namespace 中如果需要提供服务给
宿主机之外的应用去请求要怎么办呢？外部应用没办法直接路由到 `172.18.0.2` 这个地址，这时候就可以用到 DNAT 策略。

```bash
iptables -t nat -A PREROUTING -p tcp -m tcp --dport 80 -j DNAT --to-destination 172.18.0.2:80
```

这样就可以把宿主机上 80 端口的 TCP 请求转发到 Namespace 中的地址 `172.18.0.2:80`，从而实现外部的应用调用。


在容器的 Net Namespace 中，就可以通过容器的 Veth 直接与挂载在同一个 Bridge 上的容器通信，以及通过 Bridge 上创建的 iptables 的
MASQUERADE 规则访问外部网络，同时，外部也可以通过机器上的端口经过 iptables 的 DNAT 的转发访问容器内部。


### IPAM

IPAM 也是网络功能中的一个组件，用于网络 IP 地址的分配和释放，包括容器的 IP 地址和网络网关的 IP 地址。
