package cmd

import (
	"github.com/spf13/cobra"
)

var (
	NomiCmd = &cobra.Command{
		Use:   "nomi",
		Short: "Benchmarking tool that tests a fleet cluster",
		Long:  "Nomi is a benchmarking tool that tests a fleet cluster",
		Run:   nomiRun,
	}
)

func init() {
	NomiCmd.AddCommand(versionCmd)
	NomiCmd.AddCommand(runCmd)
}

func nomiRun(cmd *cobra.Command, args []string) {
	cmd.Help()
}
