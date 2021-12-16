package command

func init() {
	rootCommand.AddCommand(initCommand, runCommand)
	runCommand.Flags().BoolVarP(&tty, "tty", "t", false, "is use tty")
	runCommand.Flags().StringVarP(&resourceLimit.Memory, "memory", "m", "0m", "limit for memory")
	runCommand.Flags().StringVarP(&resourceLimit.CpuShare, "cpu-shares", "", "1024", "cpu time")
}
