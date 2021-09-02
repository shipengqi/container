# container


## 参考
- 自己动手写 Docker
- https://github.com/xianlubird/cka-pre
- https://github.com/xianlubird/mydocker
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

## TODO
- 文件的存放目录和结构（容器信息，image，network, log 等）
- image storage driver
- 项目目录重构
- cgroups 重构
- doc（几个重要功能的实现 namespace 隔离，cgroups, aufs, `-e` `--volume` `logs` cmd, `exec` cmd `exec -e`，network，portmapping）
- clean container related resources, iptables rules (route tables will auto clean)
- path.Join and fmt.Sprintf or string + string
- iptables pkg create chain and add comments with containerid
- mkdir all dirs in hook function
- recover panic
- log pipe cannot close
- 进程间 error 的传递
- 删除创建文件之前 check 文件是否存在

## Know issues
- log command error
- cannot get $PATH
- exit throw error
- container status sync error (stop, rm -f cmd)

```bash
2021-08-31T12:16:39.147+0800    ERROR   parent wait     {"error": "exit status 130"}
2021-08-31T12:16:39.151+0800    WARN    remove cgroup fail unlinkat /sys/fs/cgroup/cpuset/q.container.cgroup/cpuset.memory_spread_slab: operation not permitted
2021-08-31T12:16:39.152+0800    WARN    remove cgroup fail unlinkat /sys/fs/cgroup/memory/q.container.cgroup/memory.kmem.tcp.max_usage_in_bytes: operation not permitted
2021-08-31T12:16:39.152+0800    WARN    remove cgroup fail unlinkat /sys/fs/cgroup/cpu,cpuacct/q.container.cgroup/cpu.rt_period_us: operation not permitted
2021-08-31T12:16:39.152+0800    ERROR   container.Execute(): exit status 130


2021-08-31T12:19:29.694+0800    ERROR   parent wait     {"error": "exit status 130"}
2021-08-31T12:19:29.703+0800    ERROR   container.Execute(): exit status 130
```

