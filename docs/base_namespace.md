# Linux namespace

Linux Namespace 是 Kernel 的一个功能，它可以隔离一系列的系统资源，PIO、User ID 、Network、挂载点等。

Linux 一共实现了 6 种不同类型的 Namespace：
```bash
[root@SGDLITVM0905 ~]# ls -l /proc/self/ns
total 0
lrwxrwxrwx. 1 root root 0 Aug 17 09:48 ipc -> ipc:[4026531839]
lrwxrwxrwx. 1 root root 0 Aug 17 09:48 mnt -> mnt:[4026531840]
lrwxrwxrwx. 1 root root 0 Aug 17 09:48 net -> net:[4026531956]
lrwxrwxrwx. 1 root root 0 Aug 17 09:48 pid -> pid:[4026531836]
lrwxrwxrwx. 1 root root 0 Aug 17 09:48 user -> user:[4026531837]
lrwxrwxrwx. 1 root root 0 Aug 17 09:48 uts -> uts:[4026531838]
```

| Namespace 类型| 系统调用参数 | 内核版本 |
| ---- | ---- | ---- |
| Mount Namespace | CLONE_NEWNS | 2.4.19 |
| UTS Namespace | CLONE_NEWUTS | 2.6.19 |
| IPC Namespace | CLONE_NEWIPC | 2.6.19 |
| PID Namespace | CLONE_NEWPID | 2.6.24 |
| Network Namespace | CLONE_NEWNET | 2.6.29 |
| User Namespace | CLONE_NEWUSER | 3.8 |

Namespace 的 API 主要使用如下3 个系统调用。
- `clone` 创建新进程。根据系统调用参数来判断哪些类型的 Namespace 被创建，而且它们 的子进程也会被包含到这些 Namespace 中。
- `unshare` 将进程移出某个 Namespace
- `setns` 将进程加入到 Namespace 中。

UTS Namespace 主要用来隔离 nodename 和 domainname 两个系统标识。在 UTS Namespace 里面， 每个 Namespace 允许有自己的 hostname 。
IPC  Namespace 用来隔离 System V IPC 和 POSIX message queues。
PID Namespace 是用来隔离进程 ID 的。
Mount Namespace 用来隔离各个进程看到的挂载点视图。在不同 Namespace 的进程中，看到的文件系统层次是不一样的。在 Mount Namespace 中
调用 mount 和 umount 仅仅只会影响当前 Namespace 内的文件系统，而对全局的文件系统是没有影响的。chroot 也是将某一个子目录变成根节点。
但是， Mount Namespace 不仅能实现这个功能，而且能以更加灵活和安全的方式实现。Mount Namespace 是 Linux 第一个实现的 Namespace 类型，
因此，它的系统调用参数是 `NEWNS` (New Namespace 的缩写）。当时人们貌似没有意识到，以后还会有很多类型的 Namespace 加入。
User Namespace 主要是隔离用户的用户组 ID。
Network Namespace 是用来隔离网络设备、IP 地址端口等网络栈的 Namespace。



## setns

setns 是一个系统调用，可以根据提供的 PID 再次进入到指定的 Namespace 中。它需要先打开 `/proc/[pid]/ns/` 文件夹下对应的文件，然后使当前进程
进入到指定的 Namespace 中。

对于 Go 来说很，一个具有多线程的进程是无法使用 setns 调用进入到对应的命名空间的。但是，Go 每启动一个程序就会进入多线程状态，因此无法简单地
在 Go 里面直接调用系统调用，使当前的进程进入对应的 Mount Namespace 。这里需要借助 C 来实现这个功能。

### Cgo
Cgo 允许 Go 程序去调用C 的函数与标准库。只需要以一种特殊的方式在 Go 的源代码里写出需要调用的 C 的代码，Cgo 就可以把你的 C 源码文件和 Go 文件
整合成一个包。