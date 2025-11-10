package manager

import (
	"encoding/json"
	"fmt"
	"os"
)

type NPMPackage struct {
	ID       string             `json:"_id"`
	Rev      string             `json:"_rev"`
	Name     string             `json:"name"`
	DistTags DistTags           `json:"dist-tags"`
	Versions map[string]Version `json:"versions"`
	Time     map[string]string  `json:"time"`
	Bugs     any                `json:"bugs"`
	License  any                `json:"license"`
	Homepage string             `json:"homepage"`
	Keywords any                `json:"keywords"`

	Repository     any             `json:"repository"`
	Description    string          `json:"description"`
	Contributors   any             `json:"contributors"`
	Maintainers    []Maintainer    `json:"maintainers"`
	Readme         string          `json:"readme"`
	ReadmeFilename string          `json:"readmeFilename"`
	Users          map[string]bool `json:"users"`
}

type DistTags struct {
	Latest string `json:"latest"`
	Next   string `json:"next"`
}

type Version struct {
	Name                   string                 `json:"name"`
	Version                string                 `json:"version"`
	Author                 any                    `json:"author"`
	License                any                    `json:"license"`
	ID                     string                 `json:"_id"`
	Maintainers            []Maintainer           `json:"maintainers"`
	Homepage               string                 `json:"homepage"`
	Bugs                   any                    `json:"bugs"`
	Dist                   Dist                   `json:"dist"`
	From                   string                 `json:"_from"`
	Shasum                 string                 `json:"_shasum"`
	Engines                any                    `json:"engines"`
	GitHead                string                 `json:"gitHead"`
	Scripts                map[string]string      `json:"scripts"`
	NPMUser                NPMUser                `json:"_npmUser"`
	Repository             any                    `json:"repository"`
	NPMVersion             string                 `json:"_npmVersion"`
	Description            string                 `json:"description"`
	Directories            map[string]interface{} `json:"directories"`
	NodeVersion            string                 `json:"_nodeVersion"`
	Dependencies           map[string]string      `json:"dependencies"`
	DevDependencies        map[string]string      `json:"devDependencies"`
	HasShrinkwrap          bool                   `json:"_hasShrinkwrap"`
	Keywords               any                    `json:"keywords"`
	Contributors           any                    `json:"contributors"`
	Files                  []string               `json:"files"`
	NPMOperationalInternal NPMOperationalInternal `json:"_npmOperationalInternal"`
	NPMSignature           string                 `json:"npm-signature"`
}

type Author struct {
	URL   string `json:"url"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Maintainer struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Contributor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	URL   string `json:"url"`
}

type Bugs struct {
	URL string `json:"url"`
}

type Dist struct {
	Shasum       string      `json:"shasum"`
	Tarball      string      `json:"tarball"`
	Integrity    string      `json:"integrity"`
	Signatures   []Signature `json:"signatures"`
	FileCount    int         `json:"fileCount"`
	UnpackedSize int         `json:"unpackedSize"`
}

type Signature struct {
	Sig   string `json:"sig"`
	KeyID string `json:"keyid"`
}

type NPMUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Repository struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

type NPMOperationalInternal struct {
	Tmp  string `json:"tmp"`
	Host string `json:"host"`
}

type ParseJsonManifest struct {
}

func newParseJsonManifest() *ParseJsonManifest {
	return &ParseJsonManifest{}
}

func (p *ParseJsonManifest) parse(filePath string) (*NPMPackage, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	var npmPackage NPMPackage
	if err := json.NewDecoder(file).Decode(&npmPackage); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from file %s: %w", filePath, err)
	}

	return &npmPackage, nil
}
