package main

import (
	"aexp_assesment/cli"
	"fmt"
	"os"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
