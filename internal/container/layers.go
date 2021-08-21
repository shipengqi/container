package container

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/shipengqi/container/pkg/log"
	"github.com/shipengqi/container/pkg/utils"
)

// 1.2 中实现了使用宿主机 /root/busybox 录作为文件的根目录，但在容器内对文
// 件的操作仍然会直接影响到宿主机的 /root/busybox 录。需要实现容器和镜像隔离，
// 容器中进行的操作不会对镜像产生任何影响的功能

// NewWorkSpace Create a AUFS filesystem as container root workspace
func NewWorkSpace(rootUrl string, mntUrl string) error {
	err := CreateReadOnlyLayer(rootUrl)
	if err != nil {
		return err
	}
	err = CreateWriteLayer(rootUrl)
	if err != nil {
		return err
	}
	err = CreateMountPoint(rootUrl, mntUrl)
	if err != nil {
		return err
	}
	return nil
}

// CreateReadOnlyLayer 将 busybox.tar 解压到 busybox 目录下，作为容器的只读层
func CreateReadOnlyLayer(rootUrl string) error {
	busyboxUrl := rootUrl + "busybox/"
	busyboxTarUrl := rootUrl + "busybox.tar"
	if utils.IsNotExist(busyboxUrl) {
		if err := os.Mkdir(busyboxUrl, 0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", busyboxUrl, err)
			return err
		}
	}
	if utils.IsNotExist(busyboxTarUrl) {
		return errors.Errorf("%s not found", busyboxTarUrl)
	}
	if _, err := exec.Command("tar", "-xvf", busyboxTarUrl, "-C", busyboxUrl).CombinedOutput(); err != nil {
		log.Errorf("uncompress dir %s error %v", busyboxUrl, err)
		return err
	}

	return nil
}

// CreateWriteLayer 创建一个 writeLayer 的文件夹作为容器唯一的可写层
func CreateWriteLayer(rootUrl string) error {
	writeUrl := rootUrl + "writeLayer/"
	if utils.IsNotExist(writeUrl) {
		if err := os.Mkdir(writeUrl, 0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", writeUrl, err)
			return err
		}
	}
	return nil
}

func CreateMountPoint(rootUrl string, mntUrl string) error {
	workUrl := "/root/q.container.work"
	// 创建 mnt 文件夹，并作为挂载点
	if utils.IsNotExist(mntUrl) {
		if err := os.Mkdir(mntUrl, 0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", mntUrl, err)
			return err
		}
	}
	// 创建 work 文件夹
	if utils.IsNotExist(workUrl) {
		if err := os.Mkdir(workUrl, 0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", workUrl, err)
			return err
		}
	}

	// 把 writerLayer 和 busybox 目录 mount 到 mnt 目录下
	// dirs := "dirs=" + rootUrl + "writeLayer:" + rootUrl + "busybox"
	// https://unix.stackexchange.com/questions/477029/how-to-mount-aufs-filesystem
	// mount: unknown filesystem type 'aufs'
	// 使用 overlay 文件系统
	// mount overlay on /root/testoverlay/merge failed: Operation not supported
	// mount -t overlay overlay -o lowerdir=low1,upperdir=up1,workdir=work merge
	// mount -t overlay overlay -o lowerdir=/root/busybox,upperdir=/root/writeLayer,workdir=/root/q.container.work /root/mnt
	// 必须有 workdir
	dirs := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s",
		filepath.Join(rootUrl, "busybox"),
		filepath.Join(rootUrl, "writeLayer"),
		workUrl)
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("mount mnt: %v", err)
		return err
	}
	return nil
}

// https://askubuntu.com/questions/109413/how-do-i-use-overlayfs
// https://www.wumingx.com/k8s/docker-rootfs.html
// https://blog.csdn.net/luckyapple1028/article/details/78075358

// Overlay test
// [root@shcCDFrh75vm7 ~]# mkdir testoverlay
// [root@shcCDFrh75vm7 ~]# cd testoverlay/
// [root@shcCDFrh75vm7 testoverlay]# ls
// [root@shcCDFrh75vm7 testoverlay]# mkdir low1 low2 up1 merge work
// [root@shcCDFrh75vm7 testoverlay]#
// [root@shcCDFrh75vm7 testoverlay]#
// [root@shcCDFrh75vm7 testoverlay]# ls
// low1  low2  merge  up1  work
// [root@shcCDFrh75vm7 testoverlay]# touch ./low1/low1f
// [root@shcCDFrh75vm7 testoverlay]# touch ./low2/low2f
// [root@shcCDFrh75vm7 testoverlay]# touch ./up1/up1f
// [root@shcCDFrh75vm7 testoverlay]# mount -t overlay overlay -o lowerdir=low1,upperdir=up1 merge
// mount: mount overlay on /root/testoverlay/merge failed: Operation not supported
// [root@shcCDFrh75vm7 testoverlay]# mount -t overlay overlay -o lowerdir=low1,upperdir=up1,workdir=work merge
// [root@shcCDFrh75vm7 testoverlay]# cd work/
// [root@shcCDFrh75vm7 work]# ls
// work
// [root@shcCDFrh75vm7 work]# cd work/
// [root@shcCDFrh75vm7 work]# ls
// [root@shcCDFrh75vm7 work]# cd ..
// [root@shcCDFrh75vm7 work]# cd ..
// [root@shcCDFrh75vm7 testoverlay]# ls
// low1  low2  merge  up1  work
// [root@shcCDFrh75vm7 testoverlay]# cd merge/
// [root@shcCDFrh75vm7 merge]# ls
// low1f  up1f

// DeleteWorkSpace Delete the AUFS filesystem while container exit
func DeleteWorkSpace(rootUrl string, mntUrl string){
	DeleteMountPoint(rootUrl, mntUrl)
	DeleteWriteLayer(rootUrl)
}

func DeleteMountPoint(rootUrl string, mntUrl string){
	cmd := exec.Command("umount", mntUrl)
	cmd.Stdout=os.Stdout
	cmd.Stderr=os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v",err)
	}
	// if err := os.RemoveAll(mntUrl); err != nil {
	// 	log.Errorf("Remove dir %s error %v", mntUrl, err)
	// }
}

func DeleteWriteLayer(rootUrl string) {
	writeUrl := rootUrl + "writeLayer/"
	if err := os.RemoveAll(writeUrl); err != nil {
		log.Errorf("Remove dir %s error %v", writeUrl, err)
	}
}