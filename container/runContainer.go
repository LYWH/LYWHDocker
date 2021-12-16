package container

import (
	"LYWHDocker/log"
	"os"
	"os/exec"
	"syscall"
)

//
//  getParentProcess
//  @Description:
//  @param tty
//  @param command
//  @return *exec.Cmd
//
func getParentProcess(tty bool, command string) *exec.Cmd {
	argsAll := []string{"init", command}
	log.Mylog.Info("argsall", argsAll)
	// /proc/self代表当前进程运行的环境，exe代表本程序的启动命令(一个链接文件)，在command中使用/proc/self/exe代表自己调用自己
	//通过这种方式实现了fork一个新的进程，在参数中：先使用init调用初始化容器函数InitNewNameSpace，在里面进行挂载设置
	cmd := exec.Command("/proc/self/exe", argsAll...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWPID | syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWNET,
	}

	//syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
	//	syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC
	//启用tty需要将os的标准流重定向
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd
}

//
//  RunContainer
//  @Description:
//  @param tty
//  @param command
//
func RunContainer(tty bool, command string) {
	process := getParentProcess(tty, command)
	if err := process.Run(); err != nil {
		log.Mylog.WithField("method", "syscall.Mount").Error(err)
	}
}
