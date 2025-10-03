package main

import (
	"path"
	"path/filepath"
)

type DownloadTarball struct {
	url         string
	tarballPath string
}

func newDownloadTarball(url string, tarballPath string) *DownloadTarball {
	return &DownloadTarball{url: url, tarballPath: tarballPath}
}

func (d *DownloadTarball) download() error {
	filename := path.Base(d.url)
	filePath := filepath.Join(d.tarballPath, filename)

	_, _, err := downloadFile(d.url, filePath, "")
	return err
}
