package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type PackageJSON struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Version      string            `json:"version"`
	Author       interface{}       `json:"author"`
	Contributors interface{}       `json:"contributors"`
	License      string            `json:"license"`
	Repository   interface{}       `json:"repository"`
	Homepage     string            `json:"homepage"`
	Funding      interface{}       `json:"funding"`
	Keywords     []string          `json:"keywords"`
	Dependencies map[string]string `json:"dependencies"`
	Engines      map[string]string `json:"engines"`
	Files        []string          `json:"files"`
	Scripts      map[string]string `json:"scripts"`
	Main         string            `json:"main"`
	Bin          interface{}       `json:"bin"`
	Types        string            `json:"types"`
	Exports      interface{}       `json:"exports"`
	Private      bool              `json:"private"`
	Workspaces   []string          `json:"workspaces"`
}

type Funding struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type PackageJSONParser struct {
	filePath string
}

func newPackageJSONParser(path string) *PackageJSONParser {
	return &PackageJSONParser{filePath: path}
}

func (p *PackageJSONParser) parse() (*PackageJSON, error) {
	file, err := os.Open(p.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", p.filePath, err)
	}
	defer file.Close()

	var packageJSON PackageJSON
	if err := json.NewDecoder(file).Decode(&packageJSON); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from file %s: %w", p.filePath, err)
	}

	return &packageJSON, nil
}

func (p *PackageJSON) hasDependency(name string) bool {
	if p.Dependencies != nil {
		if _, exists := p.Dependencies[name]; exists {
			return true
		}
	}
	return false
}

func (p *PackageJSON) getDependencyVersion(name string) (string, bool) {
	if p.Dependencies != nil {
		if version, exists := p.Dependencies[name]; exists {
			return version, true
		}
	}
	return "", false
}

func (p *PackageJSON) getTotalDependencies() int {
	if p.Dependencies != nil {
		return len(p.Dependencies)
	}
	return 0
}