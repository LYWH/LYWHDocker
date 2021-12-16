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
type CpuSubSystem struct {
}

const (
	cpuLimitFile = "cpu.shares"
	//由于cpu与mamory在一个包内，所以task不需要声明
)

var cpuLog = log.Mylog.WithFields(logrus.Fields{
	"subsystem": "cpu",
})

func (cpu *CpuSubSystem) Name() string {
	return "cpu,cpuacct"
}
func (cpu *CpuSubSystem) Set(cgroupName string, res *ResourceConfig) error {
	subCGroupPath, err := getCGroupPath(cpu.Name(), cgroupName, true)
	if err != nil {
		cpuLog.WithFields(logrus.Fields{
			"method":  "Set",
			"errfrom": "getCGroupPath",
		}).Error(err) //返回错误信息
	}
	if res.CpuShare != "" {
		if err := ioutil.WriteFile(path.Join(subCGroupPath, cpuLimitFile), []byte(res.CpuShare), 0644); err != nil {
			cpuLog.WithFields(logrus.Fields{
				"method":  "Set",
				"errFrom": "ioutil.WriteFile",
			}).Error(err)
			return err
		}
	}
	return nil
}
func (cpu *CpuSubSystem) Apply(cgroupName string, pid int) error {
	subCGroupPath, err := getCGroupPath(cpu.Name(), cgroupName, false)
	if err != nil {
		cpuLog.WithFields(logrus.Fields{
			"method":  "Apply",
			"errFrom": "getCGroupPath",
		}).Error(err)
		return err
	}
	if err := ioutil.WriteFile(path.Join(subCGroupPath, taskFile), []byte(strconv.Itoa(pid)), 0644); err != nil {
		cpuLog.WithFields(logrus.Fields{
			"method":  "Apply",
			"errFrom": "getCGroupPath",
		}).Error(err)
	}
	return nil
}
func (cpu *CpuSubSystem) Remove(cgroupName string) error {
	subCGroupPath, err := getCGroupPath(cpu.Name(), cgroupName, false)
	if err != nil {
		cpuLog.WithFields(logrus.Fields{
			"method":  "Apply",
			"errFrom": "Remove",
		}).Error(err)
		return err
	}
	if err := os.RemoveAll(subCGroupPath); err != nil {
		cpuLog.WithFields(logrus.Fields{
			"method":  "Apply",
			"errFrom": "RemoveAll",
		}).Error(err)
	}
	return nil
}
