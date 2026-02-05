package tarball

import (
	"npm-packager/utils"
	"os"
	"path"
	"path/filepath"
)

type Tarball struct {
	TarballPath string
}

func NewTarball() *Tarball {
	tarballPath := os.TempDir()
	return &Tarball{TarballPath: tarballPath}
}

func (d *Tarball) Download(url string) error {
	filename := path.Base(url)
	filePath := filepath.Join(d.TarballPath, filename)

	_, _, err := utils.DownloadFile(url, filePath, "")
	return err
}
