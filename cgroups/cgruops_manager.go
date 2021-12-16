package cgroups

import (
	"LYWHDocker/cgroups/subsystems"
	"LYWHDocker/log"
	"github.com/sirupsen/logrus"
)

/*
createTime:LYWH
createData:2021/12/16
*/

//管理众多cgroup

type CGroupsManager struct {
	name     string
	resource *subsystems.ResourceConfig
}

var cGroupManagerLoger = log.Mylog.WithFields(logrus.Fields{
	"cgroupmanager": "CGroupsManager",
})

func CGroupsManagerCreater(name string, res *subsystems.ResourceConfig) *CGroupsManager {
	return &CGroupsManager{name: name, resource: res}
}

func (manager *CGroupsManager) Set() {
	for _, chain := range subsystems.SubSystemChains {
		if err := chain.Set(manager.name, manager.resource); err != nil {
			cGroupManagerLoger.WithFields(logrus.Fields{
				"method":  "Set",
				"errFrom": "CGroupsManager.Set",
			}).Error(err)
		}
	}
}

func (manager *CGroupsManager) Apply(pid int) {
	for _, chain := range subsystems.SubSystemChains {
		if err := chain.Apply(manager.name, pid); err != nil {
			cGroupManagerLoger.WithFields(logrus.Fields{
				"method":  "Apply",
				"errFrom": "CGroupsManager.Apply",
			}).Error(err)
		}
	}
}

func (manager *CGroupsManager) Remove() {
	for _, chain := range subsystems.SubSystemChains {
		if err := chain.Remove(manager.name); err != nil {
			cGroupManagerLoger.WithFields(logrus.Fields{
				"method":  "Remove",
				"errFrom": "CGroupsManager.Remove",
			}).Error(err)
		}
	}
}
