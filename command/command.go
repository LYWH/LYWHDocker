package command

import (
	"LYWHDocker/cgroups/subsystems"
	"LYWHDocker/container"
	"fmt"
	"github.com/spf13/cobra"
)

var (
	tty           = false
	resourceLimit = &subsystems.ResourceConfig{}
	myCgroupsName = "LYWHCGroups"
	Volume        = ""
)

const (
	rootUse = "root"
	initUse = "init"
	runUse  = "run"
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
		//log.Mylog.Info(runUse)
		//log.Mylog.Info(args,tty)
		////fmt.Printf("%T\n",args[0])
		container.RunContainer(tty, args[0], myCgroupsName, resourceLimit, Volume)
	},
}

func Execute() {
	if err := rootCommand.Execute(); err != nil {
		fmt.Println(err)
	}
}
