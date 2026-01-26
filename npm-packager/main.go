package main

import (
	"flag"
	"fmt"
	"npm-packager/manager"
	"os"
	"strings"
	"time"
)

func parsePackageArg(pkgArg string) (string, string) {
	parts := strings.Split(pkgArg, "@")
	pkg := parts[0]
	version := ""
	if len(parts) > 1 {
		version = parts[1]
	}
	return pkg, version
}

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
		iFlags := flag.NewFlagSet("i", flag.ExitOnError)
		globalFlag := iFlags.Bool("g", false, "Install package globally")
		productionFlag := iFlags.Bool("production", false, "Install only production dependencies")

		iFlags.Parse(os.Args[2:])
		args := iFlags.Args()

		if *globalFlag {
			if len(args) < 1 {
				fmt.Println("Usage: go-npm i -g <package-name>[@version]")
				os.Exit(1)
			}

			pkg, version := parsePackageArg(args[0])

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
		}

		if err := packageManager.ParsePackageJSON(*productionFlag); err != nil {
			fmt.Println("Error parsing package.json:", err)
			return
		}

	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go-npm add <package-name>@<version>")
			os.Exit(1)
		}
		pkg, version := parsePackageArg(os.Args[2])
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

	case "uninstall":
		uninstallFlags := flag.NewFlagSet("uninstall", flag.ExitOnError)
		globalFlag := uninstallFlags.Bool("g", false, "Uninstall package globally")
		uninstallFlags.Parse(os.Args[2:])

		if *globalFlag {
			args := uninstallFlags.Args()
			if len(args) < 1 {
				fmt.Println("Usage: go-npm uninstall -g <package-name>")
				os.Exit(1)
			}

			err := packageManager.SetupGlobal()
			if err != nil {
				fmt.Println("Error setting up global installation:", err)
				return
			}

			err = packageManager.Remove(args[0], false)
			if err != nil {
				fmt.Println("Error removing package:", err)
				return
			}
			fmt.Println("Package removed successfully")
			return

		}

		args := uninstallFlags.Args()
		if len(args) < 1 {
			fmt.Println("Usage: go-npm uninstall <package-name>")
			os.Exit(1)
		}
		err := packageManager.Remove(args[0], true)
		if err != nil {
			fmt.Println("Error removing package:", err)
			return
		}
		fmt.Println("Package removed successfully")
		return

	default:
		fmt.Println("Usage: go-npm [i|add|rm|uninstall] [package-name]")
		os.Exit(1)
	}

	if err := packageManager.InstallFromCache(); err != nil {
		fmt.Println(err)
		return
	}

	executionTime := time.Since(startTime)
	fmt.Printf("\nExecution completed in: %v\n", executionTime)
}
