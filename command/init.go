package command

func init() {
	rootCommand.AddCommand(initCommand, runCommand)
	runCommand.Flags().BoolVarP(&tty, "tty", "t", false, "is use tty")
	runCommand.Flags().StringVarP(&resourceLimit.Memory, "memory", "m", "100m", "limit for memory")
	runCommand.Flags().StringVarP(&resourceLimit.CpuShare, "cpu-shares", "", "1024", "cpu time")
	runCommand.Flags().StringVarP(&resourceLimit.CpuSet, "cpu-set", "", "0", "cpu set")
	runCommand.Flags().StringVarP(&resourceLimit.CpuMems, "cpu-mems", "", "0", "cpu share")
	runCommand.Flags().StringVarP(&Volume, "value", "v", "", "volume")
}
