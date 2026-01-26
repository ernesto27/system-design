package packagejson

import (
	"encoding/json"
	"fmt"
	"npm-packager/config"
	"os"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type Dependency struct {
	Name    string
	Version string
	Etag    string
	Nested  bool
}

type PackageJSON struct {
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	Version         any               `json:"version"`
	Author          any               `json:"author"`
	Contributors    any               `json:"contributors"`
	License         any               `json:"license"`
	Repository      any               `json:"repository"`
	Homepage        string            `json:"homepage"`
	Funding         any               `json:"funding"`
	Keywords        any               `json:"keywords"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Engines         any               `json:"engines"`
	Files           []string          `json:"files"`
	Scripts         map[string]string `json:"scripts"`
	Main            any               `json:"main"`
	Bin             any               `json:"bin"`
	Types           string            `json:"types"`
	Exports         any               `json:"exports"`
	Private         bool              `json:"private"`
	Workspaces      []string          `json:"workspaces"`
}

type Funding struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type PackageJSONParser struct {
	Config                *config.Config
	LockFileName          string
	PackageJSON           *PackageJSON
	PackageLock           *PackageLock
	FilePath              string
	OriginalContent       []byte
	LockFileContent       []byte
	LockFileContentGlobal []byte
}

type PackageLock struct {
	Name            string                 `json:"name"`
	Version         string                 `json:"version"`
	LockfileVersion int                    `json:"lockfileVersion"`
	Requires        bool                   `json:"requires"`
	Dependencies    map[string]string      `json:"dependencies"`
	DevDependencies map[string]string      `json:"devDependencies,omitempty"`
	Packages        map[string]PackageItem `json:"packages"`
}

type PackageItem struct {
	Name         string            `json:"name,omitempty"`
	Version      string            `json:"version,omitempty"`
	Resolved     string            `json:"resolved,omitempty"`
	Integrity    string            `json:"integrity,omitempty"`
	License      any               `json:"license,omitempty"`
	Etag         string            `json:"etag,omitempty"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
}

func NewPackageJSONParser(cfg *config.Config) *PackageJSONParser {
	return &PackageJSONParser{
		Config:       cfg,
		LockFileName: "go-package-lock.json",
	}
}

func (p *PackageJSONParser) Parse(filePath string) (*PackageJSON, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	var packageJSON PackageJSON
	if err := json.Unmarshal(fileContent, &packageJSON); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from file %s: %w", filePath, err)
	}

	p.PackageJSON = &packageJSON
	p.FilePath = filePath
	p.OriginalContent = fileContent

	lockFileContent, err := os.ReadFile(p.LockFileName)
	if err == nil {
		var packageLock PackageLock
		if err := json.Unmarshal(lockFileContent, &packageLock); err == nil {
			p.PackageLock = &packageLock
			p.LockFileContent = lockFileContent
		}
	}

	return &packageJSON, nil
}

func (p *PackageJSONParser) ParseDefault() (*PackageJSON, error) {
	return p.Parse("package.json")
}

func (p *PackageJSONParser) ParseLockFile() (*PackageLock, error) {
	file, err := os.Open(p.LockFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open file package-lock.json: %w", err)
	}
	defer file.Close()

	var packageLock PackageLock

	if err := json.NewDecoder(file).Decode(&packageLock); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from file package-lock.json: %w", err)
	}

	return &packageLock, nil
}

func (p *PackageJSONParser) CreateLockFile(data *PackageLock, isGlobal bool) error {
	lockFile := p.LockFileName
	if isGlobal {
		lockFile = p.Config.GlobalLockFile
	}

	file, err := os.Create(lockFile)

	if err != nil {
		return fmt.Errorf("failed to create file package-lock.json: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to write JSON to file package-lock.json: %w", err)
	}

	p.PackageLock = data

	return nil
}

