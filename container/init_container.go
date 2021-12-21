package container

import (
	"LYWHDocker/log"
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
			"err":     "error command,can't find it",
			"errFrom": "InitNewNameSpace",
		})
	}
	if err = syscall.Exec(absolutePath, cmds, os.Environ()); err != nil {
		//log.Mylog.WithField("method", "syscall.Mount").Error(err)
		initContainerLog.WithFields(logrus.Fields{
			"err":     err,
			"errfrom": "InitNewNameSpace",
		})
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
		if err = os.RemoveAll(oldPath); err != nil {
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

	if err != nil {
		initContainerLog.WithFields(logrus.Fields{
			"errFrom": "setUpMount",
			"err":     err,
		})
	}
	//切换系统目录
	if err = povitRoot(startUpPath); err != nil {
		initContainerLog.WithFields(logrus.Fields{
			"errFrom": "setUpMount",
			"err":     err,
		})
	}
	//设置挂载点
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	if err = syscall.Mount("", "/proc", "proc", uintptr(defaultMountFlags), ""); err != nil {
		initContainerLog.WithFields(logrus.Fields{
			"errFrom": "setUpMount",
			"err":     err,
		})
		return err
	}
	return nil
}

//
//  newWorkSpace
//  @Description: 为容器创建新的工作空间
//  @param rootUrl
//  @param imageName
//
func newWorkSpace(rootUrl, imageName, volumeUrl, containerID string) []string {
	readOnlyLayer := createReadOnlyLayer(rootUrl, imageName)
	readAndWriteLayer := createReadAndWriteLayer(rootUrl, containerID)
	mntLayer := createMountLayer(rootUrl, readOnlyLayer, readAndWriteLayer, containerID)
	if _, err := extraVolumeUrl(volumeUrl); err == nil {
		mountVolume(volumeUrl, mntLayer)
	}
	return []string{mntLayer, readAndWriteLayer}
}

//
//  createReadOnlyLayer
//  @Description: 创建只读层，解压busybox.tar文件内容到只读层,逻辑是如果没有该文件夹，就需要创建，并将该镜像内容解压到该文件夹，如果有，则可以直接返回
//  @param rootUrl
//  @param imageName
//  @return string
//
func createReadOnlyLayer(rootUrl, imageName string) string {
	imageName = strings.Trim(imageName, "/")
	//readOnlyPath := rootUrl + imageName + "/"
	readOnlyPath := path.Join(rootUrl, "diff", imageName)
	tarName := "./" + imageName + ".tar" //需要保证当前目录下有busybox.tar
	if has, err := DirOrFileExist(readOnlyPath); err == nil && !has {
		if err = os.MkdirAll(readOnlyPath, 0755); err != nil {
			initContainerLog.WithFields(logrus.Fields{
				"err":     err,
				"errFrom": "createReadOnlyLayer",
			})
		}
		if _, err = exec.Command("tar", "-xvf", tarName, "-C", readOnlyPath).CombinedOutput(); err != nil {
			initContainerLog.WithFields(logrus.Fields{
				"err":     err,
				"errFrom": "createReadOnlyLayer",
			})
		}

	}

	//if has, err := DirOrFileExist(readOnlyPath); err == nil && has {
	//	//说明存在，先将其删除
	//	if err = os.RemoveAll(readOnlyPath); err != nil {
	//		initContainerLog.WithFields(logrus.Fields{
	//			"err":     err,
	//			"errFrom": "createMountLayer",
	//		})
	//	}
	//}
	//if err := os.Mkdir(readOnlyPath, 0755); err != nil {
	//	initContainerLog.WithFields(logrus.Fields{
	//		"err":     err,
	//		"errFrom": "createReadOnlyLayer",
	//	})
	//}
	////解压tar文件到只读层
	//if _, err := exec.Command("tar", "-xvf", tarName, "-C", readOnlyPath).CombinedOutput(); err != nil {
	//	initContainerLog.WithFields(logrus.Fields{
	//		"err":     err,
	//		"errFrom": "createReadOnlyLayer",
	//	})
	//}
	return readOnlyPath
}

//
//  createReadAndWriteLayer
//  @Description: 创建读写层
//  @param rootUrl
//  @return string
//
func createReadAndWriteLayer(rootUrl, containerID string) string {
	readAndWriteLayer := path.Join(rootUrl, "diff", containerID)

	if has, err := DirOrFileExist(readAndWriteLayer); err == nil && has {
		deleteReadAndWriteLayer(readAndWriteLayer)
	}
	if err := os.MkdirAll(readAndWriteLayer, 0755); err != nil {
		initContainerLog.WithFields(logrus.Fields{
			"err":     err,
			"errFrom": "createReadOnlyLayer",
		})
	}
	return readAndWriteLayer
}

//
//  createMountLayer
//  @Description: 创建挂载文件夹，并将读写层和只读层mount到该文件夹中
//  @param rootPath
//  @param readOnlyPath
//  @param readAndWriteLayer
//
func createMountLayer(rootPath, readOnlyPath, readAndWriteLayer, containerID string) string {
	//mntPath := rootPath + "mnt" + "/"
	mntPath := path.Join(rootPath, "mnt", containerID)
	//if has, err := DirOrFileExist(mntPath); err == nil && has {
	//	//说明存在，先将其删除
	//	if err = os.RemoveAll(mntPath); err != nil {
	//		initContainerLog.WithFields(logrus.Fields{
	//			"err":     err,
	//			"errFrom": "createMountLayer",
	//		})
	//	}
	//}
	if err := os.MkdirAll(mntPath, 0755); err != nil {
		initContainerLog.WithFields(logrus.Fields{
			"err":     err,
			"errFrom": "createReadOnlyLayer",
		})
	}
	dirs := "dirs=" + readAndWriteLayer + ":" + readOnlyPath
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "LYWHDockerMnt", mntPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		initContainerLog.WithFields(logrus.Fields{
			"err":     err,
			"errFrom": "createMountLayer",
		})
	}
	return mntPath
}

//
//  deleteWorkSpace
//  @Description: 推出容器时需要删除工作空间
//  @param workSpacePath，切片第一个元素表示mountlayer，第二个元素表示读写层元素
//
func deleteWorkSpace(workSpacePath []string) {
	deleteMountLayer(workSpacePath[0])
	deleteReadAndWriteLayer(workSpacePath[1])
}

func deleteMountLayer(mountLayer string) {
	cmd := exec.Command("umount", "-R", mountLayer) //注意，此处时循环卸载，所以数据卷挂载也能被卸载掉
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		initContainerLog.WithFields(logrus.Fields{
			"err":     err,
			"errFrom": "deleteMountLayer",
		})
	}
	if err := os.RemoveAll(mountLayer); err != nil {
		initContainerLog.WithFields(logrus.Fields{
			"err":     err,
			"errFrom": "deleteMountLayer",
		})
	}
}
func deleteReadAndWriteLayer(readAndWriteLayer string) {
	if err := os.RemoveAll(readAndWriteLayer); err != nil {
		initContainerLog.WithFields(logrus.Fields{
			"err":     err,
			"errFrom": "deleteReadAndWriteLayer",
		})
	}
}
