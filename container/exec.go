package container

import (
	"LYWHDocker/log"
	"os"
	"os/exec"
	"strings"
)

const (
	EXEC_ENV_PROCESS_ID  = "exec_pid"
	EXEC_ENV_PROCESS_CMD = "exec_cmd"
)

/*
createTime:LYWH
createData:2021/12/22
*/

func EnterContainer(containerID string, containerCMD []string) {
	//根据容器ID获取进程ID
	pid := getProecessIDbyCID(containerID)
	if len(pid) == 0 || strings.Compare(pid, " ") == 0 {
		log.Mylog.Error("EnterContainer", "getProecessIDbyCID")
		return
	}
	cmdStr := strings.Join(containerCMD, " ")
	log.Mylog.Info(pid)
	log.Mylog.Info(cmdStr)

	cmd := exec.Command("/proc/self/exe", "exec")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	//设置环境变量
	if err := os.Setenv(EXEC_ENV_PROCESS_ID, pid); err != nil {
		log.Mylog.Error("os.Setenv", "EnterContainer", err)
		return
	}
	if err := os.Setenv(EXEC_ENV_PROCESS_CMD, cmdStr); err != nil {
		log.Mylog.Error("os.Setenv", "EnterContainer", err)
		return
	}
	if err := cmd.Run(); err != nil {
		log.Mylog.Error("EnterContainer", "run", err)
		return
	}
}

func getProecessIDbyCID(containerID string) string {
	containerInfo, err := getContainerInfo(containerID)
	if err != nil {
		log.Mylog.Error("getProecessIDbyCID", getContainerInfo, err)
		return ""
	}
	return containerInfo.Pid
}
