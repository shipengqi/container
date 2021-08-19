# Linux proc file system

Linux 下的 `/proc` 文件系统是由内核提供的，并不是一个真正的文件系统，只包含系统的运行时信息（系统内存，mount 设备信息，硬件配置信息等）。
**只存在于内存中**。以文件系统的形式，为访问内核数据的操作提供接口。实际上很多系统工具也是简单的读取这个文件系统的文件内容。（`lsmod`，就是 `cat /proc/modules`）。

这个目录下的很多数字，都是为进程创建的，数字就是 PID。几个比较重要的部分：

- `/proc/<PID>/cmdline`  进程的启动命令
- `/proc/<PID>/cwd`      链接到进程当前工作目录
- `/proc/<PID>/environ`  进程环境变量列表
- `/proc/<PID>/exe`      链接到进程的执行命令文件
- `/proc/<PID>/fd`       包含进程相关的所有文件描述符
- `/proc/<PID>/maps`     与进程相关的内存映射信息
- `/proc/<PID>/mem`      指代进程持有的内存，不可读
- `/proc/<PID>/root`     链接到进程的根目录
- `/proc/<PID>/stat`     进程的状态
- `/proc/<PID>/statm`    进程使用的内存状态
- `/proc/<PID>/status`   进程状态信息，比 `stat/statm` 更具可读性
- `/proc/self/`          链接到当前正在运行的进程，和 `/proc/<PID>` 下的结构是一样的

示例：
```bash
[root@shcCDFrh75vm7 ~]# ll /proc/2725
total 0
-r--r--r--.  1 mysql mysql 0 Aug 18 15:20 cmdline
lrwxrwxrwx.  1 mysql mysql 0 Aug 19 11:10 cwd -> /var/lib/mysql
-r--------.  1 mysql mysql 0 Aug 19 11:10 environ
lrwxrwxrwx.  1 mysql mysql 0 Aug 18 15:36 exe -> /usr/sbin/mysqld
dr-x------.  2 mysql mysql 0 Aug 18 15:19 fd
-r--r--r--.  1 mysql mysql 0 Aug 19 11:10 maps
-rw-------.  1 mysql mysql 0 Aug 19 11:10 mem
lrwxrwxrwx.  1 mysql mysql 0 Aug 19 11:10 root -> /
-r--r--r--.  1 mysql mysql 0 Aug 18 15:19 stat
-r--r--r--.  1 mysql mysql 0 Aug 18 17:34 statm
-r--r--r--.  1 mysql mysql 0 Aug 18 15:19 status
```

如果某个进程想要获取本进程的系统信息，就可以通过进程的 pid 来访问 `/proc/<pid>/` 目录。但是这个方法还需要获取进程 pid，在 fork、daemon 
等情况下 pid 还可能发生变化。为了更方便的获取本进程的信息，linux 提供了 `/proc/self/` 目录，这个目录比较独特，不同的进程访问该目录时获得的
信息是不同的，内容等价于 `/proc/<pid>`。进程可以通过访问 `/proc/self/` 目录来获取自己的系统信息，而不用每次都获取 pid。









