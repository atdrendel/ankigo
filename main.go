package main

import (
	"os"

	"github.com/atdrendel/ankigo/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
