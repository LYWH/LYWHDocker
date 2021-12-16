package subsystems

import (
	"LYWHDocker/log"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

/*
createTime:LYWH
createData:2021/12/16
*/

//实现对memory资源的限制
type MemoeySubSystem struct {
}

const (
	memoryLimitFile = "memory.limit_in_bytes"
	taskFile        = "tasks"
)

var memoryLog = log.Mylog.WithFields(logrus.Fields{
	"subsystem": "memory",
})

func (m *MemoeySubSystem) Name() string {
	return "memory"
}

//设置内存限制，原理是：在cgroup路径下写入限制内存大小的参数
func (m *MemoeySubSystem) Set(cgroupName string, res *ResourceConfig) error {
	//找到对应的路径
	subCGroupPath, err := getCGroupPath(m.Name(), cgroupName, true)
	if err != nil {
		memoryLog.WithFields(logrus.Fields{
			"method":  "Set",
			"errfrom": "getCGroupPath",
		}).Error(err) //返回错误信息
	}
	//资源限制，原理：将内存限制写入subCGroupPath的memory.limit_in_bytes文件中
	if res.Memory != "" {
		if err := ioutil.WriteFile(path.Join(subCGroupPath, memoryLimitFile), []byte(res.Memory), 0644); err != nil {
			memoryLog.WithFields(logrus.Fields{
				"method":  "Set",
				"errFrom": "ioutil.WriteFile",
			}).Error(err)
			return err
		}
	}
	return nil
}

//将进程加入cgroup，原理：将进程id加入到task文件中
func (m *MemoeySubSystem) Apply(cgroupName string, pid int) error {
	subCGroupPath, err := getCGroupPath(m.Name(), cgroupName, false)
	if err != nil {
		memoryLog.WithFields(logrus.Fields{
			"method":  "Apply",
			"errFrom": "getCGroupPath",
		}).Error(err)
		return err
	}
	if err := ioutil.WriteFile(path.Join(subCGroupPath, taskFile), []byte(strconv.Itoa(pid)), 0644); err != nil {
		memoryLog.WithFields(logrus.Fields{
			"method":  "Apply",
			"errFrom": "getCGroupPath",
		}).Error(err)
	}
	return nil
}

func (m *MemoeySubSystem) Remove(cgroupName string) error {
	subCGroupPath, err := getCGroupPath(m.Name(), cgroupName, false)
	if err != nil {
		memoryLog.WithFields(logrus.Fields{
			"method":  "Apply",
			"errFrom": "Remove",
		}).Error(err)
		return err
	}
	if err := os.RemoveAll(subCGroupPath); err != nil {
		memoryLog.WithFields(logrus.Fields{
			"method":  "Apply",
			"errFrom": "RemoveAll",
		}).Error(err)
	}
	return nil
}
