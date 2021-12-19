package container

import (
	"LYWHDocker/log"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
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

	setUpMount()

	absolutePath, err := exec.LookPath(cmds[0])
	if err != nil {
		initContainerLog.WithFields(logrus.Fields{
			"err": "error command,can't find it",
		})
	}
	if err := syscall.Exec(absolutePath, cmds, os.Environ()); err != nil {
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

//
//  povit_root
//  @Description:更改当前容器的root目录
//  @param newRoot
//  @return error
//
func povitRoot(newRoot string) error {
	//1:利用mount将newRoot生成一个挂载点
	if err := syscall.Mount(newRoot, newRoot, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		initContainerLog.WithFields(logrus.Fields{
			"errerFrom": "povitRoot",
			"err":       err,
		})
		return err
	}
	//2:在newRoot目录下生成.old_root文件夹，用于povit_root切换工作目录
	oldPath := path.Join(newRoot, ".old_root")
	if _, err := os.Stat(oldPath); err == nil {
		//说明目录存在，需要删除
		if err := os.RemoveAll(oldPath); err != nil {
			initContainerLog.WithFields(logrus.Fields{
				"errFrom": "povitRoot",
				"err":     err,
			})
		}
		return err
	}
	if err := os.Mkdir(oldPath, 0755); err != nil {
		initContainerLog.WithFields(logrus.Fields{
			"errFrom": "povitRoot",
			"err":     err,
		})
		return err
	}
	//3使用povit_root切换工作目录
	if err := syscall.PivotRoot(newRoot, oldPath); err != nil {
		initContainerLog.WithFields(logrus.Fields{
			"errFrom": "povitRoot",
			"err":     err,
		})
		return err
	}
	//4 修改当前的工作目录到切换后的根目录
	if err := os.Chdir("/"); err != nil {
		initContainerLog.WithFields(logrus.Fields{
			"errFrom": "povitRoot",
			"err":     err,
		})
		return err
	}
	//5 umount原目录
	oldPath = path.Join("/", ".old_root") //由于工作目录已经切换，所以需要更改oldPath目录
	if err := syscall.Unmount(oldPath, syscall.MNT_DETACH); err != nil {
		initContainerLog.WithFields(logrus.Fields{
			"errFrom": "povitRoot",
			"err":     err,
		})
		return err
	}
	return nil
}

func setUpMount() error {
	//将启动目录mount为当前环境的系统目录
	//将系统根目录设置为私有模式，因为povitRoot目录中有许多mount系统调用,防止容器与宿主机相互影响
	if err := syscall.Mount("/", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		initContainerLog.WithFields(logrus.Fields{
			"errFrom": "setUpMount",
			"err":     err,
		})
		return err
	}
	//获取启动目录
	startUpPath, err := os.Getwd()
	fmt.Println("启动目录:", startUpPath)
	if err != nil {
		initContainerLog.WithFields(logrus.Fields{
			"errFrom": "setUpMount",
			"err":     err,
		})
	}
	//切换系统目录
	if err := povitRoot(startUpPath); err != nil {
		initContainerLog.WithFields(logrus.Fields{
			"errFrom": "setUpMount",
			"err":     err,
		})
	}
	//设置挂载点
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	if err := syscall.Mount("", "/proc", "proc", uintptr(defaultMountFlags), ""); err != nil {
		initContainerLog.WithFields(logrus.Fields{
			"errFrom": "setUpMount",
			"err":     err,
		})
		return err
	}
	return nil
}
