package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ProjectVersion = "dev"
var ProjectBuild = "not versioned"

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show nomi version",
		Long:  "Show nomi version",
		Run:   versionRun,
	}
)

func versionRun(cmd *cobra.Command, args []string) {
	fmt.Printf("nomi version %s (build %s)\n", ProjectVersion, ProjectBuild)
}
