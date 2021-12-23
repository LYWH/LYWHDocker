package command

import (
	"LYWHDocker/cgroups/subsystems"
	"LYWHDocker/container"
	"LYWHDocker/log"
	"LYWHDocker/namespace"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var (
	tty               = false
	resourceLimit     = &subsystems.ResourceConfig{}
	myCgroupsName     = "LYWHCGroups"
	Volume            = ""
	detach            = false
	containerName     = ""
	containerIDLenggh = 15
	//inputContainerID  = ""
)

const (
	rootUse   = "root"
	initUse   = "init"
	runUse    = "run"
	commitUse = "commit"
	psUse     = "ps"
	logUse    = "log"
)

var rootCommand = &cobra.Command{
	Use:   rootUse,
	Short: "this is my Docker",
	Long:  "the docker is writed by LYWH",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var initCommand = &cobra.Command{
	Use:   initUse,
	Short: "use for init Container",
	RunE: func(cmd *cobra.Command, args []string) error {
		return container.InitNewNameSpace()
	},
}

var runCommand = &cobra.Command{
	Use:  runUse,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		//不允许终端和后台同时运行
		if tty && detach {
			log.Mylog.Error("tty and detach can't provide at the same time")
			return
		}
		//在此处生成容器ID
		containerID := container.GenerateContainerID(containerIDLenggh)
		container.RunContainer(tty, args[0], myCgroupsName, resourceLimit, Volume, containerName, containerID)
	},
}

var commitCommand = &cobra.Command{
	Use:   commitUse,
	Short: "commit the runing container",
	Long:  "commit the runing container",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		container.CommitContainer(args[0])
	},
}

var psCommand = &cobra.Command{
	Use:   psUse,
	Short: "ps displays information about a selection of the active processes.",
	Long:  "ps displays information about a selection of the active processes.",
	Run: func(cmd *cobra.Command, args []string) {
		container.OutputContainerInfo()
	},
}

var logCommand = &cobra.Command{
	Use:   logUse,
	Short: "output container log",
	Long:  "output container log",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		container.OutPutContainerLog(args[0])
	},
}

var execCommand = &cobra.Command{
	Use:   "exec [containerID] [containerCMD]",
	Short: "enter into existed container",
	Long:  "enter into existed container",
	Run: func(cmd *cobra.Command, args []string) {
		//整体思路：根据进程ID获取容器ID和CMD，然后使用系统调用setns进入namespace并指向相应的命令
		if len(os.Getenv(container.EXEC_ENV_PROCESS_ID)) != 0 { //此处是设置了环境变量后执行
			namespace.EnterNamespace()
			return
		}
		if len(args) < 2 { //参数不符合要求
			log.Mylog.Error("execCommand", "don't available args")
			return
		}
		cid, cmdstr := args[0], strings.Split(args[1], " ")
		container.EnterContainer(cid, cmdstr) //里面包含设置环境变量
	},
}

func Execute() {
	if err := rootCommand.Execute(); err != nil {
		fmt.Println(err)
	}
}
