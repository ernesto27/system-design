package main

import (
	"os"
	"path"
	"path/filepath"
)

type DownloadTarball struct {
	tarballPath string
}

func newDownloadTarball() *DownloadTarball {
	tarballPath := os.TempDir()
	return &DownloadTarball{tarballPath: tarballPath}
}

func (d *DownloadTarball) download(url string) error {
	filename := path.Base(url)
	filePath := filepath.Join(d.tarballPath, filename)

	_, _, err := downloadFile(url, filePath, "")
	return err
}
