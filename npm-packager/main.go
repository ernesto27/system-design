package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
)

const npmResgistryURL = "https://registry.npmjs.org/"

type DownloadTarball struct {
	url string
}

func newDownloadTarball(url string) *DownloadTarball {
	return &DownloadTarball{url: url}
}

func (d *DownloadTarball) download() error {
	filename := path.Base(d.url)
	tarballPath := path.Join("tarball", filename)
	return downloadFile(d.url, tarballPath)
}

type Dependency struct {
	Name    string
	Version string
}

type PackageManager struct {
	registryURL       string
	pkg               string
	dependencies      map[string]string
	extractedPath     string
	processedPackages []Dependency
}

func newPackageManager(pkg string) (*PackageManager, error) {
	manifest := newDownloadManifest(pkg)

	if err := manifest.download(); err != nil {
		return nil, err

	}

	jsonParser := newParseJsonManifest("manifest/" + pkg + ".json")
	npmPackage, err := jsonParser.parse()
	if err != nil {
		return nil, err
	}

	latestVersion := npmPackage.DistTags.Latest
	fmt.Println("Latest version:", latestVersion)

	var latestTarballURL string
	if _, exists := npmPackage.Versions[latestVersion]; exists {
		latestTarballURL = npmPackage.Versions[latestVersion].Dist.Tarball
		fmt.Println("version:", latestTarballURL)
	} else {
		fmt.Printf("Version %s not found in versions map\n", latestVersion)
	}

	tarball := newDownloadTarball(latestTarballURL)
	if err := tarball.download(); err != nil {
		return nil, err
	}

	extracted := "./node_modules/"
	extractionPath := fmt.Sprintf("%s%s", extracted, npmPackage.Name)
	tarballFile := path.Join("tarball", path.Base(latestTarballURL))
	extractor := newTGZExtractor(tarballFile, extractionPath)
	if err := extractor.extract(); err != nil {
		return nil, err
	}

	// Get packager json fron principal package
	packageJson := newPackageJSONParser(path.Join(extractionPath, "package.json"))
	data, err := packageJson.parse()
	if err != nil {
		return nil, err
	}

	return &PackageManager{
		registryURL:       "https://registry.npmjs.org/",
		pkg:               pkg,
		dependencies:      data.Dependencies,
		extractedPath:     extracted,
		processedPackages: make([]Dependency, 0),
	}, nil
}

func parseVersion(versionSpec string) string {
	versionRegex := regexp.MustCompile(`(\d+\.\d+\.\d+)`)
	matches := versionRegex.FindStringSubmatch(versionSpec)
	if len(matches) > 1 {
		return matches[1]
	}

	versionSpec = strings.ReplaceAll(versionSpec, "^", "")
	versionSpec = strings.ReplaceAll(versionSpec, "~", "")
	versionSpec = strings.ReplaceAll(versionSpec, ">=", "")
	versionSpec = strings.ReplaceAll(versionSpec, "<=", "")
	versionSpec = strings.ReplaceAll(versionSpec, ">", "")
	versionSpec = strings.ReplaceAll(versionSpec, "<", "")
	versionSpec = strings.TrimSpace(versionSpec)

	parts := strings.Fields(versionSpec)
	if len(parts) > 0 {
		return parts[0]
	}

	return versionSpec
}

func (pm *PackageManager) addDependency(name, version string) {
	pm.dependencies[name] = version
}

func (pm *PackageManager) saveDependenciesToJSON(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	dependencyData := struct {
		MainPackage  string       `json:"main_package"`
		TotalCount   int          `json:"total_dependencies"`
		Dependencies []Dependency `json:"dependencies"`
	}{
		MainPackage:  pm.pkg,
		TotalCount:   len(pm.processedPackages),
		Dependencies: pm.processedPackages,
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(dependencyData); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	fmt.Printf("Saved %d dependencies to %s\n", len(pm.processedPackages), filename)
	return nil
}

func (pm *PackageManager) downloadDependencies() error {
	processed := make(map[string]bool)
	queue := make([]Dependency, 0)

	for name, version := range pm.dependencies {
		v := parseVersion(version)
		queue = append(queue, Dependency{Name: name, Version: v})
	}

	for len(queue) > 0 {
		dep := queue[0]
		queue = queue[1:]

		if dep.Name == "wrappy" {
			// FIX THIS
			dep.Version = "1.0.2"
			fmt.Println("Debug stop")
		}

		if dep.Name == "safer-buffer" {
			fmt.Println("Debug stop")
		}

		depKey := fmt.Sprintf("%s@%s", dep.Name, dep.Version)
		if processed[depKey] {
			fmt.Printf("Skipping already processed: %s\n", depKey)
			continue
		}

		manifest := newDownloadManifest(dep.Name)

		if err := manifest.download(); err != nil {
			return err

		}

		fmt.Printf("Downloading dependency: %s: %s\n", dep.Name, dep.Version)
		processed[depKey] = true
		pm.processedPackages = append(pm.processedPackages, dep)

		tarballURL := fmt.Sprintf("%s%s/-/%s-%s.tgz", pm.registryURL, dep.Name, dep.Name, dep.Version)
		tarball := newDownloadTarball(tarballURL)
		if err := tarball.download(); err != nil {
			fmt.Printf("Error downloading %s: %v\n", dep.Name, err)
			return err
		}

		extractionPath := fmt.Sprintf("%s%s", pm.extractedPath, dep.Name)
		tarballFile := path.Join("tarball", path.Base(tarballURL))
		extractor := newTGZExtractor(tarballFile, extractionPath)
		if err := extractor.extract(); err != nil {
			fmt.Printf("Error extracting %s: %v\n", dep.Name, err)
			return err
		}

		packageJson := newPackageJSONParser(path.Join(extractionPath, "package.json"))
		data, err := packageJson.parse()
		if err != nil {
			fmt.Printf("Error parsing package.json for %s: %v\n", dep.Name, err)
			return err
		}

		for depName, depVersion := range data.Dependencies {
			if depName == "wrappy" {
				fmt.Println("Debug stop")
			}
			v := parseVersion(depVersion)
			subDepKey := fmt.Sprintf("%s@%s", depName, v)
			if !processed[subDepKey] {
				fmt.Printf("  Found sub-dependency: %s: %s (from %s)\n", depName, v, depVersion)
				queue = append(queue, Dependency{Name: depName, Version: v})
			}
		}
	}

	return nil
}

func main() {
	// if len(os.Args) < 2 {
	// 	fmt.Println("Usage: go run main.go <package-name>")
	// 	return
	// }

	// packageName := os.Args[1]

	packageManager, err := newPackageManager("express")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if err := packageManager.downloadDependencies(); err != nil {
		fmt.Println("Error downloading dependencies:", err)
		return
	}

	if err := packageManager.saveDependenciesToJSON("dependencies.json"); err != nil {
		fmt.Println("Error saving dependencies to JSON:", err)
		return
	}

}
