package command

import (
	"LYWHDocker/cgroups/subsystems"
	"LYWHDocker/container"
	"LYWHDocker/log"
	"fmt"
	"github.com/spf13/cobra"
)

var (
	tty               = false
	resourceLimit     = &subsystems.ResourceConfig{}
	myCgroupsName     = "LYWHCGroups"
	Volume            = ""
	detach            = false
	containerName     = ""
	containerIDLenggh = 15
)

const (
	rootUse   = "root"
	initUse   = "init"
	runUse    = "run"
	commitUse = "commit"
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

func Execute() {
	if err := rootCommand.Execute(); err != nil {
		fmt.Println(err)
	}
}
