package main

import (
	"fmt"
	"os"

	"github.com/giantswarm/nomi/cmd"
)

func main() {
	if err := cmd.NomiCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
