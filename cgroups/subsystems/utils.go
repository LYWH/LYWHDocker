package subsystems

import (
	"LYWHDocker/log"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path"
)

/*
createTime:LYWH
createData:2021/12/16
*/

//找到cgroup对应的路径,/sys/fs/cgroup/是路径前缀，

var utilLoger = log.Mylog.WithFields(logrus.Fields{
	"part": "util",
})

func getCGroupPath(subSystemType string, cgroupName string, autoCreate bool) (string, error) {
	rootPath := "/sys/fs/cgroup/"
	absolutePath := path.Join(rootPath, subSystemType, cgroupName)
	utilLoger.Info("=====", absolutePath, rootPath, subSystemType, cgroupName)
	//if _, err := os.Stat(absolutePath); err == nil ||( autoCreate &&os.IsNotExist(err)) { //判断文件夹是否存在
	//	//创建文件夹
	//	if err := os.Mkdir(absolutePath, 0755); err != nil {
	//		utilLoger.WithFields(logrus.Fields{
	//			"errFrom":"getCGroupPath",
	//			"path":absolutePath,
	//		}).Error(err)
	//		return "", fmt.Errorf("error ai createing %v", err)
	//	}
	//}
	//return absolutePath, nil
	if _, err := os.Stat(absolutePath); err == nil || (autoCreate && os.IsNotExist(err)) {
		if os.IsNotExist(err) {
			// 创建文件夹
			if err := os.Mkdir(absolutePath, 0755); err != nil {
				return "", fmt.Errorf("error create cgroup dir %v", err)
			}
			return absolutePath, nil
		}
		return absolutePath, nil
	} else {
		// 如果os.Stat是其他错误或者不存在cgroup目录但是也没有设置自动创建，则返回错误
		return "", fmt.Errorf("cgroup path error %v", err)
	}
}
