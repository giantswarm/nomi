package main

import (
	"fmt"
	"os"

	"github.com/giantswarm/fleemmer/cmd"
)

func main() {
	if err := cmd.FleemmerCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