func (p *PackageJSONParser) UpdateLockFile(data *PackageLock, isGlobal bool) error {
	lockFileContent := p.LockFileContent
	lockFileName := p.LockFileName

	if isGlobal {
		lockFileContent = p.LockFileContentGlobal
		lockFileName = p.Config.GlobalLockFile
	}

	if lockFileContent == nil {
		return fmt.Errorf("lock file content not cached, call Parse() first")
	}

	var existingLock PackageLock
	if err := json.Unmarshal(lockFileContent, &existingLock); err != nil {
		return fmt.Errorf("failed to parse existing lock file: %w", err)
	}

	for key, version := range data.Dependencies {
		existingLock.Dependencies[key] = version
	}

	if existingLock.Packages == nil {
		existingLock.Packages = make(map[string]PackageItem)
	}

	for key, packageItem := range data.Packages {
		_, ok := existingLock.Packages[key]
		if ok {
			p.resolveVersionMismatch(&existingLock, key, packageItem)
		}
		existingLock.Packages[key] = packageItem
	}

	updatedContent, err := json.MarshalIndent(existingLock, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal updated lock file: %w", err)
	}

	if err := os.WriteFile(lockFileName, updatedContent, 0644); err != nil {
		return fmt.Errorf("failed to write lock file: %w", err)
	}

	p.PackageLock = &existingLock
	if isGlobal {
		p.LockFileContentGlobal = updatedContent
	} else {
		p.LockFileContent = updatedContent
	}

	return nil
}

func (p *PackageJSONParser) resolveVersionMismatch(existingLock *PackageLock, key string, packageItem PackageItem) {
	for keyp, p := range existingLock.Packages {
		if p.Dependencies != nil {
			for depName := range p.Dependencies {
				if depName == packageItem.Name {
					nestedKey := keyp + "/node_modules/" + packageItem.Name
					existingLock.Packages[nestedKey] = existingLock.Packages[key]
					delete(existingLock.Packages, key)
				}
			}
		}
	}
}

func (p *PackageJSONParser) AddOrUpdateDependency(name string, version string) error {
	if p.PackageJSON == nil {
		return fmt.Errorf("package.json not loaded, call Parse() first")
	}

	if p.FilePath == "" {
		return fmt.Errorf("file path not set, call Parse() first")
	}

	if p.OriginalContent == nil {
		return fmt.Errorf("original content not cached, call Parse() first")
	}

	if p.PackageJSON.Dependencies == nil {
		p.PackageJSON.Dependencies = make(map[string]string)
	}
	p.PackageJSON.Dependencies[name] = version

	// Check if dependency already exists (using cached content)
	jsonStr := string(p.OriginalContent)
	existingValue := gjson.Get(jsonStr, "dependencies."+name)
	isNewDependency := !existingValue.Exists()

	// Use sjson to update the dependency
	var err error
	jsonStr, err = sjson.SetRaw(jsonStr, "dependencies."+name, fmt.Sprintf(`"%s"`, version))
	if err != nil {
		return fmt.Errorf("failed to update dependency: %w", err)
	}

	// Fix formatting if it's a new dependency (sjson adds it incorrectly)
	if isNewDependency {
		malformed := "\n  ,\"" + name + `":"` + version + `"}`
		wellFormed := `,` + "\n" + `    "` + name + `": "` + version + `"` + "\n  }"
		jsonStr = strings.Replace(jsonStr, malformed, wellFormed, 1)
	}

	// Write back to file
	if err := os.WriteFile(p.FilePath, []byte(jsonStr), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", p.FilePath, err)
	}

	// Update cached content for subsequent calls
	p.OriginalContent = []byte(jsonStr)

	return nil
}

