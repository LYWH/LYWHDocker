package command

import (
	"LYWHDocker/cgroups"
	"LYWHDocker/cgroups/subsystems"
	"LYWHDocker/container"
	"LYWHDocker/log"
	"LYWHDocker/network"
	"github.com/sirupsen/logrus"
	"os"
)

/*
createTime:LYWH
createData:2021/12/27
*/

//
//  RunContainer
//  @Description:
//  @param tty
//  @param command
//

func RunContainer(tty bool, cmd string, cgroupsManagerName string, res *subsystems.ResourceConfig, Volume string, containerName string,
	containerID string, imagePath string, enVar []string, networkName string, port []string) {
	process, writer, workSpaceRelatePath := container.GetParentProcess(tty, Volume, containerID, imagePath, enVar)
	if err := process.Start(); err != nil {
		log.Mylog.WithField("method", "syscall.Mount").Error(err)
		log.Mylog.WithFields(logrus.Fields{
			"err":     err,
			"errfrom": "RunContainer",
		})
		//开始限制资源
	}
	//记录容器信息
	containerinfo, err := container.RecordContainerInfo(process.Process.Pid, []string{cmd}, containerName, containerID, port)
	if err != nil {
		log.Mylog.Error("recordContainerInfo", err)
		return
	}

	if networkName != "" {
		//初始化网络
		if err := network.Init(); err != nil {
			log.Mylog.Error("error at init network\n", err)
			return
		}
		if err := network.Connect(networkName, containerinfo); err != nil {
			log.Mylog.Error("error at connect network\n", err)
			return
		}
	}

	err = container.SendCommand(writer, cmd)
	if err != nil {
		return
	}
	cgroupsManager := cgroups.CGroupsManagerCreater(cgroupsManagerName+"_"+containerID, res)
	cgroupsManager.Set()
	cgroupsManager.Apply(process.Process.Pid)
	if tty {
		if err = process.Wait(); err != nil {
			log.Mylog.Error(err)
		}
		container.DeleteConfigInfo(containerID)
		cgroupsManager.Remove()
		container.DeleteWorkSpace(workSpaceRelatePath)
		os.Exit(1)
	}

}
