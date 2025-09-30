package main

import (
	"os"
	"path/filepath"
)

type DownloadManifest struct {
	packageName  string
	manifestPath string
	etagPath     string
}

func newDownloadManifest(name string, manifestPath string, etagPath string) *DownloadManifest {
	return &DownloadManifest{
		packageName:  name,
		manifestPath: manifestPath,
		etagPath:     etagPath,
	}
}

func (d *DownloadManifest) getManifestURL() string {
	return npmResgistryURL + d.packageName
}

func (d *DownloadManifest) download() (string, error) {
	url := d.getManifestURL()
	filename := filepath.Join(d.manifestPath, d.packageName+".json")

	if _, err := os.Stat(filename); err == nil {
		return "", nil
	}

	eTag, err := downloadFile(url, filename)

	return eTag, err
}
