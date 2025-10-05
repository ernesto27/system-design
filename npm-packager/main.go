package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const npmResgistryURL = "https://registry.npmjs.org/"

type Dependency struct {
	Name    string
	Version string
	Etag    string
	Nested  bool
}

type Job struct {
	Dependency Dependency
	ParentName string // Name of the parent package requesting this dependency
	ResultChan chan<- JobResult
}

type JobResult struct {
	Dependency      Dependency
	ParentName      string
	NewDependencies map[string]string
	Error           error
}

type PackageManager struct {
	dependencies      map[string]string
	extractedPath     string
	processedPackages map[string]Dependency
	configPath        string
	manifestPath      string
	tarballPath       string
	etagPath          string
	packagesPath      string
	Etag              Etag
	isAdd             bool
	packages          Packages
	packageLock       *PackageLock
	downloadManifest  *DownloadManifest
	downloadTarball   *DownloadTarball
	extractor         *TGZExtractor
	packageCopy       *PackageCopy
	parseJsonManifest *ParseJsonManifest
	versionInfo       *VersionInfo
	packageJsonParse  *PackageJSONParser

	processedMutex sync.Mutex
	packagesMutex  sync.Mutex
	jobChan        chan Job
	resultChan     chan JobResult
	workerCount    int
	wg             sync.WaitGroup
}

type Package struct {
	Version            string `json:"version"`
	Nested             bool
	Dependencies       []Dependency `json:"dependencies"`
	ParentDependencies []string
}

type Packages map[string]Package

func newPackageManager() (*PackageManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %v", err)
	}

	configPath := filepath.Join(homeDir, ".config", "go-npm")
	if err := createDir(configPath); err != nil {
		return nil, err
	}

	etagPath := filepath.Join(configPath, "etag")
	if err := createDir(etagPath); err != nil {
		return nil, err
	}

	packagePath := filepath.Join(configPath, "packages")
	if err := createDir(packagePath); err != nil {
		return nil, err
	}

	donwloadtarball, err := newDownloadTarball(configPath)
	if err != nil {
		return nil, err
	}

	downloadManifest, err := newDownloadManifest(configPath)
	if err != nil {
		return nil, err
	}

	extractor := newTGZExtractor()
	packageCopy := newPackageCopy("", "", make(Packages))
	etag := newEtag(etagPath)
	parseJsonManifest := newParseJsonManifest()
	versionInfo := newVersionInfo()

	return &PackageManager{
		dependencies:      make(map[string]string),
		extractedPath:     "./node_modules/",
		processedPackages: make(map[string]Dependency),
		configPath:        configPath,
		manifestPath:      downloadManifest.manifestPath,
		tarballPath:       donwloadtarball.tarballPath,
		etagPath:          etagPath,
		packagesPath:      packagePath,
		Etag:              *etag,
		isAdd:             false,
		packages:          make(Packages),
		downloadTarball:   donwloadtarball,
		extractor:         extractor,
		packageCopy:       packageCopy,
		downloadManifest:  downloadManifest,
		parseJsonManifest: parseJsonManifest,
		versionInfo:       versionInfo,

		// Initialize concurrency fields
		jobChan:     make(chan Job, 100),
		resultChan:  make(chan JobResult, 100),
		workerCount: 5,
	}, nil
}

type QueueItem struct {
	Dep        Dependency
	ParentName string
}

