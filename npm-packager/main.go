package main

import (
	"fmt"
	"path"
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

func newPackageManager(pkg string, version string) (*PackageManager, error) {
	extracted := "./node_modules/"
	extractionPath := fmt.Sprintf("%s%s", extracted, pkg)
	data, err := downloadPackage(pkg, version, extractionPath)
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

func downloadPackage(pkg string, version string, extractedPath string) (*PackageJSON, error) {
	manifest := newDownloadManifest(pkg)
	if err := manifest.download(); err != nil {
		return nil, err
	}

	jsonParser := newParseJsonManifest("manifest/" + pkg + ".json")
	npmPackage, err := jsonParser.parse()
	if err != nil {
		return nil, err
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

	tarball := newDownloadTarball(tarballURL)
	if err := tarball.download(); err != nil {
		return nil, err
	}

	tarballFile := path.Join("tarball", path.Base(tarballURL))
	extractor := newTGZExtractor(tarballFile, extractedPath)
	if err := extractor.extract(); err != nil {
		return nil, err
	}

	packageJson := newPackageJSONParser(path.Join(extractedPath, "package.json"))
	data, err := packageJson.parse()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (pm *PackageManager) downloadDependencies() error {
	processed := make(map[string]bool)
	queue := make([]Dependency, 0)

	for name, version := range pm.dependencies {
		v := strings.Replace(version, "^", "", 1)
		queue = append(queue, Dependency{Name: name, Version: v})
	}

	for len(queue) > 0 {
		dep := queue[0]
		queue = queue[1:]

		depKey := fmt.Sprintf("%s@%s", dep.Name, dep.Version)
		if processed[depKey] {
			fmt.Printf("Skipping already processed: %s\n", depKey)
			continue
		}

		processed[depKey] = true
		pm.processedPackages = append(pm.processedPackages, dep)

		extractionPath := fmt.Sprintf("%s%s", pm.extractedPath, dep.Name)
		data, err := downloadPackage(dep.Name, dep.Version, extractionPath)
		if err != nil {
			return err
		}

		for depName, depVersion := range data.Dependencies {
			v := strings.Replace(depVersion, "^", "", 1)
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

	// packageManager, err := newPackageManager("express", "^5.0.0")
	packageManager, err := newPackageManager("fastify", "")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if err := packageManager.downloadDependencies(); err != nil {
		fmt.Println("Error downloading dependencies:", err)
		return
	}

}
