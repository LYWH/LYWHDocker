package container

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path"
	"strings"
)

/*
createTime:LYWH
createData:2021/12/21
*/

//实现数据卷的功能

//
//  mountVolume
//  @Description: 根据volumeUrl解析参数，然后挂在数据卷
//  @param volumeUrl
//  @param mountUrl 挂载地址，此时还没由改变容器的根目录，因此，里面处理的文件操作还是依据宿主机文件系统
//
func mountVolume(volumeUrl, mountUrl string) {
	splitResult, err := extraVolumeUrl(volumeUrl)
	if err != nil {
		initContainerLog.WithFields(logrus.Fields{
			"err":     err,
			"errFrom": "mountVolume",
		})
		return
	}
	//处理根目录
	if has, err := DirOrFileExist(splitResult[0]); err == nil && !has {
		//说明文件夹还不存在，需要创建
		if err = os.Mkdir(splitResult[0], 0755); err != nil {
			initContainerLog.Errorf("mountVolume", err)
		}
	}
	containerUrl := path.Join(mountUrl, splitResult[1])
	if has, err := DirOrFileExist(containerUrl); err == nil && !has {
		//容器内的文件夹还没挂载
		if err = os.Mkdir(containerUrl, 0755); err != nil {
			initContainerLog.Errorf("mountVolume", err)
		}
	}
	//在调用mountVolume时，mnt文件夹已将创立并挂载，因此可以在此基础上继续挂载
	dirs := "dirs=" + splitResult[0]
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "LYWHDockerVolume", containerUrl)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		initContainerLog.Errorf("errfrom", "mountVolume", err)
	}
}

//
//  extraVolumeUrl
//  @Description: 提取volumeUrl中的两部分，判断是否合理
//  @param volumeUrl
//
func extraVolumeUrl(volumeUrl string) ([]string, error) {
	splitResult := strings.Split(volumeUrl, ":")
	if len(splitResult) != 2 || splitResult[0] == "" || splitResult[1] == "" {
		return nil, fmt.Errorf("don't availabe volume args")
	}
	return splitResult, nil
}
