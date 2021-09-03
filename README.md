# container

## Usage

启动容器：
```bash
./container run -it busybox /bin/sh

# 添加环境变量
./container run -it -e testhome=/opt/test busybox /bin/sh

# 挂载 volume
./container run -it -v /root/testvolume=/testvolume busybox /bin/sh

# 限制资源
./container run -it -m 100m --cpus 1 busybox /bin/sh

# expose port
./container run -it -m 100m --cpus 1 -p 8080:80 busybox /bin/sh

# 后台运行
./container run -d -m 100m --cpus 1 -p 8080:80 busybox /bin/sh
```

进入容器：
```bash
./container exec -it <container id> /bin/bash 
```

停止容器：
```bash
./container stop <container id>
```

删除容器：
```bash
./container rm <container id>
```

查看容器列表：
```bash
./container ps
```

提交容器：
```bash
./container commit <container id> <image name>
```

创建网络：
```bash
./container network create --driver bridge --subnet 192.168.99.0/24 testbridge
```

查看网络列表：
```bash
./container network ls
```

删除网络：
```bash
./container network rm testbridge
```

## Release
- 具体功能的实现可以参考 [release 的各个版本](https://github.com/shipengqi/container/releases?after=v1.5)

## TODO
- 结构重构（容器信息，image，network, log 等）
- image storage driver
- [文档](./docs)，重要模块的实现原理
  - namespace 隔离
  - cgroups,
  - aufs/overlay
  - `-e` 参数
  - `--volume` 参数
  - `logs` 命令, 
  - `exec` 命令 
  - `exec -e` 参数
  - network
  - port mapping
- 容器退出或者强制删除后，清楚相关资源（workspace，mount point，iptables rules，container info，etc.）
- 创建 iptables chain，iptables rules 添加包含 containerid 的注释
- 构建 hook 函数，用来准备容器运行资源（创建目录）
- recover panic
- log pipe cannot close
- 进程间 error 的传递
- 删除创建文件之前 check 文件是否存在
- 所有容器资源目录重命名

## Know issues
- logs command error
- cannot get $PATH
- container status sync error (stop, rm -f cmd)
- exit throw error
  ```bash
    2021-08-31T12:16:39.147+0800    ERROR   parent wait     {"error": "exit status 130"}
    2021-08-31T12:16:39.151+0800    WARN    remove cgroup fail unlinkat /sys/fs/cgroup/cpuset/q.container.cgroup/cpuset.memory_spread_slab: operation not permitted
    2021-08-31T12:16:39.152+0800    WARN    remove cgroup fail unlinkat /sys/fs/cgroup/memory/q.container.cgroup/memory.kmem.tcp.max_usage_in_bytes: operation not permitted
    2021-08-31T12:16:39.152+0800    WARN    remove cgroup fail unlinkat /sys/fs/cgroup/cpu,cpuacct/q.container.cgroup/cpu.rt_period_us: operation not permitted
    2021-08-31T12:16:39.152+0800    ERROR   container.Execute(): exit status 130
    
    
    2021-08-31T12:19:29.694+0800    ERROR   parent wait     {"error": "exit status 130"}
    2021-08-31T12:19:29.703+0800    ERROR   container.Execute(): exit status 130
  ```

## 参考
- <<自己动手写 Docker>>
- https://github.com/xianlubird/mydocker
- https://github.com/xianlubird/cka-pre
- https://cloud.tencent.com/developer/section/1145002 text/tabwriter
- https://github.com/vishvananda/netlink 操作网络接口、路由表等配置的库
- https://github.com/vishvananda/netns Go 语言版 netns
- https://github.com/huataihuang/cloud-atlas
- https://cloud-atlas.readthedocs.io/zh_CN/latest/docker/index.html
- https://xie.infoq.cn/article/11d413217d5186feed013122e
- https://github.com/sevlyar/go-daemon
- https://blog.csdn.net/kikajack/article/details/80457841
- https://www.cnblogs.com/liyuanhong/p/13585654.html
- https://github.com/opencontainers/runc
- https://github.com/coreos/go-iptables
- https://github.com/containernetworking/plugins
- https://www.frozentux.net/iptables-tutorial/cn/iptables-tutorial-cn-1.1.19.html