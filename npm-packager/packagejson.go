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
