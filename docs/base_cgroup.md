# Linux Cgroups

Linux Cgroups (Control Groups ）提供了对一组进程及将来子进程的资源限制、控制和统计的能力，这些资源包括 CPU、内存、存储、网络等。通过 Cgroups，
可以方便地限制某个进程的资源占用，并且可以实时地监控进程的监控和统计信息。

Cgroups 的 3 个组件
- cgroup 是对进程分组管理的一种机制，一个 cgroup 包含一组进程，井可以在这个 cgroup 上增加 Linux subsystem 的各种参数配置，将一组进程和 
  一组 subsystem 的系统参数关联起来。在用户层看来，cgroup 技术就是把系统中的所有进程组织成一颗一颗独立的树，每棵树都包含系统的所有进程，树的每个
  节点是一个进程组，而每颗树又和一个或者多个 subsystem 关联，树的作用是将进程分组，而 subsystem 的作用就是对这些组进行操作。
- subsystem 是一组资源控制的模块，一个 subsystem 就是一个内核模块，他被关联到一颗 cgroup 树之后，就会在树的每个节点（进程组）上做具体的操作。
  subsystem 经常被称作 resource controller。一般包含如下几项。
  - blkio 设置对块设备（比如硬盘）输入输出的访问控制。
  - cpu 设置cgroup 中进程的CPU 被调度的策略。
  - cpuacct 可以统计cgroup 中进程的CPU 占用。
  - cpuset 在多核机器上设置 cgroup 中进程可以使用的CPU 和内存（此处内存仅使用于 NUMA 架构） 。
  - devices 控制 cgroup 中进程对设备的访问。
  - freezer 用于挂起（suspend ）和恢复（resume) cgroup 中的进程。
  - memory 用于控制 cgroup 中进程的内存占用。
  - net_cls 用于将 cgroup 中进程产生的网络包分类，以便 Linux 的 tc (traffic controller）可以根据分类区分出来自某个 cgroup 的包并做限流或监控。
  - net_prio 设置 cgroup 中进程产生的网络流量的优先级。
  - ns 这个 subsystem 比较特殊，它的作用是使 cgroup 中的进程在新的 Namespace 中 fork 新进程（NEWNS）时，创建出一个新的 cgroup ，这个 
  cgroup 包含新的 Namespace 中的进程。
- hierarchy 的功能是把一组 cgroup 串成一个树状的结构，一个这样的树便是一个 hierarchy，通过这种树状结构，Cgroups 可以做到继承。比如，系统对一组
  定时的任务进程通过 cgroupl 限制了 CPU 的使用率，然后其中有一个定时 dump 日志的进程还需要限制磁盘 IO ，为了避免限制了磁盘 IO 之后影响到其他进程，
  就可以创建 cgroup2 ，使其继承于cgroupl 井限制磁盘的 IO ，这样 cgroup2 便继承了cgroupl 中对 CPU 使用率限制，并且增加了磁盘 IO 的限制而不
  影响到 cgroupl 中的其他进程。

三个组件相互的关系
Kernel 接口

https://fuckcloudnative.io/posts/understanding-cgroups-part-1-basics/
https://tech.meituan.com/2015/03/31/cgroups.htmlsel 

## /proc/<pid>/mountinfo

如何找到挂载了 subsystem 的 hierarchy 的挂载目录。

```bash
[root@SGDLITVM0905 self]# cat /proc/self/mountinfo
17 39 0:17 / /sys rw,nosuid,nodev,noexec,relatime shared:6 - sysfs sysfs rw,seclabel
18 39 0:3 / /proc rw,nosuid,nodev,noexec,relatime shared:5 - proc proc rw
19 39 0:5 / /dev rw,nosuid shared:2 - devtmpfs devtmpfs rw,seclabel,size=8119352k,nr_inodes=2029838,mode=755
20 17 0:16 / /sys/kernel/security rw,nosuid,nodev,noexec,relatime shared:7 - securityfs securityfs rw
21 19 0:18 / /dev/shm rw,nosuid,nodev shared:3 - tmpfs tmpfs rw,seclabel
22 19 0:12 / /dev/pts rw,nosuid,noexec,relatime shared:4 - devpts devpts rw,seclabel,gid=5,mode=620,ptmxmode=000
23 39 0:19 / /run rw,nosuid,nodev shared:23 - tmpfs tmpfs rw,seclabel,mode=755
24 17 0:20 / /sys/fs/cgroup ro,nosuid,nodev,noexec shared:8 - tmpfs tmpfs ro,seclabel,mode=755
25 24 0:21 / /sys/fs/cgroup/systemd rw,nosuid,nodev,noexec,relatime shared:9 - cgroup cgroup rw,seclabel,xattr,release_agent=/usr/lib/systemd/systemd-cgroups-agent,name=systemd
26 17 0:22 / /sys/fs/pstore rw,nosuid,nodev,noexec,relatime shared:20 - pstore pstore rw
27 24 0:23 / /sys/fs/cgroup/net_cls,net_prio rw,nosuid,nodev,noexec,relatime shared:10 - cgroup cgroup rw,seclabel,net_prio,net_cls
28 24 0:24 / /sys/fs/cgroup/blkio rw,nosuid,nodev,noexec,relatime shared:11 - cgroup cgroup rw,seclabel,blkio
29 24 0:25 / /sys/fs/cgroup/cpu,cpuacct rw,nosuid,nodev,noexec,relatime shared:12 - cgroup cgroup rw,seclabel,cpuacct,cpu
30 24 0:26 / /sys/fs/cgroup/memory rw,nosuid,nodev,noexec,relatime shared:13 - cgroup cgroup rw,seclabel,memory
31 24 0:27 / /sys/fs/cgroup/freezer rw,nosuid,nodev,noexec,relatime shared:14 - cgroup cgroup rw,seclabel,freezer
32 24 0:28 / /sys/fs/cgroup/cpuset rw,nosuid,nodev,noexec,relatime shared:15 - cgroup cgroup rw,seclabel,cpuset
33 24 0:29 / /sys/fs/cgroup/perf_event rw,nosuid,nodev,noexec,relatime shared:16 - cgroup cgroup rw,seclabel,perf_event
34 24 0:30 / /sys/fs/cgroup/hugetlb rw,nosuid,nodev,noexec,relatime shared:17 - cgroup cgroup rw,seclabel,hugetlb
35 24 0:31 / /sys/fs/cgroup/pids rw,nosuid,nodev,noexec,relatime shared:18 - cgroup cgroup rw,seclabel,pids
36 24 0:32 / /sys/fs/cgroup/devices rw,nosuid,nodev,noexec,relatime shared:19 - cgroup cgroup rw,seclabel,devices
37 17 0:33 / /sys/kernel/config rw,relatime shared:21 - configfs configfs rw
39 0 253:0 / / rw,relatime shared:1 - xfs /dev/mapper/centos7-root rw,seclabel,attr2,inode64,noquota
40 17 0:15 / /sys/fs/selinux rw,relatime shared:22 - selinuxfs selinuxfs rw
...
```

通过 `/proc/self/mountinfo`，可以找出与当前进程相关的 mount 信息。
