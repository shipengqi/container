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

func volumeUrlExtract(volume string) []string {
	var volumeUrls []string
	volumeUrls = strings.Split(volume, ":")
	return volumeUrls
}