package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Etag struct {
	packages map[string]Dependency
	etagPath string
	etagData map[string]EtagEntry
}

type EtagEntry struct {
	Etag string `json:"etag"`
}

func newEtag(etagPath string) *Etag {

	etagData := make(map[string]EtagEntry)
	etagFilePath := filepath.Join(etagPath, "etag.json")

	if existingData, err := os.ReadFile(etagFilePath); err == nil {
		if err := json.Unmarshal(existingData, &etagData); err != nil {
			fmt.Printf("Warning: failed to unmarshal existing etag data: %v\n", err)
		}
	}

	return &Etag{
		etagPath: etagPath,
		etagData: etagData,
	}
}

func (e *Etag) setPackages(packages map[string]Dependency) {
	e.packages = packages
}

func (e *Etag) get(packageName string) string {
	if entry, ok := e.etagData[packageName]; ok {
		return entry.Etag
	}
	return ""
}

func (e *Etag) save() error {
	etagFilePath := filepath.Join(e.etagPath, "etag.json")

	for pkgName, dep := range e.packages {
		if dep.Etag != "" {
			e.etagData[pkgName] = EtagEntry{
				Etag: dep.Etag,
			}
		}
	}

	jsonData, err := json.MarshalIndent(e.etagData, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal etag data: %w", err)
	}

	if err := os.WriteFile(etagFilePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write etag file: %w", err)
	}

	return nil
}
