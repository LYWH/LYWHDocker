package subsystems

import (
	"fmt"
	"os"
	"path"
)

/*
createTime:LYWH
createData:2021/12/16
*/

//找到cgroup对应的路径,/sys/fs/cgroup/是路径前缀，
func getCGroupPath(subSystemType string, cgroupName string, autoCreate bool) (string, error) {
	rootPath := "/sys/fs/cgroup/"
	absolutePath := path.Join(rootPath, subSystemType, cgroupName)
	if _, err := os.Stat(absolutePath); err == nil && autoCreate { //判断文件夹是否存在
		//创建文件夹
		if err := os.Mkdir(absolutePath, 0755); err != nil {
			return "", fmt.Errorf("error ai createing %v", err)
		}
	}
	return absolutePath, nil
}
