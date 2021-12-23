package container

import (
	"LYWHDocker/config"
	"LYWHDocker/log"
	"os"
	"path"
	"strings"
)

/*
createTime:LYWH
createData:2021/12/23
*/

func RemoveContainerByCID(containerID string) {
	containerInfo, err := getContainerInfo(containerID)
	if err != nil {
		log.Mylog.Error("getContainerInfo", "RemoveContainerByCID", err)
		return
	}
	if strings.Compare(containerInfo.Status, STOP) == 0 { //判断容器是否停止
		mountLayer := path.Join(config.RootUrl, "mnt", containerID)
		readAndWriteLayer := path.Join(config.RootUrl, "diff", containerID)
		deleteWorkSpace([]string{mountLayer, readAndWriteLayer})
		containerRecordAndLogPath := path.Join(DefaultInfoLocation, containerInfo.Id)
		if err = os.RemoveAll(containerRecordAndLogPath); err != nil {
			log.Mylog.Error("os.RemoveAll", "RemoveContainerByCID", err)
		}
	} else {
		log.Mylog.Warn("errer container ID")
	}
}
