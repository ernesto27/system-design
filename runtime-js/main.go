package main

import (
	"fmt"
	"os"
	"runtimejs/runtimejs"
)

func main() {
	args := os.Args

	if len(args) < 2 {
		fmt.Println("Please provide a filename as an argument")
		os.Exit(1)
	}

	filename := args[1]

	runtimeJS, err := runtimejs.NewRuntimeJS(filename)
	if err != nil {
		fmt.Println("Error creating runtimeJS:", err)
		os.Exit(1)
	}
	runtimeJS.RunEventLoop()
}
