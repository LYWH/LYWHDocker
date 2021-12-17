package container

import (
	"LYWHDocker/cgroups"
	"LYWHDocker/cgroups/subsystems"
	"LYWHDocker/log"
	"github.com/sirupsen/logrus"
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

var runContainerLog = log.Mylog.WithFields(logrus.Fields{
	"part": "runcontainer",
})

func getParentProcess(tty bool) (*exec.Cmd, *os.File) {
	//此处也不需要传递命令参数，命令的传递需要通过专门的发送和接收函数
	//生成管道
	reader, writer, err := getPip()
	if err != nil {
		runContainerLog.WithFields(logrus.Fields{
			"err": err,
		})
		return nil, nil
	}
	//argsAll := []string{"init", command}
	// /proc/self代表当前进程运行的环境，exe代表本程序的启动命令(一个链接文件)，在command中使用/proc/self/exe代表自己调用自己
	//通过这种方式实现了fork一个新的进程，在参数中：先使用init调用初始化容器函数InitNewNameSpace，在里面进行挂载设置
	cmd := exec.Command("/proc/self/exe", "init")
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
	//指定子进程继承管道的reader端
	cmd.ExtraFiles = []*os.File{reader}
	return cmd, writer
}

//
//  getPip
//  @Description:
//  @return *os.File
//  @return *os.File
//  @return error
//
func getPip() (*os.File, *os.File, error) {
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return reader, writer, err
}

func sendCommand(writer *os.File, cmd string) error {
	if _, err := writer.WriteString(cmd); err != nil {

		runContainerLog.WithFields(logrus.Fields{
			"err": err,
		})
		return err
	}
	if err := writer.Close(); err != nil {
		runContainerLog.WithFields(logrus.Fields{
			"err": err,
		})
		return err
	}
	return nil
}

//
//  RunContainer
//  @Description:
//  @param tty
//  @param command
//
func RunContainer(tty bool, cmd string, cgroupsManagerName string, res *subsystems.ResourceConfig) {
	process, writer := getParentProcess(tty)
	if err := process.Start(); err != nil {
		log.Mylog.WithField("method", "syscall.Mount").Error(err)
		//开始限制资源
	}
	err := sendCommand(writer, cmd)
	if err != nil {
		return
	}
	cgroupsManager := cgroups.CGroupsManagerCreater(cgroupsManagerName, res)
	cgroupsManager.Set()
	cgroupsManager.Apply(process.Process.Pid)
	if err := process.Wait(); err != nil {
		log.Mylog.Error(err)
	}
	cgroupsManager.Remove()
	os.Exit(1)
}
