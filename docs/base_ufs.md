# Linux Union File System

## pivot_root

`pivot_root` 是一个系统调用，主要功能是去改变当前的 root 文件系统。`pivot_root` 可以将当前进程的文件系统移动到 `put_old` 文件夹中，然后使用
 `new_root` 成为新的 root 文件系统。`new_root` 和 `put_old` 不能同时存在当前 root 的同一个文件系统中。`pivot_root` 和 `chroot` 的
主要区别是，`pivot_root` 是把整个系统切换到一个新的 root 目录，而移除对之前 root 文件系统的依赖，这样你就能够 umount 原先的 root 文件系统。
而 chroot 是针对某个进程，系统的其他部分依旧运行于老的 root 目录中。

## busybox

获取 busybox 的 rootfs：

```bash
docker pull busybox

# 启动一个容器
docker run -it busybox:<tag> /bin/sh
# 或者
docker run -d busybox top -b


# 导出 rootfs 到文件
docker export - o busybox.tar <container id>

# 解压
tar -xvf busybox . tar -C busybox/

# [root@shcCDFrh75vm7 ~]# ls busybox/
bin  dev  etc  home  lib  lib64  proc  root  sys  tmp  usr  var  version.txt
```


使用 AUFS 创建容器文件系统的实现过程如下。
启动容器的时候：

1. 建只读层（busybox）
2. 创建容器读写层（writeLayer）
3. 创建挂载点（mnt），井把只读层和读写层挂载到挂载点
4. 将挂载点作为容器的根目录


容器退出：
1. umount 挂载点
2. 删除挂载点
3. 删除读写层（writeLayer）

实现 volume，创建容器文件系统的过程如下：

1. 建只读层
2. 创建容器读写层
3. 创建挂载点，井把只读层和读写层挂载到挂载点
4. 接下来首先判断 volume 是否为空，如果是，就表示用户并没有使用 volume 结束创建 volume。
5. 如果不为空，则解析 volume 字符串。
6. 来挂载数据卷。

挂载数据卷的过程：
1. 读取宿主机文件目录 Url，创建宿主机文件目录（/root/<parentUrl>）
2. 读取容器挂载点目录 Url，在容器文件系统里创建挂载点（/root/mnt/<containerUrl>）
3. 把宿主机文件目录挂载到容器挂载点