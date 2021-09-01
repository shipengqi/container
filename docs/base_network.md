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

## Linux 路由表 和 route 命令
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

## ip 命令

### netns

`ip netns` 用来管理 `network namespace` 的。

## docker 的四种网络模型

- bridge 桥接模式
- host 模式
- container 模式
- none 模式

## 容器跨主机网络
Linux Bridge 是没有办法路由到宿主机外部的。所以，不同宿主机 Bridge 上的容器还是得通过映射到端口的方式来实现互相访问。那么，多个容器就没办法同
时使用同一个端口了，而且访问的地址也不是容器自身的 IP 地址，并且暴漏到宿主机端口上，可以直接提供宿主机的地址来访问容器的服务，也不安全。

如何做到容器跨主机网络？

### 跨主机容器网络的 IPAM
IPAM，通过将 IP 地址分配信息的位图存放在文件中，实现了容器和网关的 IP 地址分配。但通常情况下，没办法在多个宿主机上使用同一个文件来做 IP 地址分配，
如果每个机器只负责容器网络在自己宿主机上的 IP 地址分配，那么就可能会造成不同机器上分配到的容器 IP 地址重复的问题。如果同一个容器网络中的 IP 重复了，
就会导致不可预期的访问问题。

可以采用中心化的一致性 KV-store 来作为记录IP 地址分配的存储。对于跨主机的容器网络，把 IP 地址分配信息的位图存放在中心的一致性 KV-store 中，来保证每个宿主机
上分配给容器的 IP 不冲突。
常见的一致性 KV-store 有 etcd 、consul 、zookeeper 等，都可以用来实现跨主机网络的 IP 地址分配需求。

### 跨主机容器网络通信的常见实现方式

容器在宿主机上的 IP 是通过配置 route 的方式配置路由访问到的，但是在宿主机之间是访问不到这个地址的。那么如何实现跨主机容器之间的通信？

通常有两种实现方式：
- 封包，础设施要求低，只需要宿主机之间能联通即可，性能损耗大（带宽，资源）
- 路由，无封包，性能好，对基础网络设施有要求需要支持一些路由的配置或者特定的网络协议。

#### 封包
既然在宿主机之间不知道容器的地址要怎么路由和访问，就可以把容器之间的请求外面包装上宿主机的地址发送，这样跨主机的容器之间的通信就转换成了宿主机之间的通信，到达另
外一个容器所在的宿主机后再解开外面的包装，拿到真正的容器请求，就能实现跨主机的容器间通信了。

#### 路由
另一种实现跨主机容器网络通信的方式是路由，这种方式的原理是让宿主机的网络“知道”容器的地址要怎么路由、路由到哪台机器上，这种方式一般需要网络设备的支持，比如修
改路由器的路由表，将容器 IP 地址的下一跳修改到这个容器所在的宿主机上，来送达跨主机容器间的请求。


