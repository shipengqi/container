# Linux Cgroups

Linux Cgroups (Control Groups ）提供了对一组进程及将来子进程的资源限制、控制和统计的能力，这些资源包括 CPU、内存、存储、网络等。通过 Cgroups，
可以方便地限制某个进程的资源占用，并且可以实时地监控进程的监控和统计信息。

Cgroups 的 3 个组件
- cgroup 是对进程分组管理的一种机制，一个 cgroup 包含一组进程，井可以在这个 cgroup 上增加 Linux subsystem 的各种参数配置，将一组进程和 
  一组 subsystem 的系统参数关联起来。
- subsystem 是一组资源控制的模块，一般包含如下几项。
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
