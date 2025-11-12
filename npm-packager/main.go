package main

import (
	"fmt"
	"npm-packager/manager"
	"os"
	"strings"
	"time"
)

func main() {
	startTime := time.Now()

	var param string
	if len(os.Args) > 1 {
		param = os.Args[1]
	}

	deps, err := manager.BuildDependencies()
	if err != nil {
		fmt.Println("Error building dependencies:", err)
		return
	}

	packageManager, err := manager.New(deps)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	switch param {
	case "i":
		if err := packageManager.ParsePackageJSON(); err != nil {
			fmt.Println("Error parsing package.json:", err)
			return
		}

	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go-npm add <package-name>@<version>")
			os.Exit(1)
		}
		pkgArg := os.Args[2]
		parts := strings.Split(pkgArg, "@")

		pkg := parts[0]
		version := ""
		if len(parts) > 1 {
			version = parts[1]
		}
		fmt.Println("pkg:", pkg)
		fmt.Println("version:", version)

		err = packageManager.Add(pkg, version, false)
		if err != nil {
			fmt.Println("Error adding package:", err)
			return
		}

	case "rm":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go-npm rm <package-name>")
			os.Exit(1)
		}
		err := packageManager.Remove(os.Args[2], true)
		if err != nil {
			fmt.Println("Error removing package:", err)
			return
		}
		fmt.Println("Package removed successfully")
		return
	case "g":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go-npm g <package-name>[@version]")
			os.Exit(1)
		}

		pkgArg := os.Args[2]
		parts := strings.Split(pkgArg, "@")

		pkg := parts[0]
		version := ""
		if len(parts) > 1 {
			version = parts[1]
		}

		err := packageManager.SetupGlobal()
		if err != nil {
			fmt.Println("Error setting up global installation:", err)
			return
		}

		err = packageManager.InstallGlobal(pkg, version)
		if err != nil {
			fmt.Println("Error installing globally:", err)
			return
		}

		executionTime := time.Since(startTime)
		fmt.Printf("\nExecution completed in: %v\n", executionTime)
		return

	default:
		fmt.Println("Usage: go-npm [i|add|rm|g] [package-name]")
		os.Exit(1)
	}

	if err := packageManager.InstallFromCache(); err != nil {
		fmt.Println(err)
		return
	}

	executionTime := time.Since(startTime)
	fmt.Printf("\nExecution completed in: %v\n", executionTime)
}
