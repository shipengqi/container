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

// NewWorkSpace Create an Overlay filesystem as container root workspace
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

// CreateReadOnlyLayer 将 busybox.tar 解压到 busybox 目录下，作为容器的只读层
func CreateReadOnlyLayer(imgName string) error {
	uncompressUrl := filepath.Join(RootUrl, imgName)
	imageTarUrl := filepath.Join(RootUrl, fmt.Sprintf("%s.tar", imgName))
	if utils.IsNotExist(uncompressUrl) {
		if err := os.Mkdir(uncompressUrl, 0777); err != nil {
			return errors.Errorf("mkdir dir: %s, %v", uncompressUrl, err)
		}
	}
	if utils.IsNotExist(imageTarUrl) {
		return errors.Errorf("%s not found", imageTarUrl)
	}
	if _, err := exec.Command("tar", "-xvf", imageTarUrl, "-C", uncompressUrl).CombinedOutput(); err != nil {
		return errors.Errorf("uncompress dir: %s, %v", uncompressUrl, err)
	}

	return nil
}

// CreateWriteLayer 创建一个 writeLayer 的文件夹作为容器唯一的可写层
func CreateWriteLayer(containerId string) error {
	writeUrl := fmt.Sprintf(WriteLayerUrl, containerId)
	if utils.IsNotExist(writeUrl) {
		if err := os.Mkdir(writeUrl, 0777); err != nil {
			return errors.Errorf("mkdir dir: %s, %v", writeUrl, err)
		}
	}
	return nil
}

func CreateMountPoint(imgName, containerId string) (string, error) {
	// 创建 mnt/<containerId> 文件夹，并作为挂载点
	mntUrl := fmt.Sprintf(MntUrl, containerId)
	writeUrl := fmt.Sprintf(WriteLayerUrl, containerId)
	readUrl := filepath.Join(RootUrl, imgName)
	if utils.IsNotExist(mntUrl) {
		if err := os.Mkdir(mntUrl, 0777); err != nil {
			return "", errors.Errorf("mkdir dir: %s, %v", mntUrl, err)
		}
	}
	// 创建 work 文件夹
	if utils.IsNotExist(TmpWorkUrl) {
		if err := os.Mkdir(TmpWorkUrl, 0777); err != nil {
			return "", errors.Errorf("mkdir dir: %s, %v", TmpWorkUrl, err)
		}
	}

	dirs := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", readUrl, writeUrl, TmpWorkUrl)
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", errors.Wrap(err, "mount")
	}
	return mntUrl, nil
}

// DeleteWorkSpace Delete the overlay filesystem while container exit
func DeleteWorkSpace(volume, containerId string) {
	mntUrl := fmt.Sprintf(MntUrl, containerId)
	if len(volume) > 0 {
		volumeUrls := volumeUrlExtract(volume)
		if len(volumeUrls) == 2 && len(volumeUrls[0]) > 0 && len(volumeUrls[1]) > 0 {
			DeleteMountPointWithVolume(mntUrl, volumeUrls)
		}
	} else {
		DeleteMountPoint(mntUrl)
	}

	DeleteWriteLayer(containerId)
}

func DeleteMountPoint(mntUrl string) {
	cmd := exec.Command("umount", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Warnt(err.Error())
	}
	// if err := os.RemoveAll(mntUrl); err != nil {
	// 	log.Errorf("Remove dir %s error %v", mntUrl, err)
	// }
}

func DeleteWriteLayer(containerId string) {
	writeUrl := fmt.Sprintf(WriteLayerUrl, containerId)
	if err := os.RemoveAll(writeUrl); err != nil {
		log.Warnf("remove dir: %s, %v", writeUrl, err)
	}
}

func DeleteMountPointWithVolume(mntUrl string, volumeUrls []string) {
	containerUrl := filepath.Join(mntUrl, volumeUrls[1])
	cmd := exec.Command("umount", containerUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Warnf("umount volume: %v", err)
	}

	cmd = exec.Command("umount", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Warnf("Umount mountpoint: %v", err)
	}

	if err := os.RemoveAll(mntUrl); err != nil {
		log.Warnf("Remove mountpoint dir: %s, %v", mntUrl, err)
	}
}

func CreateVolume(containerId, volume string) error {
	mntUrl := fmt.Sprintf(MntUrl, containerId)
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

	volumeUrl := volumeUrls[1]
	containerVolumeUrl := filepath.Join(mntUrl, volumeUrl)
	log.Debugf("volume container url: %s", containerVolumeUrl)
	if utils.IsNotExist(containerVolumeUrl) {
		if err := os.Mkdir(containerVolumeUrl, 0777); err != nil {
			log.Infof("Mkdir container dir %s error. %v", containerVolumeUrl, err)
		}
	}
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

func volumeUrlExtract(volume string) []string {
	var volumeUrls []string
	volumeUrls = strings.Split(volume, ":")
	return volumeUrls
}