func (pm *PackageManager) parsePackageJSON() error {
	_, err := os.Stat("1package-lock.json")
	if err == nil {
		return pm.parsePackageJSONLock()

	}

	if _, err := os.Stat("package.json"); os.IsNotExist(err) {
		return fmt.Errorf("package.json not found in the current directory")
	}

	// // Get package json dependencies
	packageJSON := newPackageJSONParser()
	data, err := packageJSON.parse("package.json")
	if err != nil {
		return err
	}

	fmt.Println(data)

	queue := make([]QueueItem, 0)

	for name, version := range data.Dependencies {
		queue = append(queue, QueueItem{
			Dep:        Dependency{Name: name, Version: version},
			ParentName: "package.json",
		})
	}

	packageLock := PackageLock{}
	packageLock.Packages = make(map[string]PackageItem)
	packagesVersion := make(map[string]QueueItem)

	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]

		if item.Dep.Name == "" {
			continue
		}

		etag := pm.Etag.get(item.Dep.Name)
		currentEtag, _, err := pm.downloadManifest.download(item.Dep.Name, etag)
		if err != nil {
			return err
		}

		npmPackage, err := pm.parseJsonManifest.parse(filepath.Join(pm.manifestPath, item.Dep.Name+".json"))
		if err != nil {
			return err
		}

		var packageResolved string
		version := pm.versionInfo.getVersion(item.Dep.Version, npmPackage)
		if _, ok := packagesVersion[item.Dep.Name]; ok {
			if packagesVersion[item.Dep.Name].Dep.Version != version {
				fmt.Println("Package Repeated:", item.Dep.Name)
				fmt.Println("Resolved version:", version)
				packageResolved = "node_modules/" + item.ParentName + "/node_modules/" + item.Dep.Name
			} else {
				continue
			}
		} else {
			packageResolved = "node_modules/" + item.Dep.Name
		}

		configPackageVersion := filepath.Join(pm.packagesPath, item.Dep.Name+"@"+version)
		tarballURL := fmt.Sprintf("%s%s/-/%s-%s.tgz", npmResgistryURL, item.Dep.Name, item.Dep.Name, version)
		if !folderExists(configPackageVersion) {

			err = pm.downloadTarball.download(tarballURL)
			if err != nil {
				return err
			}

			err = pm.extractor.extract(
				filepath.Join(pm.tarballPath, path.Base(tarballURL)),
				configPackageVersion,
			)

			if err != nil {
				return err
			}
		}

		data, err := pm.packageJsonParse.parse(filepath.Join(pm.packagesPath, item.Dep.Name+"@"+version, "package.json"))
		if err != nil {
			return err
		}

		// Mark package as processed
		packagesVersion[item.Dep.Name] = QueueItem{
			Dep:        Dependency{Name: item.Dep.Name, Version: version},
			ParentName: item.ParentName,
		}
		pckItem := PackageItem{
			Version:  version,
			Resolved: tarballURL,
			Etag:     currentEtag,
		}
		packageLock.Packages[packageResolved] = pckItem

		for name, version := range data.Dependencies {
			queue = append(queue, QueueItem{
				Dep:        Dependency{Name: name, Version: version},
				ParentName: item.Dep.Name,
			})
		}
	}

	pm.packageLock = &packageLock

	// fmt.Println("Dependencies found in package.json:")
	// for name, version := range data.Dependencies {
	// 	fmt.Printf("  %s: %s\n", name, version)
	// }

	// pm.dependencies = data.Dependencies

	return nil
}

func (pm *PackageManager) setDependencies(pkg string, version string) {
	pm.isAdd = true
	pm.dependencies[pkg] = version
}

func (pm *PackageManager) parsePackageJSONLock() error {
	packageJSON := newPackageJSONParser()

	data, err := packageJSON.parseLockFile()
	if err != nil {
		return err
	}

	pm.packageLock = data
	return nil
}

func (pm *PackageManager) downloadFromPackageLock() error {
	// Remove node_modules
	err := os.RemoveAll(pm.extractedPath)
	if err != nil {
		return fmt.Errorf("failed to remove existing node_modules: %v", err)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(pm.packageLock.Packages))

	for name, item := range pm.packageLock.Packages {
		if name == "" {
			continue
		}

		wg.Add(1)
		go func(name string, item PackageItem) {
			defer wg.Done()

			namePkg := strings.TrimPrefix(name, "node_modules/")
			pkgName := namePkg
			if strings.Contains(namePkg, "/node_modules/") {
				parts := strings.Split(namePkg, "/node_modules/")
				pkgName = parts[len(parts)-1]
			}

			pathPkg := path.Join(pm.packagesPath, pkgName+"@"+item.Version)
			fmt.Println(pkgName, pathPkg)

			exists := folderExists(pathPkg)
			if !exists {
				// Download tarball
				err := pm.downloadTarball.download(item.Resolved)
				if err != nil {
					errChan <- err
					return
				}

				err = pm.extractor.extract(
					filepath.Join(pm.tarballPath, path.Base(item.Resolved)),
					pathPkg,
				)
				if err != nil {
					errChan <- err
					return
				}
			}

			// Preserve the nested structure by using the full path
			// e.g., node_modules/body-parser/node_modules/debug
			targetPath := path.Join(pm.extractedPath, namePkg)
			err := pm.packageCopy.copyDirectory(pathPkg, targetPath)
			if err != nil {
				errChan <- err
				return
			}
		}(name, item)
	}

	wg.Wait()
	close(errChan)

	// Return first error if any
	for err := range errChan {
		return err
	}

	return nil
}

func main() {

	startTime := time.Now()

	fmt.Println("All args:", os.Args)

	var param string
	if len(os.Args) > 1 {
		param = os.Args[1]
	}

	packageManager, err := newPackageManager()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	switch param {
	case "i":

		// If pakcage.json lock exists read that file
		// if version package exist in config,  copy to node_modules
		// if not download tarball and extract in cache and copy

		if err := packageManager.parsePackageJSON(); err != nil {
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

		packageManager.setDependencies(pkg, version)

	default:
		os.Exit(1)
	}

	// if err := packageManager.downloadDependencies(); err != nil {
	// 	fmt.Println("Error downloading dependencies:", err)
	// 	return
	// }

	if err := packageManager.downloadFromPackageLock(); err != nil {
		fmt.Println("Error downloading dependencies:", err)
		return
	}

	executionTime := time.Since(startTime)
	fmt.Printf("\nExecution completed in: %v\n", executionTime)
}
