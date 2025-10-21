package main

import (
	"fmt"
	"npm-packager/etag"
	"npm-packager/extractor"
	"npm-packager/manifest"
	"npm-packager/packagecopy"
	"npm-packager/packagejson"
	"npm-packager/tarball"
	"npm-packager/utils"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const npmResgistryURL = "https://registry.npmjs.org/"

type Job struct {
	Dependency packagejson.Dependency
	ParentName string // Name of the parent package requesting this dependency
	ResultChan chan<- JobResult
}

type JobResult struct {
	Dependency      packagejson.Dependency
	ParentName      string
	NewDependencies map[string]string
	Error           error
}

type PackageManager struct {
	dependencies      map[string]string
	extractedPath     string
	processedPackages map[string]packagejson.Dependency
	configPath        string
	packagesPath      string
	Etag              etag.Etag
	isAdd             bool
	packages          Packages
	packageLock       *packagejson.PackageLock
	manifest          *manifest.Manifest
	tarball           *tarball.Tarball
	extractor         *extractor.TGZExtractor
	packageCopy       *packagecopy.PackageCopy
	parseJsonManifest *ParseJsonManifest
	versionInfo       *VersionInfo
	packageJsonParse  *packagejson.PackageJSONParser
	downloadMu        sync.Mutex
	downloadLocks     map[string]*sync.Mutex
}

type Package struct {
	Version            string `json:"version"`
	Nested             bool
	Dependencies       []packagejson.Dependency `json:"dependencies"`
	ParentDependencies []string
}

type Packages map[string]Package

func newPackageManager() (*PackageManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %v", err)
	}

	configPath := filepath.Join(homeDir, ".config", "go-npm")
	if err := utils.CreateDir(configPath); err != nil {
		return nil, err
	}

	packagePath := filepath.Join(configPath, "packages")
	if err := utils.CreateDir(packagePath); err != nil {
		return nil, err
	}

	manifest, err := manifest.NewManifest(configPath, npmResgistryURL)
	if err != nil {
		return nil, err
	}
	etag, err := etag.NewEtag(configPath)
	if err != nil {
		return nil, err
	}

	donwloadtarball := tarball.NewTarball()
	extractor := extractor.NewTGZExtractor()
	packageCopy := packagecopy.NewPackageCopy()
	parseJsonManifest := newParseJsonManifest()
	versionInfo := newVersionInfo()
	packageJsonParse := packagejson.NewPackageJSONParser()

	return &PackageManager{
		dependencies:      make(map[string]string),
		extractedPath:     "./node_modules/",
		processedPackages: make(map[string]packagejson.Dependency),
		configPath:        configPath,
		packagesPath:      packagePath,
		Etag:              *etag,
		isAdd:             false,
		packages:          make(Packages),
		tarball:           donwloadtarball,
		extractor:         extractor,
		packageCopy:       packageCopy,
		manifest:          manifest,
		parseJsonManifest: parseJsonManifest,
		versionInfo:       versionInfo,
		packageJsonParse:  packageJsonParse,
		downloadLocks:     make(map[string]*sync.Mutex),
	}, nil
}

type QueueItem struct {
	Dep        packagejson.Dependency
	ParentName string
}

