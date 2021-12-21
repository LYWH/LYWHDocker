package container

import (
	"github.com/sirupsen/logrus"
	"os/exec"
)

/*
createTime:LYWH
createData:2021/12/21
*/

func CommitContainer(imageName string) {
	mntUrl := "./mnt"
	imageTarUrl := "./" + imageName + ".tar"
	if _, err := exec.Command("tar", "-czf", imageTarUrl, mntUrl).CombinedOutput(); err != nil {
		initContainerLog.WithFields(logrus.Fields{
			"err":     err,
			"errFrom": "commitContainer",
		})
	}
}
