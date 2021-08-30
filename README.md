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
## TODO
- 文件的存放目录和结构（容器信息，image，network, log 等）
- image storage driver
- ipam bytes 实现
- 项目目录重构
- network bridge driver 接口
- cgroups 重构
- 日志 console 和 file 分离
- 解决所有 bug （$PATH，exit throw error, log command error etc.）
- doc 重写（docs，readme）
- https://github.com/sevlyar/go-daemon 实现 fork
- default network bridge
- clean container related resources
