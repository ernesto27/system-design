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
	ResultChan chan<- JobResult
}

type JobResult struct {
	Dependency      Dependency
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

	// Concurrency fields
	processedMutex sync.Mutex
	processed      map[string]bool
	jobChan        chan Job
	resultChan     chan JobResult
	workerCount    int
	wg             sync.WaitGroup
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

	// Get package json dependencies
	packageJSON := newPackageJSONParser("package.json")
	data, err := packageJSON.parse()
	if err != nil {
		return nil, err
	}

	fmt.Println("Dependencies found in package.json:")
	for name, version := range data.Dependencies {
		fmt.Printf("  %s: %s\n", name, version)
	}

	return &PackageManager{
		dependencies:      data.Dependencies,
		extractedPath:     "./node_modules/",
		processedPackages: make(map[string]Dependency),
		configPath:        configPath,
		manifestPath:      manifestPath,
		tarballPath:       tarballPath,
		etagPath:          etagPath,

		// Initialize concurrency fields
		processed:   make(map[string]bool),
		jobChan:     make(chan Job, 100),
		resultChan:  make(chan JobResult, 100),
		workerCount: 5,
	}, nil
}

func downloadPackage(
	pkg string,
	version string,
	extractedPath string,
	manifestPath string,
	tarballPath string,
	etagPath string,
) (*PackageJSON, string, error) {
	manifest := newDownloadManifest(pkg, manifestPath, etagPath)
	etag, err := manifest.download()
	if err != nil {
		return nil, "", err
	}

	jsonParser := newParseJsonManifest(filepath.Join(manifestPath, pkg+".json"))
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

	tarball := newDownloadTarball(tarballURL, tarballPath)
	if err := tarball.download(); err != nil {
		return nil, "", err
	}

	tarballFile := filepath.Join(tarballPath, path.Base(tarballURL))
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

		data, etag, err := downloadPackage(dep.Name, dep.Version, extractionPath, pm.manifestPath, pm.tarballPath, pm.etagPath)
		dep.Etag = etag

		result := JobResult{
			Dependency: dep,
			Error:      err,
		}

		if err == nil && data != nil {
			result.NewDependencies = data.Dependencies
		}

		job.ResultChan <- result
	}
}

func (pm *PackageManager) downloadDependencies() error {
	if err := os.RemoveAll(pm.extractedPath); err != nil {
		return fmt.Errorf("failed to remove existing node_modules: %v", err)
	}

	for i := 0; i < pm.workerCount; i++ {
		pm.wg.Add(1)
		go pm.worker()
	}

	queue := make([]Dependency, 0)

	for name, version := range pm.dependencies {
		queue = append(queue, Dependency{Name: name, Version: version})
	}

	activeJobs := 0
	done := make(chan struct{})

	go func() {
		defer close(done)

		for len(queue) > 0 || activeJobs > 0 {
			if len(queue) > 0 && activeJobs < pm.workerCount {
				dep := queue[0]
				queue = queue[1:]

				depKey := dep.Name

				pm.processedMutex.Lock()
				if pm.processed[depKey] {
					pm.processedMutex.Unlock()
					fmt.Printf("Skipping already processed: %s\n", depKey)
					continue
				}
				pm.processed[depKey] = true
				pm.processedPackages[depKey] = dep
				pm.processedMutex.Unlock()

				job := Job{
					Dependency: dep,
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
					if !pm.processed[subDepKey] {
						fmt.Printf("  Found sub-dependency: %s: %s\n", depName, depVersion)
						queue = append(queue, Dependency{Name: depName, Version: depVersion})
					}
					pm.processedMutex.Unlock()
				}
			}
		}
	}()

	<-done

	close(pm.jobChan)
	pm.wg.Wait()

	return nil
}

func main() {
	startTime := time.Now()

	fmt.Println("All args:", os.Args)

	var param string
	if len(os.Args) > 1 {
		param = os.Args[1]
	}

	switch param {
	case "i":
		packageManager, err := newPackageManager()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if err := packageManager.downloadDependencies(); err != nil {
			fmt.Println("Error downloading dependencies:", err)
			return
		}

	default:
		os.Exit(1)
	}

	executionTime := time.Since(startTime)
	fmt.Printf("\nExecution completed in: %v\n", executionTime)
}
