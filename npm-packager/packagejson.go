package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type PackageJSON struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Version      any               `json:"version"`
	Author       any               `json:"author"`
	Contributors any               `json:"contributors"`
	License      string            `json:"license"`
	Repository   any               `json:"repository"`
	Homepage     string            `json:"homepage"`
	Funding      any               `json:"funding"`
	Keywords     any               `json:"keywords"`
	Dependencies map[string]string `json:"dependencies"`
	Engines      any               `json:"engines"`
	Files        []string          `json:"files"`
	Scripts      map[string]string `json:"scripts"`
	Main         any               `json:"main"`
	Bin          any               `json:"bin"`
	Types        string            `json:"types"`
	Exports      any               `json:"exports"`
	Private      bool              `json:"private"`
	Workspaces   []string          `json:"workspaces"`
}

type Funding struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type PackageJSONParser struct {
	lockFileName string
}

func newPackageJSONParser() *PackageJSONParser {
	return &PackageJSONParser{
		lockFileName: "go-package-lock.json",
	}
}

func (p *PackageJSONParser) parse(filePath string) (*PackageJSON, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	var packageJSON PackageJSON
	if err := json.NewDecoder(file).Decode(&packageJSON); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from file %s: %w", filePath, err)
	}

	return &packageJSON, nil
}

func (p *PackageJSONParser) parseLockFile() (*PackageLock, error) {
	file, err := os.Open(p.lockFileName)
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

func (p *PackageJSONParser) createLockFile(data *PackageLock) error {
	file, err := os.Create(p.lockFileName)

	if err != nil {
		return fmt.Errorf("failed to create file package-lock.json: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to write JSON to file package-lock.json: %w", err)
	}

	return nil
}

type PackageLock struct {
	Name            string                 `json:"name"`
	Version         string                 `json:"version"`
	LockfileVersion int                    `json:"lockfileVersion"`
	Requires        bool                   `json:"requires"`
	Packages        map[string]PackageItem `json:"packages"`
}

type PackageItem struct {
	Name         string            `json:"name,omitempty"`
	Version      string            `json:"version,omitempty"`
	Resolved     string            `json:"resolved,omitempty"`
	Integrity    string            `json:"integrity,omitempty"`
	License      string            `json:"license,omitempty"`
	Etag         string            `json:"etag,omitempty"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
}
