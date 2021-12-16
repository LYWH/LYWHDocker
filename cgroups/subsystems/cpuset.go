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

type CpuSetSubSystem struct {
}

const (
	cpuSetCpuLimitFile    = "cpuset.cpus"
	cpuSetMemoryLimitFile = "cpuset.mems"
)

var cpuSetLog = log.Mylog.WithFields(logrus.Fields{
	"subSystem": "cpuset",
})

func (cpuset *CpuSetSubSystem) Name() string {
	return "cpuset"
}

func (cpuset *CpuSetSubSystem) Set(cgroupName string, res *ResourceConfig) error {
	subCGroupPath, err := getCGroupPath(cpuset.Name(), cgroupName, true)
	if err != nil {
		cpuSetLog.WithFields(logrus.Fields{
			"method":  "Set",
			"errfrom": "getCGroupPath",
		}).Error(err) //返回错误信息
	}
	if res.CpuSet != "" {
		if err := ioutil.WriteFile(path.Join(subCGroupPath, cpuSetCpuLimitFile), []byte(res.CpuSet), 0644); err != nil {
			cpuSetLog.WithFields(logrus.Fields{
				"method":  "Set",
				"errFrom": "ioutil.WriteFile",
			}).Error(err)
			return err
		}
	}

	if res.CpuMems != "" {
		if err := ioutil.WriteFile(path.Join(subCGroupPath, cpuSetMemoryLimitFile), []byte(res.CpuMems), 0644); err != nil {
			cpuSetLog.WithFields(logrus.Fields{
				"method":  "Set",
				"errFrom": "ioutil.WriteFile",
			}).Error(err)
			return err
		}
	}
	return nil
}

func (cpuset *CpuSetSubSystem) Apply(cgroupName string, pid int) error {
	subCGroupPath, err := getCGroupPath(cpuset.Name(), cgroupName, false)
	if err != nil {
		cpuSetLog.WithFields(logrus.Fields{
			"method":  "Apply",
			"errFrom": "getCGroupPath",
		}).Error(err)
		return err
	}
	if err := ioutil.WriteFile(path.Join(subCGroupPath, taskFile), []byte(strconv.Itoa(pid)), 0644); err != nil {
		cpuSetLog.WithFields(logrus.Fields{
			"method":  "Apply",
			"errFrom": "getCGroupPath",
		}).Error(err)
	}
	return nil
}

func (cpuset *CpuSetSubSystem) Remove(cgroupName string) error {
	subCGroupPath, err := getCGroupPath(cpuset.Name(), cgroupName, false)
	if err != nil {
		cpuSetLog.WithFields(logrus.Fields{
			"method":  "Apply",
			"errFrom": "Remove",
		}).Error(err)
		return err
	}
	if err := os.RemoveAll(subCGroupPath); err != nil {
		cpuSetLog.WithFields(logrus.Fields{
			"method":  "Apply",
			"errFrom": "RemoveAll",
		}).Error(err)
	}
	return nil
}