func (pm *PackageManager) parsePackageJSON() error {
	data, err := pm.packageJsonParse.ParseDefault()
	if err != nil {
		return err
	}

	if pm.packageJsonParse.PackageLock != nil {
		packagesToAdd := pm.packageJsonParse.ResolveDependencies()

		for _, pkg := range packagesToAdd {
			err = pm.add(pkg.Name, pkg.Version, true)
			if err != nil {
				return err
			}
		}

		pm.packageLock = pm.packageJsonParse.PackageLock
		return nil
	}

	err = pm.download(*data)
	if err != nil {
		return err
	}

	err = pm.packageJsonParse.CreateLockFile(pm.packageLock)
	if err != nil {
		return err
	}

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
		go func(name string, item packagejson.PackageItem) {
			defer wg.Done()

			namePkg := strings.TrimPrefix(name, "node_modules/")
			pkgName := namePkg
			if strings.Contains(namePkg, "/node_modules/") {
				parts := strings.Split(namePkg, "/node_modules/")
				pkgName = parts[len(parts)-1]
			}

			pathPkg := path.Join(pm.packagesPath, pkgName+"@"+item.Version)

			exists := utils.FolderExists(pathPkg)
			if !exists {
				// Download tarball
				err := pm.tarball.Download(item.Resolved)
				if err != nil {
					errChan <- err
					return
				}

				err = pm.extractor.Extract(
					filepath.Join(pm.tarball.TarballPath, path.Base(item.Resolved)),
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
			err := pm.packageCopy.CopyDirectory(pathPkg, targetPath)
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

func (pm *PackageManager) removePackagesFromNodeModules(pkgList []string) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(pkgList))

	for _, pkg := range pkgList {
		wg.Add(1)
		go func(pkgName string) {
			defer wg.Done()

			pkgPath := filepath.Join(pm.extractedPath, pkgName)

			if err := os.RemoveAll(pkgPath); err != nil {
				errChan <- fmt.Errorf("failed to remove package %s: %w", pkgName, err)
			}
		}(pkg)
	}

	wg.Wait()
	close(errChan)

	// Return first error if any
	for err := range errChan {
		return err
	}

	return nil
}

func (pm *PackageManager) add(pkgName string, version string, isInstall bool) error {
	packageJson, err := pm.packageJsonParse.ParseDefault()
	if err != nil {
		return err
	}

	if !isInstall {
		// Check if pkgName exists
		if _, exists := packageJson.Dependencies[pkgName]; exists {
			if version != "" && packageJson.Dependencies[pkgName] == version {
				fmt.Println("Package", pkgName, "already exists in dependencies with the same version", version)
				return nil
			}
		}
	}

	// Download package and its dependencies
	packageJsonAdd := packagejson.PackageJSON{
		Dependencies: map[string]string{
			pkgName: version,
		},
	}
	err = pm.download(packageJsonAdd)
	if err != nil {
		return err
	}

	// New package or update add in package.json
	err = pm.packageJsonParse.AddOrUpdateDependency(pkgName, version)
	if err != nil {
		return err
	}

	// update package.json lock file
	err = pm.packageJsonParse.UpdateLockFile(pm.packageLock)
	if err != nil {
		return err
	}

	pm.packageLock = pm.packageJsonParse.PackageLock

	return nil
}

func (pm *PackageManager) remove(pkg string) error {
	packageJson, err := pm.packageJsonParse.ParseDefault()
	if err != nil {
		return err
	}
	fmt.Println(packageJson)

	pkgToRemove := pm.packageJsonParse.ResolveDependenciesToRemove(pkg)
	fmt.Println(pkgToRemove)

	err = pm.removePackagesFromNodeModules(pkgToRemove)
	if err != nil {
		return err
	}

	// err = pm.packageJsonParse.RemoveDependencies(pkg)
	// if err != nil {
	// 	return err
	// }

	return nil
}

func (pm *PackageManager) download(packageJson packagejson.PackageJSON) error {
	queue := make([]QueueItem, 0)

	for name, version := range packageJson.Dependencies {
		queue = append(queue, QueueItem{
			Dep:        packagejson.Dependency{Name: name, Version: version},
			ParentName: "package.json",
		})
	}

	packageLock := packagejson.PackageLock{}
	packageLock.Packages = make(map[string]packagejson.PackageItem)
	packageLock.Dependencies = make(map[string]string)
	packagesVersion := make(map[string]QueueItem)

	var (
		wg             sync.WaitGroup
		mapMutex       sync.Mutex
		activeWorkers  int
		workerMutex    sync.Mutex
		processingPkgs = make(map[string]bool)
	)

	errChan := make(chan error, 1)
	done := make(chan struct{})

	workChan := make(chan QueueItem, len(queue))
	for _, item := range queue {
		packageLock.Dependencies[item.Dep.Name] = item.Dep.Version
		workChan <- item
	}

	for {
		workerMutex.Lock()
		workers := activeWorkers
		workerMutex.Unlock()

		if len(workChan) == 0 && workers == 0 {
			break
		}

		select {
		case item := <-workChan:
			workerMutex.Lock()
			activeWorkers++
			workerMutex.Unlock()

			wg.Add(1)

			go func(item QueueItem) {
				defer func() {
					wg.Done()
					workerMutex.Lock()
					activeWorkers--
					workerMutex.Unlock()
				}()

				if item.Dep.Name == "" {
					return
				}

				select {
				case <-done:
					return
				default:
				}

				// Get or create a lock for this package's manifest
				pm.downloadMu.Lock()
				pkgLock, exists := pm.downloadLocks[item.Dep.Name]
				if !exists {
					pkgLock = &sync.Mutex{}
					pm.downloadLocks[item.Dep.Name] = pkgLock
				}
				pm.downloadMu.Unlock()

				// Lock only for manifest download and parse (not for tarball downloads)
				pkgLock.Lock()

				manifestPath := filepath.Join(pm.manifest.Path, item.Dep.Name+".json")
				var currentEtag string

				// Check if manifest already exists (inside lock to avoid race)
				if _, err := os.Stat(manifestPath); err == nil {
					// Manifest already downloaded by another goroutine, skip download
					currentEtag = pm.Etag.Get(item.Dep.Name)
				} else {
					// First goroutine to process this package, download the manifest
					etag := pm.Etag.Get(item.Dep.Name)
					var downloadErr error
					currentEtag, _, downloadErr = pm.manifest.Download(item.Dep.Name, etag)
					if downloadErr != nil {
						pkgLock.Unlock()
						select {
						case errChan <- downloadErr:
							close(done)
						default:
						}
						return
					}
				}

				npmPackage, err := pm.parseJsonManifest.parse(manifestPath)
				pkgLock.Unlock()
				// Unlock immediately after parsing - different versions can now download tarballs in parallel

				if err != nil {
					select {
					case errChan <- err:
						close(done)
					default:
					}
					return
				}

				version := pm.versionInfo.getVersion(item.Dep.Version, npmPackage)
				packageKey := item.Dep.Name + "@" + version

				if version == "" {
					fmt.Println("Version not found for package:", item.Dep.Name, "with constraint:", item.Dep.Version)
				}

				var packageResolved string
				var shouldProcessDeps bool

				mapMutex.Lock()
				if existingPkg, ok := packagesVersion[item.Dep.Name]; ok {
					if existingPkg.Dep.Version != version {
						fmt.Println("Package Repeated:", item.Dep.Name)
						fmt.Println("Resolved version:", version)
						packageResolved = "node_modules/" + item.ParentName + "/node_modules/" + item.Dep.Name

						if _, processed := processingPkgs[packageKey]; processed {
							shouldProcessDeps = false
						} else {
							processingPkgs[packageKey] = true
							shouldProcessDeps = true
						}
					} else {
						// Same version already resolved at top level, skip
						mapMutex.Unlock()
						return
					}
				} else {
					// First time seeing this package name - install at top level
					packageResolved = "node_modules/" + item.Dep.Name
					packagesVersion[item.Dep.Name] = QueueItem{
						Dep:        packagejson.Dependency{Name: item.Dep.Name, Version: version},
						ParentName: item.ParentName,
					}

					if _, processed := processingPkgs[packageKey]; processed {
						shouldProcessDeps = false
					} else {
						processingPkgs[packageKey] = true
						shouldProcessDeps = true
					}
				}
				mapMutex.Unlock()

				configPackageVersion := filepath.Join(pm.packagesPath, item.Dep.Name+"@"+version)

				// Extract tarball name for scoped packages (@scope/package -> package)
				tarballName := item.Dep.Name
				if strings.HasPrefix(item.Dep.Name, "@") && strings.Contains(item.Dep.Name, "/") {
					parts := strings.Split(item.Dep.Name, "/")
					tarballName = parts[1]
				}

				tarballURL := fmt.Sprintf("%s%s/-/%s-%s.tgz", npmResgistryURL, item.Dep.Name, tarballName, version)

				// Download and extract only if we haven't processed this exact version yet
				if shouldProcessDeps && !utils.FolderExists(configPackageVersion) {
					err = pm.tarball.Download(tarballURL)
					if err != nil {
						select {
						case errChan <- err:
							close(done)
						default:
						}
						return
					}

					err = pm.extractor.Extract(
						filepath.Join(pm.tarball.TarballPath, path.Base(tarballURL)),
						configPackageVersion,
					)

					if err != nil {
						select {
						case errChan <- err:
							close(done)
						default:
						}
						return
					}
				}

				// Add to package lock
				mapMutex.Lock()
				pckItem := packagejson.PackageItem{
					Name:     item.Dep.Name,
					Version:  version,
					Resolved: tarballURL,
					Etag:     currentEtag,
				}
				packageLock.Packages[packageResolved] = pckItem
				mapMutex.Unlock()

				// Only process dependencies if this is the first time we see this version
				if shouldProcessDeps {
					packageJsonPath := filepath.Join(pm.packagesPath, item.Dep.Name+"@"+version, "package.json")

					data, err := pm.packageJsonParse.Parse(packageJsonPath)
					if err != nil {
						select {
						case errChan <- err:
							close(done)
						default:
						}
						return
					}

					mapMutex.Lock()
					for name, version := range data.Dependencies {
						pkgItem := packageLock.Packages[packageResolved]
						if pkgItem.Dependencies == nil {
							pkgItem.Dependencies = make(map[string]string)
						}
						pkgItem.Dependencies[name] = version
						packageLock.Packages[packageResolved] = pkgItem

						workChan <- QueueItem{
							Dep:        packagejson.Dependency{Name: name, Version: version},
							ParentName: item.Dep.Name,
						}
					}
					mapMutex.Unlock()
				}
			}(item)
		default:
			// No work available, check if we're done
			workerMutex.Lock()
			if activeWorkers == 0 {
				workerMutex.Unlock()
				break
			}
			workerMutex.Unlock()
		}
	}

	wg.Wait()
	close(errChan)

	if err := <-errChan; err != nil {
		return err
	}
	pm.packageLock = &packageLock

	return nil
}

func main() {

	startTime := time.Now()

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

		// packageManager.setDependencies(pkg, version)
		err = packageManager.add(pkg, version, false)
		if err != nil {
			fmt.Println("Error adding package:", err)
			return
		}
	case "rm":
		err := packageManager.remove(os.Args[2])
		if err != nil {
			panic(err)
		}
		return

	default:
		os.Exit(1)
	}

	if err := packageManager.downloadFromPackageLock(); err != nil {
		fmt.Println(err)
		return
	}

	executionTime := time.Since(startTime)
	fmt.Printf("\nExecution completed in: %v\n", executionTime)
}