容器退出后必须清理 iptables 规则，否则会导致冲突
network 删除后必须清理 iptables 规则和路由表
```bash
[root@shcCDFrh75vm7 container]# iptables -L PREROUTING -t nat --line-number
Chain PREROUTING (policy ACCEPT)
num  target     prot opt source               destination
1    PREROUTING_direct  all  --  anywhere             anywhere
2    PREROUTING_ZONES_SOURCE  all  --  anywhere             anywhere
3    PREROUTING_ZONES  all  --  anywhere             anywhere
4    DNAT       tcp  --  anywhere             anywhere             tcp dpt:http to:192.168.99.6:80
5    DNAT       tcp  --  anywhere             anywhere             tcp dpt:http to:192.168.99.4:80
6    DNAT       tcp  --  anywhere             anywhere             tcp dpt:http to:192.168.99.5:80
7    DNAT       tcp  --  anywhere             anywhere             tcp dpt:http to:192.168.99.6:80
8    DNAT       tcp  --  anywhere             anywhere             tcp dpt:http to:192.168.99.7:80
[root@shcCDFrh75vm7 container]# iptables -D PREROUTING -t nat 8
Bad argument `8'
Try `iptables -h' or 'iptables --help' for more information.
[root@shcCDFrh75vm7 container]# iptables -t nat -D PREROUTING 8
[root@shcCDFrh75vm7 container]# iptables -L PREROUTING -t nat --line-number
Chain PREROUTING (policy ACCEPT)
num  target     prot opt source               destination
1    PREROUTING_direct  all  --  anywhere             anywhere
2    PREROUTING_ZONES_SOURCE  all  --  anywhere             anywhere
3    PREROUTING_ZONES  all  --  anywhere             anywhere
4    DNAT       tcp  --  anywhere             anywhere             tcp dpt:http to:192.168.99.6:80
5    DNAT       tcp  --  anywhere             anywhere             tcp dpt:http to:192.168.99.4:80
6    DNAT       tcp  --  anywhere             anywhere             tcp dpt:http to:192.168.99.5:80
7    DNAT       tcp  --  anywhere             anywhere             tcp dpt:http to:192.168.99.6:80
[root@shcCDFrh75vm7 container]# iptables -t nat -D PREROUTING 7
[root@shcCDFrh75vm7 container]# iptables -t nat -D PREROUTING 6
[root@shcCDFrh75vm7 container]# iptables -t nat -D PREROUTING 5
[root@shcCDFrh75vm7 container]# iptables -t nat -D PREROUTING 4
[root@shcCDFrh75vm7 container]# iptables -L PREROUTING -t nat --line-number
Chain PREROUTING (policy ACCEPT)
num  target     prot opt source               destination
1    PREROUTING_direct  all  --  anywhere             anywhere
2    PREROUTING_ZONES_SOURCE  all  --  anywhere             anywhere
3    PREROUTING_ZONES  all  --  anywhere             anywhere
[root@shcCDFrh75vm7 container]# iptables -L  POSTROUTING -t nat --line-number
Chain POSTROUTING (policy ACCEPT)
num  target     prot opt source               destination
1    RETURN     all  --  192.168.122.0/24     base-address.mcast.net/24
2    RETURN     all  --  192.168.122.0/24     255.255.255.255
3    MASQUERADE  tcp  --  192.168.122.0/24    !192.168.122.0/24     masq ports: 1024-65535
4    MASQUERADE  udp  --  192.168.122.0/24    !192.168.122.0/24     masq ports: 1024-65535
5    MASQUERADE  all  --  192.168.122.0/24    !192.168.122.0/24
6    POSTROUTING_direct  all  --  anywhere             anywhere
7    POSTROUTING_ZONES_SOURCE  all  --  anywhere             anywhere
8    POSTROUTING_ZONES  all  --  anywhere             anywhere
9    MASQUERADE  all  --  192.168.99.0/24      anywhere
10   MASQUERADE  all  --  192.168.99.0/24      anywhere
11   MASQUERADE  all  --  192.168.99.0/24      anywhere
12   MASQUERADE  all  --  192.168.99.0/24      anywhere
13   MASQUERADE  all  --  192.168.99.0/24      anywhere
14   MASQUERADE  all  --  192.168.99.0/24      anywhere
15   MASQUERADE  all  --  192.168.99.0/24      anywhere
16   MASQUERADE  all  --  192.168.99.0/24      anywhere
17   MASQUERADE  all  --  192.168.99.0/24      anywhere
[root@shcCDFrh75vm7 container]# iptables -t nat -D POSTROUTING 17
[root@shcCDFrh75vm7 container]# iptables -t nat -D POSTROUTING 16 15
Bad argument `15'
Try `iptables -h' or 'iptables --help' for more information.
[root@shcCDFrh75vm7 container]# iptables -t nat -D POSTROUTING 16
[root@shcCDFrh75vm7 container]# iptables -t nat -D POSTROUTING 15
[root@shcCDFrh75vm7 container]# iptables -t nat -D POSTROUTING 14
[root@shcCDFrh75vm7 container]# iptables -t nat -D POSTROUTING 13
[root@shcCDFrh75vm7 container]# iptables -t nat -D POSTROUTING 12
[root@shcCDFrh75vm7 container]# iptables -t nat -D POSTROUTING 11
[root@shcCDFrh75vm7 container]# iptables -t nat -D POSTROUTING 10
[root@shcCDFrh75vm7 container]# iptables -L  POSTROUTING -t nat --line-number
Chain POSTROUTING (policy ACCEPT)
num  target     prot opt source               destination
1    RETURN     all  --  192.168.122.0/24     base-address.mcast.net/24
2    RETURN     all  --  192.168.122.0/24     255.255.255.255
3    MASQUERADE  tcp  --  192.168.122.0/24    !192.168.122.0/24     masq ports: 1024-65535
4    MASQUERADE  udp  --  192.168.122.0/24    !192.168.122.0/24     masq ports: 1024-65535
5    MASQUERADE  all  --  192.168.122.0/24    !192.168.122.0/24
6    POSTROUTING_direct  all  --  anywhere             anywhere
7    POSTROUTING_ZONES_SOURCE  all  --  anywhere             anywhere
8    POSTROUTING_ZONES  all  --  anywhere             anywhere
9    MASQUERADE  all  --  192.168.99.0/24      anywhere

```

