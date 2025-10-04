package main

import (
	"path/filepath"
)

type DownloadManifest struct {
	manifestPath string
}

func newDownloadManifest(manifestPath string) *DownloadManifest {
	return &DownloadManifest{
		manifestPath: manifestPath,
	}
}

func (d *DownloadManifest) download(pkg string, currentEtag string) (string, int, error) {
	url := npmResgistryURL + pkg
	filename := filepath.Join(d.manifestPath, pkg+".json")

	// if _, err := os.Stat(filename); err == nil {
	// 	return "", 0, nil
	// }

	eTag, statusCode, err := downloadFile(url, filename, currentEtag)

	return eTag, statusCode, err
}
