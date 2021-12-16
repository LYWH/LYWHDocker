package container

import (
	"LYWHDocker/log"
	"os"
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
func InitNewNameSpace(cmd string, args []string) error {
	//先将挂载方式设置成私有方式，方式新的namespace中挂载后影响宿主机的proc
	log.Mylog.Info("====", cmd, args)
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

	argsv := []string{cmd}
	//使用cmd程序替换掉init初始化进程
	if err := syscall.Exec(cmd, argsv, os.Environ()); err != nil {
		log.Mylog.WithField("method", "syscall.Mount").Error(err)
		return err
	}
	return nil
}