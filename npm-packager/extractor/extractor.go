package extractor

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type TGZExtractor struct {
	bufferSize int
}

func NewTGZExtractor() *TGZExtractor {
	return &TGZExtractor{
		bufferSize: 32 * 1024,
	}
}

func (e *TGZExtractor) Extract(srcPath, destPath string) error {
	file, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", srcPath, err)
	}
	defer file.Close()

	bufReader := bufio.NewReaderSize(file, e.bufferSize)

	gzr, err := gzip.NewReader(bufReader)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	copyBuffer := make([]byte, e.bufferSize)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		relativePath := e.stripPackagePrefix(header.Name)
		if relativePath == "" {
			continue
		}
		target := filepath.Join(destPath, relativePath)

		if !e.isValidPath(target, destPath) {
			fmt.Printf("Skipping unsafe path: %s\n", header.Name)
			continue
		}

		switch header.Typeflag {
		case tar.TypeReg:
			if err := e.extractFile(tr, target, header, copyBuffer); err != nil {
				return err
			}
		default:
			fmt.Printf("Skipping unsupported file type: %c for %s\n", header.Typeflag, header.Name)
		}
	}

	return nil
}

func (e *TGZExtractor) isValidPath(target string, destPath string) bool {
	cleanDest := filepath.Clean(destPath) + string(os.PathSeparator)
	cleanTarget := filepath.Clean(target)
	return strings.HasPrefix(cleanTarget, cleanDest)
}

func (e *TGZExtractor) extractFile(tr *tar.Reader, target string, header *tar.Header, copyBuffer []byte) error {
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory for %s: %w", target, err)
	}

	f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", target, err)
	}
	defer f.Close()

	_, err = io.CopyBuffer(f, tr, copyBuffer)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", target, err)
	}

	return nil
}

func (e *TGZExtractor) stripPackagePrefix(path string) string {
	if idx := strings.Index(path, "/"); idx != -1 {
		return path[idx+1:]
	}
	return ""
}
