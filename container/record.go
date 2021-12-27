package container

import (
	"LYWHDocker/log"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"text/tabwriter"
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
	Pid         string
	Id          string
	Name        string
	Command     string
	CreateTime  string
	Status      string
	PortMapping []string //端口映射
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
func GenerateContainerID(n int) string {
	if n < 0 || n > 32 {
		n = 32
	}
	hashByte := sha256.Sum256([]byte(strconv.Itoa(int(time.Now().UnixNano()))))
	return fmt.Sprintf("%x", hashByte[:n])
}

func RecordContainerInfo(containerPID int, cmd []string, containerName string, containerID string) (*ContainerInfo, error) {
	//containerID := GenerateContainerID(IDLength)
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
		return nil, err
	}
	//将相应的信息写入对应文件夹中
	//创建文件夹
	infoPath := path.Join(DefaultInfoLocation, containerID)
	if err = os.MkdirAll(infoPath, 0711); err != nil {
		log.Mylog.Error("infoPath", "os.MkdirAll", err)
		return nil, err
	}
	infoJsonFile := path.Join(infoPath, ConfigName)
	//创建json文件
	configFile, err := os.OpenFile(infoJsonFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0711)
	defer configFile.Close()
	if err != nil {
		log.Mylog.Error("infoJsonFile", "os.OpenFile", err)
		return nil, err
	}
	//将创建的json文件写入其中
	if _, err = configFile.WriteString(string(jsonByte)); err != nil {
		log.Mylog.Error("configFile.WriteString", jsonByte, err)
	}
	return &containerInfo, nil
}

func DeleteConfigInfo(containerId string) {
	dirUrl := path.Join(DefaultInfoLocation, containerId)
	if err := os.RemoveAll(dirUrl); err != nil {
		log.Mylog.Error("os.RemoveAll", "deleteConfigInfo", err)
	}
}

//
//  OutputContainerInfo
//  @Description: 格式化展示容器信息
//
func OutputContainerInfo() {
	files, err := ioutil.ReadDir(DefaultInfoLocation)
	if err != nil {
		log.Mylog.Error("outputContainerInfo", "ioutil.ReadDir", err)
		return
	}
	var containersInfo []*ContainerInfo

	for _, file := range files { //读取DefaultInfoLocation文件夹下所有文件夹的内容
		containerInfo, err := getContainerInfo(file.Name())
		if err != nil {
			log.Mylog.Error("getContainerInfo", err)
			return
		}
		containersInfo = append(containersInfo, containerInfo)
	}
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, item := range containersInfo {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id, item.Name, item.Pid, item.Status, item.Command, item.CreateTime)
	}
	if err = w.Flush(); err != nil {
		log.Mylog.Error("w.Flush", err)
	}
}

//
//  getContainerInfo
//  @Description: 读取某一个进程的信息
//  @param containerID
//  @return *ContainerInfo
//  @return error
//
func getContainerInfo(containerID string) (*ContainerInfo, error) {
	containerPath := path.Join(DefaultInfoLocation, containerID, ConfigName)
	if has, err := DirOrFileExist(containerPath); !has && err == nil {
		log.Mylog.Error("getContainerInfo", containerPath, err)
		return nil, err
	}
	var containerInfo ContainerInfo
	//读取相应的Json文件
	content, err := ioutil.ReadFile(containerPath)
	if err != nil {
		log.Mylog.Error("getContainerInfo", "ioutil.ReadFile", err)
		return nil, err
	}
	if err = json.Unmarshal(content, &containerInfo); err != nil {
		log.Mylog.Error("getContainerInfo", "json.Unmarshal", err)
		return nil, err
	}
	return &containerInfo, nil
}