func (p *PackageJSONParser) ResolveDependencies() (toInstall []Dependency, toRemove []Dependency) {
	toInstall = []Dependency{}
	toRemove = []Dependency{}

	for name, versionInJSON := range p.PackageJSON.Dependencies {
		versionInLock, exists := p.PackageLock.Dependencies[name]
		if !exists || versionInJSON != versionInLock {
			toInstall = append(toInstall, Dependency{
				Name:    name,
				Version: versionInJSON,
			})
		}
	}

	for name, versionInJSON := range p.PackageJSON.DevDependencies {
		versionInLock, exists := p.PackageLock.DevDependencies[name]
		if !exists || versionInJSON != versionInLock {
			toInstall = append(toInstall, Dependency{
				Name:    name,
				Version: versionInJSON,
			})
		}
	}

	for name, versionInLock := range p.PackageLock.Dependencies {
		_, existsInDeps := p.PackageJSON.Dependencies[name]
		_, existsInDevDeps := p.PackageJSON.DevDependencies[name]

		if !existsInDeps && !existsInDevDeps {
			toRemove = append(toRemove, Dependency{
				Name:    name,
				Version: versionInLock,
			})
		}
	}

	return toInstall, toRemove
}

func (p *PackageJSONParser) ResolveDependenciesToRemove(pkg string) []string {
	pkgToKeep := make(map[string]bool)

	for directDep := range p.PackageLock.Dependencies {
		if directDep == pkg {
			continue
		}

		visited := make(map[string]bool)
		queue := []string{directDep}

		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]

			if visited[current] {
				continue
			}
			visited[current] = true
			pkgToKeep[current] = true

			pkgPath := "node_modules/" + current
			pkgItem := p.PackageLock.Packages[pkgPath]

			for childDep := range pkgItem.Dependencies {
				if !visited[childDep] {
					queue = append(queue, childDep)
				}
			}
		}
	}

	pkgToRemove := []string{}
	visited := make(map[string]bool)
	queue := []string{pkg}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true

		pkgPath := "node_modules/" + current
		pkgItem := p.PackageLock.Packages[pkgPath]

		if !pkgToKeep[current] {
			pkgToRemove = append(pkgToRemove, current)
		}

		for childDep := range pkgItem.Dependencies {
			if !visited[childDep] {
				queue = append(queue, childDep)
			}
		}
	}

	return pkgToRemove
}

func (p *PackageJSONParser) RemoveDependencies(pkg string) error {
	if p.PackageJSON == nil {
		return fmt.Errorf("package.json not loaded, call Parse() first")
	}

	if p.PackageJSON.Dependencies == nil {
		return fmt.Errorf("no dependencies found in package.json")
	}

	_, exists := p.PackageJSON.Dependencies[pkg]
	if !exists {
		return fmt.Errorf("dependency '%s' not found in package.json", pkg)
	}

	jsonStr := string(p.OriginalContent)
	var err error
	jsonStr, err = sjson.Delete(jsonStr, "dependencies."+pkg)
	if err != nil {
		return fmt.Errorf("failed to remove dependency from package.json: %w", err)
	}

	if err := os.WriteFile(p.FilePath, []byte(jsonStr), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", p.FilePath, err)
	}

	delete(p.PackageJSON.Dependencies, pkg)
	p.OriginalContent = []byte(jsonStr)

	return nil
}

func (p *PackageJSONParser) RemoveFromLockFile(pkg string, pkgToRemove []string, isGlobal bool) error {
	if p.PackageLock == nil {
		return fmt.Errorf("package lock not loaded")
	}

	delete(p.PackageLock.Dependencies, pkg)

	for _, pkgName := range pkgToRemove {
		delete(p.PackageLock.Packages, "node_modules/"+pkgName)
	}

	packagesToDelete := []string{}
	for key := range p.PackageLock.Packages {
		for _, pkgName := range pkgToRemove {
			if strings.Contains(key, "/node_modules/"+pkgName) {
				packagesToDelete = append(packagesToDelete, key)
			}
		}
	}
	for _, key := range packagesToDelete {
		delete(p.PackageLock.Packages, key)
	}

	err := p.CreateLockFile(p.PackageLock, isGlobal)
	if err != nil {
		return err
	}

	return nil
}
