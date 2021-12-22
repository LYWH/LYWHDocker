package container

import (
	"LYWHDocker/log"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

/*
createTime:LYWH
createData:2021/12/22
*/

var (
	containerLog = "container.log"
)

//
//  GetContainerLogFile
//  @Description:
//  @param containerID
//  @return *os.File
//
func GetContainerLogFile(containerID string) *os.File {
	logPath := path.Join(DefaultInfoLocation, containerID)
	if has, err := DirOrFileExist(logPath); !has && err == nil {
		//文件未存在，还需要创建
		if err = os.MkdirAll(logPath, 0755); err != nil {
			log.Mylog.Error("GetContainerLogFile", "os.MkdirAll", err)
			return nil
		}
	}
	logFile, err := os.Create(path.Join(logPath, containerLog))
	if err != nil {
		log.Mylog.Error("GetContainerLogFile", "os.os.Create", err)
		return nil
	}
	return logFile
}

//
//  OutPutContainerLog
//  @Description:
//  @param containerID
//
func OutPutContainerLog(containerID string) {
	logFile, err := os.OpenFile(path.Join(DefaultInfoLocation, containerID, containerLog), os.O_RDONLY, 0644)
	if err != nil {
		log.Mylog.Error("outPutContainerLog", "os.OpenFile", err)
		return
	}
	defer logFile.Close()
	content, err := ioutil.ReadAll(logFile)
	if err != nil {
		log.Mylog.Error("outPutContainerLog", "ioutil.ReadAll", err)
		return
	}
	_, err = fmt.Fprint(os.Stdout, string(content))
	if err != nil {
		log.Mylog.Error("outPutContainerLog", "fmt.Fprint", err)
		return
	}
}
