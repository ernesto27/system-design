package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type Engine struct {
	data map[string]int64
	file *os.File
	mu   sync.Mutex
}

var keyValueSeparator = " "

func NewEngine() (*Engine, error) {
	file, err := os.OpenFile("data.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error opening file data:", err)
		return nil, err
	}

	return &Engine{
		data: make(map[string]int64),
		file: file,
		mu:   sync.Mutex{},
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

	scanner := bufio.NewScanner(c.file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, keyValueSeparator)
		if len(parts) >= 2 {
			m[parts[0]] = parts[1]
			i = append(i, Item{
				Key:   parts[0],
				Value: parts[1],
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
