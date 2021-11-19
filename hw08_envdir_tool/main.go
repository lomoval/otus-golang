package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("incorrect number of parameters")
		os.Exit(-1)
	}
	envs, err := ReadDir(os.Args[1])
	if err != nil {
		fmt.Printf("failed to directory with environments %s", err)
		os.Exit(-1)
	}

	os.Exit(RunCmd(os.Args[2:], envs))
}
