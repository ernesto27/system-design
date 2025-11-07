package manager

import (
	"fmt"
	"npm-packager/binlink"
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
)

const npmRegistryURL = "https://registry.npmjs.org/"

type Job struct {
	Dependency packagejson.Dependency
	ParentName string
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
	binLinker         *binlink.BinLinker
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

type QueueItem struct {
	Dep        packagejson.Dependency
	ParentName string
}

func New() (*PackageManager, error) {
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

	manifest, err := manifest.NewManifest(configPath, npmRegistryURL)
	if err != nil {
		return nil, err
	}
	etag, err := etag.NewEtag(configPath)
	if err != nil {
		return nil, err
	}

	downloadTarball := tarball.NewTarball()
	extractor := extractor.NewTGZExtractor()
	packageCopy := packagecopy.NewPackageCopy()
	parseJsonManifest := newParseJsonManifest()
	versionInfo := newVersionInfo()
	packageJsonParse := packagejson.NewPackageJSONParser()
	binLinker := binlink.NewBinLinker("./node_modules/")

	return &PackageManager{
		dependencies:      make(map[string]string),
		extractedPath:     "./node_modules/",
		processedPackages: make(map[string]packagejson.Dependency),
		configPath:        configPath,
		packagesPath:      packagePath,
		Etag:              *etag,
		isAdd:             false,
		packages:          make(Packages),
		tarball:           downloadTarball,
		extractor:         extractor,
		packageCopy:       packageCopy,
		manifest:          manifest,
		parseJsonManifest: parseJsonManifest,
		versionInfo:       versionInfo,
		packageJsonParse:  packageJsonParse,
		binLinker:         binLinker,
		downloadLocks:     make(map[string]*sync.Mutex),
	}, nil
}

func (pm *PackageManager) ParsePackageJSON() error {
	data, err := pm.packageJsonParse.ParseDefault()
	if err != nil {
		return err
	}

	if pm.packageJsonParse.PackageLock != nil {
		packagesToAdd, packagesToRemove := pm.packageJsonParse.ResolveDependencies()

		for _, pkg := range packagesToAdd {
			err = pm.Add(pkg.Name, pkg.Version, true)
			if err != nil {
				return err
			}
		}

		for _, pkg := range packagesToRemove {
			err = pm.Remove(pkg.Name, false)
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

func (pm *PackageManager) DownloadFromPackageLock() error {
	packagesToInstall := make(map[string]packagejson.PackageItem)
	for pkgPath := range pm.packageLock.Packages {
		exists := utils.FolderExists(pkgPath)
		if !exists {
			packagesToInstall[pkgPath] = pm.packageLock.Packages[pkgPath]
		}
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(packagesToInstall))
	for name, item := range packagesToInstall {
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
				if item.Resolved == "" {
					fmt.Printf("Skipping package %s - empty resolved URL in lock file\n", item.Name)
					return
				}
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

	for err := range errChan {
		return err
	}

	// Link bin executables
	if err := pm.binLinker.LinkAllPackages(); err != nil {
		return fmt.Errorf("failed to link bin executables: %w", err)
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

	for err := range errChan {
		return err
	}

	return nil
}

func (pm *PackageManager) Add(pkgName string, version string, isInstall bool) error {
	packageJson, err := pm.packageJsonParse.ParseDefault()
	if err != nil {
		return err
	}

	if !isInstall {
		if _, exists := packageJson.Dependencies[pkgName]; exists {
			if version != "" && packageJson.Dependencies[pkgName] == version {
				fmt.Println("Package", pkgName, "already exists in dependencies with the same version", version)
				return nil
			}
		}
	}

	packageJsonAdd := packagejson.PackageJSON{
		Dependencies: map[string]string{
			pkgName: version,
		},
	}
	err = pm.download(packageJsonAdd)
	if err != nil {
		return err
	}

	err = pm.packageJsonParse.AddOrUpdateDependency(pkgName, version)
	if err != nil {
		return err
	}

	err = pm.packageJsonParse.UpdateLockFile(pm.packageLock)
	if err != nil {
		return err
	}

	pm.packageLock = pm.packageJsonParse.PackageLock

	return nil
}

func (pm *PackageManager) Remove(pkg string, removeFromPackageJson bool) error {
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

	if removeFromPackageJson {
		err = pm.packageJsonParse.RemoveDependencies(pkg)
		if err != nil {
			return err
		}
	}

	err = pm.packageJsonParse.RemoveFromLockFile(pkg, pkgToRemove)
	if err != nil {
		return err
	}
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

	for name, version := range packageJson.DevDependencies {
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

				pm.downloadMu.Lock()
				pkgLock, exists := pm.downloadLocks[item.Dep.Name]
				if !exists {
					pkgLock = &sync.Mutex{}
					pm.downloadLocks[item.Dep.Name] = pkgLock
				}
				pm.downloadMu.Unlock()

				pkgLock.Lock()

				manifestPath := filepath.Join(pm.manifest.Path, item.Dep.Name+".json")
				var currentEtag string

				if _, err := os.Stat(manifestPath); err == nil {
					currentEtag = pm.Etag.Get(item.Dep.Name)
				} else {
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
						mapMutex.Unlock()
						return
					}
				} else {
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

				tarballName := item.Dep.Name
				if strings.HasPrefix(item.Dep.Name, "@") && strings.Contains(item.Dep.Name, "/") {
					parts := strings.Split(item.Dep.Name, "/")
					tarballName = parts[1]
				}

				tarballURL := fmt.Sprintf("%s%s/-/%s-%s.tgz", npmRegistryURL, item.Dep.Name, tarballName, version)

				if shouldProcessDeps && !utils.FolderExists(configPackageVersion) {
					if tarballURL == "" || version == "" {
						fmt.Printf("Skipping download for %s - invalid URL or empty version\n", item.Dep.Name)
						return
					}
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

				mapMutex.Lock()
				pckItem := packagejson.PackageItem{
					Name:     item.Dep.Name,
					Version:  version,
					Resolved: tarballURL,
					Etag:     currentEtag,
				}
				packageLock.Packages[packageResolved] = pckItem
				mapMutex.Unlock()

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
