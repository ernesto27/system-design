package main

import (
	"path"
	"path/filepath"
)

type DownloadTarball struct {
	tarballPath string
}

func newDownloadTarball(configPath string) (*DownloadTarball, error) {
	tarballPath := filepath.Join(configPath, "tarball")
	if err := createDir(tarballPath); err != nil {
		return nil, err
	}
	return &DownloadTarball{tarballPath: tarballPath}, nil
}

func (d *DownloadTarball) download(url string) error {
	filename := path.Base(url)
	filePath := filepath.Join(d.tarballPath, filename)

	_, _, err := downloadFile(url, filePath, "")
	return err
}
