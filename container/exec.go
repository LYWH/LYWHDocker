package container

import (
	"LYWHDocker/log"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
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
	env := getEnvByPID(pid)
	if env != nil {
		//将容器进程的环境变量放入到exec进程中，第二次调用自己时新的子进程继承这些环境变量
		//加入系统环境变量的原因是：EXEC_ENV_PROCESS_ID、EXEC_ENV_PROCESS_CMD被设置成系统环境变量，加入系统环境变量后第二次执行才可以读取到
		cmd.Env = append(os.Environ(), env...)
	}
	if err := cmd.Run(); err != nil {
		log.Mylog.Error("EnterContainer", "run", err)
		return
	}
}

func getEnvByPID(PID string) []string {
	envPath := path.Join("/proc", PID, "environ")
	contentBytes, err := ioutil.ReadFile(envPath)
	if err != nil {
		log.Mylog.Error("getEnvByPID", "ioutil.ReadFile", err)
		return nil
	}
	//在environ文件中，分隔符使用的是\u0000
	env := strings.Split(string(contentBytes), "\u0000")
	return env
}

func getProecessIDbyCID(containerID string) string {
	containerInfo, err := getContainerInfo(containerID)
	if err != nil {
		log.Mylog.Error("getProecessIDbyCID", getContainerInfo, err)
		return ""
	}
	return containerInfo.Pid
}
