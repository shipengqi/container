package container

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/shipengqi/container/pkg/log"
	"github.com/shipengqi/container/pkg/utils"
)

// 1.2 中实现了使用宿主机 /root/busybox 录作为文件的根目录，但在容器内对文
// 件的操作仍然会直接影响到宿主机的 /root/busybox 录。需要实现容器和镜像隔离，
// 容器中进行的操作不会对镜像产生任何影响的功能

// NewWorkSpace Create a AUFS filesystem as container root workspace
func NewWorkSpace(rootUrl, mntUrl, volume string) error {
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

	// if volume length > 0，创建 volume
	if len(volume) > 0 {
		return CreateVolume(mntUrl, volume)
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

// DeleteWorkSpace Delete the AUFS filesystem while container exit
func DeleteWorkSpace(rootUrl, mntUrl, volume string) {
	if len(volume) > 0 {
		volumeUrls := volumeUrlExtract(volume)
		if len(volumeUrls) == 2 && len(volumeUrls[0]) > 0 && len(volumeUrls[1]) > 0{
			DeleteMountPointWithVolume(mntUrl, volumeUrls)
		}
	} else {
		DeleteMountPoint(mntUrl)
	}

	DeleteWriteLayer(rootUrl)
}

func DeleteMountPoint(mntUrl string) {
	cmd := exec.Command("umount", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
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

func DeleteMountPointWithVolume(mntUrl string, volumeUrls []string){
	containerUrl := mntUrl + volumeUrls[1]
	cmd := exec.Command("umount", containerUrl)
	cmd.Stdout=os.Stdout
	cmd.Stderr=os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Umount volume failed. %v",err)
	}

	cmd = exec.Command("umount", mntUrl)
	cmd.Stdout=os.Stdout
	cmd.Stderr=os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Umount mountpoint failed. %v",err)
	}

	if err := os.RemoveAll(mntUrl); err != nil {
		log.Infof("Remove mountpoint dir %s error %v", mntUrl, err)
	}
}

func CreateVolume(mntUrl, volume string) error {
	volumeUrls := volumeUrlExtract(volume)
	if len(volumeUrls) < 2 {
		return errors.New("volume parameter is not correct")
	}
	if len(volumeUrls[0]) == 0 || len(volumeUrls[1]) == 0 {
		return errors.New("volume parameter cannot not be empty")
	}
	return MountVolume(mntUrl, volumeUrls)
}

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

	containerUrl := volumeUrls[1]
	containerVolumeUrl := mntUrl + containerUrl
	log.Debugf("volume container url: %s", containerVolumeUrl)
	if utils.IsNotExist(containerVolumeUrl) {
		if err := os.Mkdir(containerVolumeUrl, 0777); err != nil {
			log.Infof("Mkdir container dir %s error. %v", containerVolumeUrl, err)
		}
	}
	workUrl := "/root/q.container.work"
	dirs := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", parentROUrl, parentUrl, workUrl)
	log.Debugf("volume dirs: %s", dirs)
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, containerVolumeUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("mount volume: %v", err)
		return err
	}
	return nil
}

func volumeUrlExtract(volume string) []string {
	var volumeUrls []string
	volumeUrls = strings.Split(volume, ":")
	return volumeUrls
}
