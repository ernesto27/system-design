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

type VersionConflict struct {
	PackageName     string
	RequestedBy     map[string]string // parent package -> version requested
}

type PackageManager struct {
	dependencies      map[string]string
	extractedPath     string
	processedPackages map[string]Dependency
	configPath        string
	manifestPath      string
	tarballPath       string
	etagPath          string
	Etag              Etag
	isAdd             bool

	// Concurrency fields
	processedMutex sync.Mutex
	jobChan        chan Job
	resultChan     chan JobResult
	workerCount    int
	wg             sync.WaitGroup

	// Version conflict tracking
	versionRequests map[string]map[string]string // package -> (parent -> version)
	versionConflicts []VersionConflict
}

func newPackageManager() (*PackageManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %v", err)
	}

	configPath := filepath.Join(homeDir, ".config", "go-npm")
	if err := createDir(configPath); err != nil {
		return nil, err
	}

	manifestPath := filepath.Join(configPath, "manifest")
	if err := createDir(manifestPath); err != nil {
		return nil, err
	}

	tarballPath := filepath.Join(configPath, "tarball")
	if err := createDir(tarballPath); err != nil {
		return nil, err
	}

	etagPath := filepath.Join(configPath, "etag")
	if err := createDir(etagPath); err != nil {
		return nil, err
	}

	etag := newEtag(etagPath)

	return &PackageManager{
		dependencies:      make(map[string]string),
		extractedPath:     "./node_modules/",
		processedPackages: make(map[string]Dependency),
		configPath:        configPath,
		manifestPath:      manifestPath,
		tarballPath:       tarballPath,
		etagPath:          etagPath,
		Etag:              *etag,
		isAdd:             false,

		// Initialize concurrency fields
		jobChan:     make(chan Job, 100),
		resultChan:  make(chan JobResult, 100),
		workerCount: 5,

		// Initialize version tracking
		versionRequests: make(map[string]map[string]string),
		versionConflicts: make([]VersionConflict, 0),
	}, nil
}

func (pm *PackageManager) parsePackageJSON() error {
	// Get package json dependencies
	packageJSON := newPackageJSONParser("package.json")
	data, err := packageJSON.parse()
	if err != nil {
		return err
	}

	fmt.Println("Dependencies found in package.json:")
	for name, version := range data.Dependencies {
		fmt.Printf("  %s: %s\n", name, version)
	}

	pm.dependencies = data.Dependencies

	return nil
}

func (pm *PackageManager) setDependencies(pkg string, version string) {
	pm.isAdd = true
	pm.dependencies[pkg] = version
}

func (pm *PackageManager) downloadPackage(pkg string, version string, extractedPath string, etag string) (*PackageJSON, string, error) {
	manifest := newDownloadManifest(pkg, pm.manifestPath)
	etag, _, err := manifest.download(etag)
	if err != nil {
		return nil, "", err
	}

	jsonParser := newParseJsonManifest(filepath.Join(pm.manifestPath, pkg+".json"))
	npmPackage, err := jsonParser.parse()
	if err != nil {
		return nil, "", err
	}

	versionInfo := newVersionInfo(version, npmPackage)
	pkgVersion := versionInfo.getVersion()

	var tarballName string
	if strings.HasPrefix(pkg, "@") {
		parts := strings.Split(pkg, "/")
		if len(parts) == 2 {
			tarballName = parts[1]
		} else {
			tarballName = pkg
		}
	} else {
		tarballName = pkg
	}

	tarballURL := fmt.Sprintf("%s%s/-/%s-%s.tgz", npmResgistryURL, pkg, tarballName, pkgVersion)

	tarball := newDownloadTarball(tarballURL, pm.tarballPath)
	if err := tarball.download(); err != nil {
		return nil, "", err
	}

	tarballFile := filepath.Join(pm.tarballPath, path.Base(tarballURL))
	extractor := newTGZExtractor(tarballFile, extractedPath)
	if err := extractor.extract(); err != nil {
		return nil, "", err
	}

	packageJson := newPackageJSONParser(path.Join(extractedPath, "package.json"))
	data, err := packageJson.parse()
	if err != nil {
		return nil, "", err
	}

	return data, etag, nil
}

func (pm *PackageManager) worker() {
	defer pm.wg.Done()

	for job := range pm.jobChan {
		dep := job.Dependency
		extractionPath := fmt.Sprintf("%s%s", pm.extractedPath, dep.Name)

		etag := pm.Etag.get(dep.Name)

		data, etag, err := pm.downloadPackage(dep.Name, dep.Version, extractionPath, etag)
		dep.Etag = etag

		result := JobResult{
			Dependency: dep,
			ParentName: job.ParentName,
			Error:      err,
		}

		if err == nil && data != nil {
			result.NewDependencies = data.Dependencies
		}

		job.ResultChan <- result
	}
}

