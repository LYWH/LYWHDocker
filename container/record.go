package container

import (
	"LYWHDocker/log"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

/*
createTime:LYWH
createData:2021/12/21
*/
//负责记录容器的各种信息，用于docker ps读取
const (
	IDLength = 15
)

type ContainerInfo struct {
	Pid        string
	Id         string
	Name       string
	Command    string
	CreateTime string
	Status     string
}

var (
	RUNNING             = "running"
	STOP                = "stop"
	EXITED              = "extied"
	DefaultInfoLocation = "/var/run/LYWHDocker"
	ConfigName          = "containerInfo.json"
)

//
//  generateContainerID
//  @Description: 生成容器ID
//  @param n 容器id长度
//  @return string
//
func generateContainerID(n int) string {
	if n < 0 || n > 32 {
		n = 32
	}
	hashByte := sha256.Sum256([]byte(strconv.Itoa(int(time.Now().UnixNano()))))
	return fmt.Sprintf("%x", hashByte[:n])
}

func recordContainerInfo(containerPID int, cmd []string, containerName string) (string, error) {
	containerID := generateContainerID(IDLength)
	createTime := time.Now().Format("2006-01-02 15:04:05")
	if containerName == "" {
		containerName = containerID
	}
	//容器信息变量
	containerInfo := ContainerInfo{
		Pid:        strconv.Itoa(containerPID),
		Id:         containerID,
		Name:       containerName,
		Command:    strings.Join(cmd, " "),
		CreateTime: createTime,
		Status:     RUNNING,
	}
	//序列化为json
	jsonByte, err := json.Marshal(containerInfo)
	if err != nil {
		log.Mylog.Error("containerInfo", "Marshal", err)
		return "", err
	}
	//将相应的信息写入对应文件夹中
	//创建文件夹
	infoPath := path.Join(DefaultInfoLocation, containerID)
	if err = os.MkdirAll(infoPath, 0711); err != nil {
		log.Mylog.Error("infoPath", "os.MkdirAll", err)
		return "", err
	}
	infoJsonFile := path.Join(infoPath, ConfigName)
	//创建json文件
	configFile, err := os.OpenFile(infoJsonFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0711)
	defer configFile.Close()
	if err != nil {
		log.Mylog.Error("infoJsonFile", "os.OpenFile", err)
		return "", err
	}
	//将创建的json文件写入其中
	if _, err = configFile.WriteString(string(jsonByte)); err != nil {
		log.Mylog.Error("configFile.WriteString", jsonByte, err)
	}
	return containerID, nil
}

func deleteConfigInfo(containerId string) {
	dirUrl := path.Join(DefaultInfoLocation, containerId)
	if err := os.RemoveAll(dirUrl); err != nil {
		log.Mylog.Error("os.RemoveAll", "deleteConfigInfo", err)
	}
}
