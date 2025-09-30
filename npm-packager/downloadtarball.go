package main

import (
	"os"
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

	// Check if file already exists,  do not download again
	if _, err := os.Stat(filePath); err == nil {
		return nil
	}

	_, err := downloadFile(d.url, filePath)
	return err
}
