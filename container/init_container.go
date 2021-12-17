package container

import (
	"LYWHDocker/log"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

//主要使用系统调用实现资源隔离与限制，最后使用
//
//  InitNewNameSpace
//  @Description:
//  @param cmd
//  @param args
//  @return error
//
var initContainerLog = log.Mylog.WithFields(logrus.Fields{
	"part": "initcontainer",
})

func InitNewNameSpace() error {
	//先将挂载方式设置成私有方式，方式新的namespace中挂载后影响宿主机的proc
	cmds := getCommands()
	privateMountFlag := syscall.MS_REC | syscall.MS_PRIVATE
	//设置传递和私有的挂载方式，使得新的namespace中挂载后影响宿主机的proc
	if err := syscall.Mount("", "/proc", "proc", uintptr(privateMountFlag), ""); err != nil {
		log.Mylog.WithField("method", "syscall.Mount").Error(err)
		return err
	}
	//syscall.MS_NOEXEC:不允许执行在这个文件系统执行程序
	// syscall.MS_NOSUID:当从这个文件系统执行程序时，不可使用set-user-ID和set-group-ID
	//syscall.MS_NODEV:不允许访问特殊设备
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	if err := syscall.Mount("", "/proc", "proc", uintptr(defaultMountFlags), ""); err != nil {
		log.Mylog.WithField("method", "syscall.Mount").Error(err)
		return err
	}

	//使用cmd程序替换掉init初始化进程
	//此处先要去寻找命令的绝对路径

	path, err := exec.LookPath(cmds[0])
	if err != nil {
		initContainerLog.WithFields(logrus.Fields{
			"err": "error command,can't find it",
		})
	}
	if err := syscall.Exec(path, cmds, os.Environ()); err != nil {
		log.Mylog.WithField("method", "syscall.Mount").Error(err)
		return err
	}
	return nil
}

func getCommands() []string {
	//由于标准输入、输出和错误三个文件描述符否是默认被子进程继承的，因此管道文件描述符的index为3
	reader := os.NewFile(uintptr(3), "pipe")
	if reader == nil {
		initContainerLog.WithFields(logrus.Fields{
			"err": "no pipe",
		})
		return nil
	}
	cmds, err := ioutil.ReadAll(reader)
	if err != nil {
		initContainerLog.WithFields(logrus.Fields{
			"err": "no commands",
		})
		return nil
	}
	//按照空行分割cmd，此处ioutil.ReadAll返回的是[]byte类型
	return strings.Split(string(cmds), " ")
}