```bash
[root@shcCDFrh75vm7 container]# ./container network create --driver bridge --subnet 192.168.99.0/24 testbridge
2021-09-01T17:17:47.235+0800    DEBUG   ***** [NETWORK-CREATE] PreRun *****
2021-09-01T17:17:47.238+0800    DEBUG   subnet: 192.168.99.1/24
2021-09-01T17:17:47.245+0800    DEBUG   ***** [NETWORK-CREATE] PostRun *****
[root@shcCDFrh75vm7 container]# ./container run -it -m 100m --cpus 1 --network testbridge  busybox /bin/sh
2021-09-01T17:18:14.110+0800    INFO    image name: busybox
2021-09-01T17:18:14.110+0800    INFO    command: [/bin/sh]
2021-09-01T17:18:14.110+0800    DEBUG   ***** RUN Run *****
2021-09-01T17:18:14.163+0800    DEBUG   container id: 7196002797, name: 7196002797
2021-09-01T17:18:14.168+0800    DEBUG   load network: /var/run/q.container/network/network/testbridge
2021-09-01T17:18:14.168+0800    DEBUG   &{testbridge bridge 192.168.99.1/24}
2021-09-01T17:18:14.168+0800    DEBUG   ip: 192.168.99.1
2021-09-01T17:18:14.168+0800    DEBUG   mask: ffffff00
2021-09-01T17:18:14.169+0800    DEBUG   alloc ip: 192.168.99.12
2021-09-01T17:18:14.173+0800    INFO    initializing
2021-09-01T17:18:14.207+0800    DEBUG   ep ip: 192.168.99.12
2021-09-01T17:18:14.207+0800    DEBUG   ep ip range: {192.168.99.1 ffffff00}
2021-09-01T17:18:14.208+0800    INFO    send cmd: /bin/sh
2021-09-01T17:18:14.208+0800    INFO    send cmd: /bin/sh success
2021-09-01T17:18:14.208+0800    INFO    tty true
2021-09-01T17:18:14.208+0800    DEBUG   setting mount
2021-09-01T17:18:14.208+0800    DEBUG   pwd: /root/mnt/7196002797
2021-09-01T17:18:14.225+0800    DEBUG   find cmd path: /bin/sh
2021-09-01T17:18:14.225+0800    DEBUG   syscall.Exec cmd path: /bin/sh
/ # /bin/ip link show
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
62: cif-71960@if63: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1500 qdisc noqueue qlen 1000
    link/ether f2:ba:9c:0a:3c:84 brd ff:ff:ff:ff:ff:ff
/ #

```

打开一个新的 terminal，可以看到 `cif-71960@if63` 在宿主机上 `testbridge` 网桥上的另一端 `71960@if62`：
```bash
[root@shcCDFrh75vm7 container]# ip link show
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN mode DEFAULT group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
2: ens32: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP mode DEFAULT group default qlen 1000
    link/ether 00:50:56:b0:31:14 brd ff:ff:ff:ff:ff:ff
4: virbr0-nic: <BROADCAST,MULTICAST> mtu 1500 qdisc pfifo_fast state DOWN mode DEFAULT group default qlen 1000
    link/ether 52:54:00:b3:a0:49 brd ff:ff:ff:ff:ff:ff
61: testbridge: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP mode DEFAULT group default qlen 1000
    link/ether 4e:52:21:65:45:a1 brd ff:ff:ff:ff:ff:ff
63: 71960@if62: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue master testbridge state UP mode DEFAULT group default qlen 1000
    link/ether 4e:52:21:65:45:a1 brd ff:ff:ff:ff:ff:ff link-netnsid 0

```
