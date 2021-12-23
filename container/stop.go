package container

import (
	"LYWHDocker/log"
	"encoding/json"
	"io/ioutil"
	"path"
	"strconv"
	"syscall"
)

/*
createTime:LYWH
createData:2021/12/23
*/

func StopContainerByID(containID string) {
	containerInfo, err := getContainerInfo(containID)
	if err != nil {
		log.Mylog.Error("getContainerInfo", "stopContainerByID", err)
		return
	}
	containerPid, _ := strconv.Atoi(containerInfo.Pid)
	if err = syscall.Kill(containerPid, syscall.SIGTERM); err != nil {
		log.Mylog.Error("syscall.Kill", "stopContainerByID", err)
		return
	}
	containerInfo.Status = STOP
	containerInfo.Pid = " " //需要重置PID
	containerInfoBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Mylog.Error("json.Marshal", "stopContainerByID", err)
	}
	filePath := path.Join(DefaultInfoLocation, containerInfo.Id, ConfigName)
	if err = ioutil.WriteFile(filePath, containerInfoBytes, 0711); err != nil {
		log.Mylog.Error("ioutil.WriteFile", "stopContainerByID", err)
	}

}
