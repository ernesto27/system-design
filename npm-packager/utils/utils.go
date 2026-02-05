package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func DownloadFile(url, filename string, etag string) (string, int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return etag, resp.StatusCode, nil
	}

	if resp.StatusCode != http.StatusOK {
		return "", resp.StatusCode, fmt.Errorf("HTTP error: %s, %d %s", url, resp.StatusCode, resp.Status)
	}

	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", resp.StatusCode, fmt.Errorf("failed to create directory structure: %w", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return "", resp.StatusCode, fmt.Errorf("failed to create file: %w", err)
	}

	_, err = io.Copy(file, resp.Body)
	file.Close()

	if err != nil {
		os.Remove(filename)
		return "", resp.StatusCode, fmt.Errorf("failed to write file: %w", err)
	}

	return resp.Header.Get("ETag"), resp.StatusCode, nil
}

func CreateDir(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.Mkdir(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
		}
		fmt.Printf("Created directory: %s\n", dirPath)
	}
	return nil
}

func CreateDepKey(name, version, parentName string) string {
	return name + "@" + version + "@" + parentName
}

func FolderExists(dirPath string) bool {
	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && info.IsDir()
}
