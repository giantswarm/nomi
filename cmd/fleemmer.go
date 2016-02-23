package cmd

import (
	"github.com/spf13/cobra"
)

var (
	FleemmerCmd = &cobra.Command{
		Use:   "fleemmer",
		Short: "Benchmarking tool that tests a fleet cluster",
		Long:  "Fleemmer is a benchmarking tool that tests a fleet cluster",
		Run:   fleemmerRun,
	}
)

func init() {
	FleemmerCmd.AddCommand(versionCmd)
	FleemmerCmd.AddCommand(runCmd)
}

func fleemmerRun(cmd *cobra.Command, args []string) {
	cmd.Help()
}
