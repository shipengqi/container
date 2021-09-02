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


```go
func pivotRoot(root string) error {
    /*
      由于老 root 和新的 root (/root/mnt/<containerID>) 当前都是在一个文件系统下，所以这里我们把 root 重新 mount 了一次，这样老的 root 和新的 root 就不再同一个文件系统下
      bind mount 是把相同的内容换了一个挂载点的挂载方法
    */
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return errors.Wrap(err, "mount rootfs to itself")
	}
	// 创建 rootfs/.pivot_root 用来存储 old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	if utils.IsNotExist(pivotDir) {
		if err := os.Mkdir(pivotDir, 0777); err != nil {
			return err
		}
	}

    // https://github.com/xianlubird/mydocker/issues/62
    // Fix pivot_root: invalid argument
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return errors.Wrap(err, "mount /")
	}

	// pivot_root 使用新的 rootfs, 将老的 old_root 是挂载在 rootfs/.pivot_root
	// 挂载点现在依然可以在 mount 命令中看到
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return errors.Wrap(err, "syscall.PivotRoot")
	}
	// 修改当前的工作目录到根目录
	if err := syscall.Chdir("/"); err != nil {
		return errors.Wrap(err, "syscall.Chdir")
	}
	pivotDir = filepath.Join("/", ".pivot_root")
	// umount rootfs/.pivot_root
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return errors.Wrap(err, "unmount pivot_root")
	}
	// 删除临时文件夹
	return os.Remove(pivotDir)
}
```

执行 mount 的时候，因为已经在 mount namespace 中，所以 mount 和 umount 仅仅只会影响当前 Namespace 内的文件系统
```go
func setUpMount() error {
	// 获取当前路径
	pwd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "os.Getwd")
	}
	log.Debugf("pwd: %s", pwd)
	err = pivotRoot(pwd)
	if err != nil {
		return err
	}

	// mount proc
	mflags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	err = syscall.Mount("proc", "/proc", "proc", uintptr(mflags), "")
	if err != nil {
		return errors.Wrap(err, "mount proc")
	}

	// tmpfs 是一种基于内存的文件系统，可以使用 RAM 或 swap 分区来存储。
	err = syscall.Mount("tmpfs", "/dev", "tmpfs",
		syscall.MS_NOSUID | syscall.MS_STRICTATIME, "mode=755")
	if err != nil {
		return errors.Wrap(err, "mount tmpfs")
	}

	return nil
}
```

### 创建文件系统 
NewWorkSpace 函数是用来创建容器文件系统的
CreateReadOnlyLayer 函数新建 busybox 文件夹，将 `busybox.tar` 解压到 busybox 目录下， 作为容器的只读层。
CreateWriteLayer 函数创建了一个名为 writeLayer 的文件夹，作为容器唯一的可写层。 
CreateMountPoint 函数，首先创建了 mnt 文件夹，作为挂载点，然后把 writeLayer 目录和 busybox 目录 mount 到 mnt 目录下。


```go
func NewWorkSpace(volume, imgName, containerId string) (string, error) {
	err := CreateReadOnlyLayer(imgName)
	if err != nil {
		return "", err
	}
	err = CreateWriteLayer(containerId)
	if err != nil {
		return "", err
	}

	mntUrl, err := CreateMountPoint(imgName, containerId)
	if err != nil {
		return "", err
	}

	// if volume length > 0，创建 volume
	if len(volume) > 0 {
		return mntUrl, CreateVolume(containerId, volume)
	}
	return mntUrl, nil
}
```

将容器使用的宿主机目录改为 `mntUrl`。这样，启动容器就会使用 overlay 文件系统。


```go
	mntUrl, err := NewWorkSpace(volume, imgName, containerId)
	if err != nil {
		return nil, nil, errors.Wrap(err, "new workspace")
	}
	cmd.Dir = mntUrl
```

### volume
`run -v /root/volume:/containerVolume`

```go
func MountVolume(mntUrl string, volumeUrls []string) error {
	parentUrl := volumeUrls[0]
	parentROUrl := volumeUrls[0] + ".ro"
	if utils.IsNotExist(parentUrl) {
		if err := os.Mkdir(parentUrl, 0777); err != nil {
			log.Infof("Mkdir parent dir %s error. %v", parentUrl, err)
		}
	}
	if utils.IsNotExist(parentROUrl) {
		if err := os.Mkdir(parentROUrl, 0777); err != nil {
			log.Infof("Mkdir parent ro dir %s error. %v", parentROUrl, err)
		}
	}

	volumeUrl := volumeUrls[1]
	containerVolumeUrl := filepath.Join(mntUrl, volumeUrl)
	log.Debugf("volume container url: %s", containerVolumeUrl)
	if utils.IsNotExist(containerVolumeUrl) {
		if err := os.Mkdir(containerVolumeUrl, 0777); err != nil {
			log.Infof("Mkdir container dir %s error. %v", containerVolumeUrl, err)
		}
	}
	// overlay 必须有一个 lowerdir 所以这里创建了一个 parentROUrl
	dirs := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", parentROUrl, parentUrl, TmpWorkUrl)
	log.Debugf("volume dirs: %s", dirs)
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, containerVolumeUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "mount volume")
	}
	return nil
}
```