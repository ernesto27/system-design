package binlink

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type BinLinker struct {
	nodeModulesPath string
	binPath         string
	isGlobal        bool
}

type PackageJSON struct {
	Name string          `json:"name"`
	Bin  json.RawMessage `json:"bin"`
}

func NewBinLinker(nodeModulesPath string) *BinLinker {
	return &BinLinker{
		nodeModulesPath: nodeModulesPath,
		binPath:         filepath.Join(nodeModulesPath, ".bin"),
		isGlobal:        false,
	}
}

func (bl *BinLinker) SetGlobalMode(nodeModulesPath string, globalBinPath string) {
	bl.nodeModulesPath = nodeModulesPath
	bl.binPath = globalBinPath
	bl.isGlobal = true
}

func (bl *BinLinker) CreateBinDirectory() error {
	if err := os.MkdirAll(bl.binPath, 0755); err != nil {
		return fmt.Errorf("failed to create .bin directory: %w", err)
	}
	return nil
}

func (bl *BinLinker) LinkAllPackages() error {
	if err := bl.CreateBinDirectory(); err != nil {
		return err
	}

	entries, err := os.ReadDir(bl.nodeModulesPath)
	if err != nil {
		return fmt.Errorf("failed to read node_modules: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == ".bin" {
			continue
		}

		pkgPath := filepath.Join(bl.nodeModulesPath, entry.Name())

		// Handle scoped packages (@scope/package)
		if entry.Name()[0] == '@' {
			scopedEntries, err := os.ReadDir(pkgPath)
			if err != nil {
				continue
			}
			for _, scopedEntry := range scopedEntries {
				if scopedEntry.IsDir() {
					scopedPkgPath := filepath.Join(pkgPath, scopedEntry.Name())
					if err := bl.LinkPackage(scopedPkgPath); err != nil {
						fmt.Printf("Warning: failed to link %s: %v\n", scopedPkgPath, err)
					}
				}
			}
		} else {
			if err := bl.LinkPackage(pkgPath); err != nil {
				fmt.Printf("Warning: failed to link %s: %v\n", pkgPath, err)
			}
		}
	}

	return nil
}

func (bl *BinLinker) LinkPackage(pkgPath string) error {
	packageJSONPath := filepath.Join(pkgPath, "package.json")

	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return nil
	}

	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil
	}

	if len(pkg.Bin) == 0 {
		return nil
	}

	bins, err := bl.parseBinField(pkg.Name, pkg.Bin)
	if err != nil {
		return err
	}

	for binName, binPath := range bins {
		if err := bl.createSymlink(pkgPath, binName, binPath); err != nil {
			return err
		}
	}

	return nil
}

func (bl *BinLinker) parseBinField(pkgName string, binField json.RawMessage) (map[string]string, error) {
	bins := make(map[string]string)

	// Try parsing as string first
	var binString string
	if err := json.Unmarshal(binField, &binString); err == nil {
		// For scoped packages (@scope/name), use only the name part after /
		binName := pkgName
		if pkgName[0] == '@' {
			if idx := filepath.Base(pkgName); idx != "" {
				binName = idx
			}
		}
		bins[binName] = binString
		return bins, nil
	}

	// Try parsing as object
	var binObject map[string]string
	if err := json.Unmarshal(binField, &binObject); err == nil {
		return binObject, nil
	}

	return nil, fmt.Errorf("invalid bin field format")
}

func (bl *BinLinker) createSymlink(pkgPath, binName, binRelativePath string) error {
	binRelativePath = filepath.Clean(binRelativePath)

	var targetPath string
	linkPath := filepath.Join(bl.binPath, binName)

	if bl.isGlobal {
		// For global installations, use absolute path
		targetPath = filepath.Join(pkgPath, binRelativePath)
	} else {
		// For local installations, use relative path
		// Get package name relative to node_modules
		// For scoped packages: pkgPath = "node_modules/@babel/cli" -> pkgName = "@babel/cli"
		// For normal packages: pkgPath = "node_modules/nodemon" -> pkgName = "nodemon"
		relPath, err := filepath.Rel(bl.nodeModulesPath, pkgPath)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}
		targetPath = filepath.Join("..", relPath, binRelativePath)
	}

	// Check if symlink already exists and is correct
	if existingTarget, err := os.Readlink(linkPath); err == nil {
		if existingTarget == targetPath {
			return nil
		}
		if err := os.Remove(linkPath); err != nil {
			return fmt.Errorf("failed to remove existing link %s: %w", linkPath, err)
		}
	}

	if err := os.Symlink(targetPath, linkPath); err != nil {
		return fmt.Errorf("failed to create symlink %s -> %s: %w", linkPath, targetPath, err)
	}

	fmt.Printf("Linked bin: %s -> %s\n", binName, targetPath)
	return nil
}

func (bl *BinLinker) UnlinkPackage(pkgName string) error {
	pkgPath := filepath.Join(bl.nodeModulesPath, pkgName)
	packageJSONPath := filepath.Join(pkgPath, "package.json")

	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return nil
	}

	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil
	}

	if len(pkg.Bin) == 0 {
		return nil
	}

	bins, err := bl.parseBinField(pkg.Name, pkg.Bin)
	if err != nil {
		return err
	}

	for binName := range bins {
		linkPath := filepath.Join(bl.binPath, binName)
		if err := os.Remove(linkPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove symlink %s: %w", linkPath, err)
		}
		fmt.Printf("Unlinked bin: %s\n", binName)
	}

	return nil
}
