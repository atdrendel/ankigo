package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/atdrendel/ankigo/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		if errors.Is(err, cmd.ErrCancelled) || errors.Is(err, cmd.ErrSilent) {
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