func (pm *PackageManager) downloadDependencies() error {
	if !pm.isAdd {
		if err := os.RemoveAll(pm.extractedPath); err != nil {
			return fmt.Errorf("failed to remove existing node_modules: %v", err)
		}
	}

	for i := 0; i < pm.workerCount; i++ {
		pm.wg.Add(1)
		go pm.worker()
	}

	type QueueItem struct {
		Dep        Dependency
		ParentName string
	}

	queue := make([]QueueItem, 0)

	for name, version := range pm.dependencies {
		queue = append(queue, QueueItem{
			Dep:        Dependency{Name: name, Version: version},
			ParentName: "package.json",
		})
	}

	activeJobs := 0
	done := make(chan struct{})

	go func() {
		defer close(done)

		for len(queue) > 0 || activeJobs > 0 {
			if len(queue) > 0 && activeJobs < pm.workerCount {
				item := queue[0]
				queue = queue[1:]

				dep := item.Dep
				depKey := dep.Name

				pm.processedMutex.Lock()

				// Track version request
				if pm.versionRequests[depKey] == nil {
					pm.versionRequests[depKey] = make(map[string]string)
				}
				pm.versionRequests[depKey][item.ParentName] = dep.Version

				if _, exists := pm.processedPackages[depKey]; exists {
					pm.processedMutex.Unlock()
					// fmt.Printf("Skipping already processed: %s\n", depKey)
					continue
				}
				pm.processedPackages[depKey] = dep
				pm.processedMutex.Unlock()

				job := Job{
					Dependency: dep,
					ParentName: item.ParentName,
					ResultChan: pm.resultChan,
				}

				pm.jobChan <- job
				activeJobs++
			} else if activeJobs > 0 {
				result := <-pm.resultChan
				activeJobs--

				if result.Error != nil {
					fmt.Printf("Error processing %s@%s: %v\n", result.Dependency.Name, result.Dependency.Version, result.Error)
					os.Exit(1)
				}

				_, ok := pm.processedPackages[result.Dependency.Name]
				if ok {
					pm.processedPackages[result.Dependency.Name] = result.Dependency
				}

				// Add new dependencies to queue
				for depName, depVersion := range result.NewDependencies {
					subDepKey := depName

					pm.processedMutex.Lock()
					if _, exists := pm.processedPackages[subDepKey]; !exists {
						// fmt.Printf("  Found sub-dependency: %s: %s\n", depName, depVersion)
						queue = append(queue, QueueItem{
							Dep:        Dependency{Name: depName, Version: depVersion},
							ParentName: result.Dependency.Name,
						})
					} else {
						// Track version request even if already processed
						if pm.versionRequests[subDepKey] == nil {
							pm.versionRequests[subDepKey] = make(map[string]string)
						}
						pm.versionRequests[subDepKey][result.Dependency.Name] = depVersion
					}
					pm.processedMutex.Unlock()
				}
			}
		}
	}()

	<-done

	close(pm.jobChan)
	pm.wg.Wait()

	// Detect and report version conflicts
	pm.detectVersionConflicts()

	pm.Etag.setPackages(pm.processedPackages)
	if err := pm.Etag.save(); err != nil {
		return fmt.Errorf("failed to save etag data: %v", err)
	}

	return nil
}

func (pm *PackageManager) detectVersionConflicts() {
	fmt.Println("\n=== Version Conflict Detection ===")

	hasConflicts := false

	for pkgName, requesters := range pm.versionRequests {
		if len(requesters) > 1 {
			// Check if all versions are the same
			versions := make(map[string]bool)
			for _, version := range requesters {
				versions[version] = true
			}

			// If more than one unique version is requested, it's a conflict
			if len(versions) > 1 {
				hasConflicts = true

				conflict := VersionConflict{
					PackageName: pkgName,
					RequestedBy: requesters,
				}
				pm.versionConflicts = append(pm.versionConflicts, conflict)

				fmt.Printf("\n⚠️  Package: %s\n", pkgName)
				fmt.Printf("   Requested by %d different packages with different versions:\n", len(requesters))

				for parent, version := range requesters {
					fmt.Printf("   - %s requires %s\n", parent, version)
				}

				// Show which version was actually installed (the first one encountered)
				if installedDep, ok := pm.processedPackages[pkgName]; ok {
					fmt.Printf("   ✓ Installed version: %s\n", installedDep.Version)
				}
			}
		}
	}

	if !hasConflicts {
		fmt.Println("No version conflicts detected!")
	} else {
		fmt.Printf("\n⚠️  Total conflicts detected: %d\n", len(pm.versionConflicts))
	}
	fmt.Println("==================================")
}

func main() {

	// etag, _ := downloadFile("https://registry.npmjs.org/express", "/tmp/express.json", "W/\"b8dd7dcd28522e9c7b03891e5602b80f\"")
	// fmt.Println("ETag:", etag)
	// return

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

	if err := packageManager.downloadDependencies(); err != nil {
		fmt.Println("Error downloading dependencies:", err)
		return
	}

	executionTime := time.Since(startTime)
	fmt.Printf("\nExecution completed in: %v\n", executionTime)
}
