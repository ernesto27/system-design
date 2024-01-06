package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Engine struct {
	m          map[string]int64
	file       *os.File
	fileDelete *os.File
	mu         sync.Mutex
	muDelete   sync.Mutex
}

var keyValueSeparator = " "

func NewEngine(filename string, filenameDelete string) (*Engine, error) {
	var fileData string
	var fileRemove string

	// TODO - find a better way to do this
	if filename == "" && filenameDelete == "" {
		configFolderPath, err := getConfigFolder()
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		if _, err := os.Stat(configFolderPath); os.IsNotExist(err) {
			err := os.Mkdir(configFolderPath, 0700)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
		}

		fileData = configFolderPath + "/" + "db.txt"
		fileRemove = configFolderPath + "/" + "delete.txt"
	} else {
		fileData = filename
		fileRemove = filenameDelete
	}

	file, err := os.OpenFile(fileData, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error opening file data:", err)
		return nil, err
	}

	fileDelete, err := os.OpenFile(fileRemove, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error opening file delete:", err)
		return nil, err
	}

	return &Engine{
		m:          make(map[string]int64),
		file:       file,
		fileDelete: fileDelete,
		mu:         sync.Mutex{},
		muDelete:   sync.Mutex{},
	}, nil
}

func getConfigFolder() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	homeDir := currentUser.HomeDir
	configFolder := ".config/keyvaluedb"
	configFolderPath := filepath.Join(homeDir, configFolder)
	return configFolderPath, nil
}

func (c *Engine) Get(key string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.m[key]; !ok {
		return "", fmt.Errorf("key not found")
	}

	_, err := c.file.Seek(c.m[key]+int64(len(key))+1, 0)
	if err != nil {
		fmt.Println("Error seeking file:", err)
		return "", err
	}

	buffer := make([]byte, 1)
	var content []byte

	for {
		n, err := c.file.Read(buffer)
		if err != nil {
			fmt.Println("Error reading file:", err)
			break
		}

		if n == 0 {
			break
		}

		if buffer[0] == '\n' {
			break
		}

		content = append(content, buffer[0])
	}
	return string(content), nil
}

func (c *Engine) Set(key string, value string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if strings.Contains(key, " ") {
		return fmt.Errorf("key cannot contain spaces")
	}

	return c.setRaw(key, value)
}

func (c *Engine) setRaw(key string, value string) error {
	offset, err := c.saveToFile(key, value)
	if err != nil {
		return err
	}

	c.setKey(key, offset)
	return nil
}

func (c *Engine) setKey(key string, value int64) {
	c.m[key] = value
}

const limit = int64(1024 * 1024)

func (c *Engine) saveToFile(key string, value string) (int64, error) {
	offset, err := c.file.Seek(0, io.SeekEnd)
	if err != nil {
		fmt.Println("Error seeking file:", err)
		return 0, err
	}

	_, err = c.file.WriteString(key + keyValueSeparator + value + "\n")
	if err != nil {
		fmt.Println("Error appending text:", err)
		return 0, err
	}

	err = c.file.Sync()
	if err != nil {
		return 0, err
	}

	return offset, nil
}

const Seconds = 50

func (c *Engine) CompactFile() {
	for {
		time.Sleep(time.Duration(Seconds) * time.Second)
		fmt.Println("Compacting file...")
		c.mu.Lock()

		// TODO - do something with this
		backupFile, err := os.OpenFile("backup.txt", os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644)
		if err != nil {
			fmt.Println("Error creating backup file:", err)
			c.mu.Unlock()
			continue
		}

		_, err = io.Copy(backupFile, c.file)
		if err != nil {
			fmt.Println("Error copying file contents to backup file:", err)
			c.mu.Unlock()
			backupFile.Close()
			continue
		}

		_, m := c.GetMapFromFile()

		err = c.file.Truncate(0)
		if err != nil {
			fmt.Println(err)
			c.mu.Unlock()
			continue
		}

		for k, v := range m {
			c.setRaw(k, v)
		}

		c.file.Seek(0, 0)
		c.mu.Unlock()
		backupFile.Close()

	}
}

func (c *Engine) Restore() {
	c.mu.Lock()
	defer c.mu.Unlock()

	items, _ := c.GetMapFromFile()

	for _, v := range items {
		c.setKey(v.Key, v.Offset)
	}

	c.file.Seek(0, 0)
}

type Item struct {
	Key    string
	Value  string
	Offset int64
}

func (c *Engine) GetMapFromFile() ([]Item, map[string]string) {
	m := make(map[string]string)
	i := []Item{}

	_, err := c.file.Seek(0, 0)
	if err != nil {
		fmt.Println(err)
		return i, m
	}

	var totalBytesRead int64
	scanner := bufio.NewScanner(c.file)

	for scanner.Scan() {
		line := scanner.Text()
		offset := totalBytesRead
		parts := strings.Split(line, keyValueSeparator)
		if len(parts) >= 2 {
			m[parts[0]] = parts[1]
			totalBytesRead += int64(len(line) + 1)
			i = append(i, Item{
				Key:    parts[0],
				Value:  parts[1],
				Offset: offset,
			})
		}
	}

	return i, m
}

func (e *Engine) Delete(key string) error {
	e.muDelete.Lock()
	defer e.muDelete.Unlock()
	_, err := e.fileDelete.Seek(0, io.SeekEnd)
	if err != nil {
		fmt.Println("Error seeking file:", err)
		return err
	}

	_, err = e.fileDelete.WriteString(key + "\n")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.m, key)

	return nil
}

const secondsDelete = 5

func (e *Engine) DeleteFromFile() {
	for {
		time.Sleep(secondsDelete * time.Second)
		fmt.Println("Deleting from file...")
		e.muDelete.Lock()

		_, err := e.fileDelete.Seek(0, 0)
		if err != nil {
			fmt.Println(err)
			e.muDelete.Unlock()
			continue
		}

		scanner := bufio.NewScanner(e.fileDelete)

		content := []string{}
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				content = append(content, line)
			}
		}

		err = e.deleteKeyFromFile(content)
		if err != nil {
			fmt.Println(err)
			e.muDelete.Unlock()
			continue
		}

		err = e.fileDelete.Truncate(0)
		if err != nil {
			fmt.Println(err)
			e.muDelete.Unlock()
			continue
		}

		e.muDelete.Unlock()
	}
}

func (c *Engine) deleteKeyFromFile(keys []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	tempFile, err := os.CreateTemp("", "tempfile_")
	if err != nil {
		return err
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, c.file)
	if err != nil {
		return err
	}

	_, err = c.file.Seek(0, 0)
	if err != nil {
		fmt.Println(err)
		return err
	}

	var bs []byte
	buf := bytes.NewBuffer(bs)

	scanner := bufio.NewScanner(c.file)
	for scanner.Scan() {
		l := scanner.Text()

		parts := strings.Split(l, keyValueSeparator)
		if len(parts) >= 2 {
			found := false
			for _, k := range keys {
				if parts[0] == k {
					found = true
					break
				}
			}

			if !found {
				buf.WriteString(l + "\n")
			}
		}
	}

	_, err = c.file.Seek(0, 0)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = c.file.Truncate(0)
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = buf.WriteTo(c.file)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (c *Engine) GetFileContent(f *os.File) []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := f.Seek(0, 0)
	if err != nil {
		fmt.Println(err)
		return []string{}
	}

	scanner := bufio.NewScanner(f)

	var content []string
	for scanner.Scan() {
		line := scanner.Text()
		content = append(content, line)
	}

	return content
}

func (c *Engine) Close() {
	c.file.Close()
	c.fileDelete.Close()
}
