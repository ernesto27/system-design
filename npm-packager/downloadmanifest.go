package main

import (
	"path/filepath"
)

type DownloadManifest struct {
	packageName  string
	manifestPath string
}

func newDownloadManifest(name string, manifestPath string) *DownloadManifest {
	return &DownloadManifest{
		packageName:  name,
		manifestPath: manifestPath,
	}
}

func (d *DownloadManifest) getManifestURL() string {
	return npmResgistryURL + d.packageName
}

func (d *DownloadManifest) download(currentEtag string) (string, int, error) {
	url := d.getManifestURL()
	filename := filepath.Join(d.manifestPath, d.packageName+".json")

	// if _, err := os.Stat(filename); err == nil {
	// 	return "", 0, nil
	// }

	eTag, statusCode, err := downloadFile(url, filename, currentEtag)

	return eTag, statusCode, err
}
