package command

func init() {
	rootCommand.AddCommand(initCommand, runCommand, commitCommand, psCommand, logCommand, execCommand, stopCommand, removeCommand, networkCommand)
	networkCommand.AddCommand(networkCreateCommand, networkListCommand, networkDeleteCommand)

	runCommand.Flags().BoolVarP(&tty, "tty", "t", false, "is use tty")
	runCommand.Flags().StringVarP(&resourceLimit.Memory, "memory", "m", "100m", "limit for memory")
	runCommand.Flags().StringVarP(&resourceLimit.CpuShare, "cpu-shares", "", "1024", "cpu time")
	runCommand.Flags().StringVarP(&resourceLimit.CpuSet, "cpu-set", "", "0", "cpu set")
	runCommand.Flags().StringVarP(&resourceLimit.CpuMems, "cpu-mems", "", "0", "cpu share")
	runCommand.Flags().StringVarP(&Volume, "value", "v", "", "volume")
	runCommand.Flags().BoolVarP(&detach, "detach", "d", false, "making the container process detach")
	runCommand.Flags().StringVarP(&containerName, "name", "n", "", "the container name")
	runCommand.Flags().StringVarP(&imageTarPath, "imageTarPath", "i", "./busybox.tar", "the image tar file path of the container")
	runCommand.Flags().StringSliceVarP(&envVar, "set-environment", "e", []string{}, "set environment")
	runCommand.Flags().StringVarP(&networkName, "net", "", "", "the network name")
	runCommand.Flags().StringSliceVarP(&port, "port", "p", []string{}, "port mapping")

	networkCreateCommand.Flags().StringVarP(&driver, "driver", "", "bridge", "network driver")
	networkCreateCommand.Flags().StringVarP(&subnet, "subnet", "", "", "subnet address")
	networkCreateCommand.MarkFlagRequired(driver)
	networkCreateCommand.MarkFlagRequired(subnet)
}
