package container

import (
	"LYWHDocker/config"
	"github.com/sirupsen/logrus"
	"os/exec"
	"path"
)

/*
createTime:LYWH
createData:2021/12/21
*/

func CommitContainer(containerID, imageName string) {
	mntUrl := path.Join(config.RootUrl, "mnt", containerID)
	imageTarUrl := "./" + imageName + ".tar"
	if _, err := exec.Command("tar", "-czf", imageTarUrl, "-C", mntUrl, ".").CombinedOutput(); err != nil {
		initContainerLog.WithFields(logrus.Fields{
			"err":     err,
			"errFrom": "commitContainer",
		})
	}
}
