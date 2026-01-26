package manifest

import (
	"npm-packager/utils"
	"path/filepath"
)

type Manifest struct {
	npmResgistryURL string
	Path            string
}

func NewManifest(configPath string, npmRegistryURL string) (*Manifest, error) {
	pathM := filepath.Join(configPath, "manifest")
	if err := utils.CreateDir(pathM); err != nil {
		return nil, err
	}

	return &Manifest{
		Path:            pathM,
		npmResgistryURL: npmRegistryURL,
	}, nil
}

func (m *Manifest) Download(pkg string, currentEtag string) (string, int, error) {
	url := m.npmResgistryURL + pkg
	filename := filepath.Join(m.Path, pkg+".json")

	eTag, statusCode, err := utils.DownloadFile(url, filename, currentEtag)

	return eTag, statusCode, err
}
