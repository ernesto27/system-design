package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type Engine struct {
	data       map[string]int64
	file       *os.File
	fileDelete *os.File
	mu         sync.Mutex
	muDelete   sync.Mutex
}

var keyValueSeparator = " "

func NewEngine() (*Engine, error) {
	file, err := os.OpenFile("data.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error opening file data:", err)
		return nil, err
	}

	fileDelete, err := os.OpenFile("delete.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error opening file delete:", err)
		return nil, err
	}

	return &Engine{
		data:       make(map[string]int64),
		file:       file,
		fileDelete: fileDelete,
		mu:         sync.Mutex{},
		muDelete:   sync.Mutex{},
	}, nil
}

func (e *Engine) Set(key string, value string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if strings.Contains(key, " ") {
		return fmt.Errorf("key cannot contain spaces")
	}

	return e.setRaw(key, value)
}

func (e *Engine) Get(key string) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.data[key]; !ok {
		return "", fmt.Errorf("key not found")
	}

	_, err := e.file.Seek(e.data[key]+int64(len(key))+1, 0)
	if err != nil {
		fmt.Println("Error seeking file:", err)
		return "", err
	}

	buffer := make([]byte, 1)
	var content []byte

	for {
		n, err := e.file.Read(buffer)
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

const Seconds = 5

func (e *Engine) CompactFile() {
	for {
		time.Sleep(time.Duration(Seconds) * time.Second)
		fmt.Println("Compacting file...")
		e.mu.Lock()

		tempFile, err := os.OpenFile("temp.txt", os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644)
		if err != nil {
			fmt.Println("Error creating backup file:", err)
			e.mu.Unlock()
			continue
		}

		_, err = io.Copy(tempFile, e.file)
		if err != nil {
			fmt.Println("Error copying file contents to backup file:", err)
			e.mu.Unlock()
			tempFile.Close()
			continue
		}

		_, m := e.GetMapFromFile()

		err = e.file.Truncate(0)
		if err != nil {
			fmt.Println(err)
			e.mu.Unlock()
			continue
		}

		for k, v := range m {
			e.setRaw(k, v)
		}

		e.file.Seek(0, 0)
		e.mu.Unlock()
		tempFile.Close()
	}
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

func (e *Engine) setRaw(key string, value string) error {
	offset, err := e.saveToFile(key, value)
	if err != nil {
		return err
	}

	e.setKey(key, offset)
	return nil
}

func (e *Engine) setKey(key string, value int64) {
	e.data[key] = value
}

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

	return offset, nil
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

func (e *Engine) Restore() {
	e.mu.Lock()
	defer e.mu.Unlock()

	items, _ := e.GetMapFromFile()

	for _, v := range items {
		e.setKey(v.Key, v.Offset)
	}
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
	delete(e.data, key)

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

	_, err := c.file.Seek(0, 0)
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

	_, err = c.file.Seek(0, 0)
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

func (c *Engine) Close() {
	c.file.Close()
}
