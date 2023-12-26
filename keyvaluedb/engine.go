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
	m    map[string]int64
	file *os.File
	mu   sync.Mutex
}

func NewEngine(filename string) *Engine {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		panic(err)
	}

	return &Engine{
		m:    make(map[string]int64),
		file: file,
		mu:   sync.Mutex{},
	}
}

func (c *Engine) Get(key string) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.m[key]; !ok {
		return ""
	}

	_, err := c.file.Seek(c.m[key]+int64(len(key))+1, 0)
	if err != nil {
		fmt.Println("Error seeking file:", err)
		return ""
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
	return string(content)
}

func (c *Engine) Set(key string, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.setRaw(key, value)
}

func (c *Engine) setRaw(key string, value string) {
	offset, err := c.file.Seek(0, io.SeekEnd)
	if err != nil {
		fmt.Println("Error seeking file:", err)
		panic(err)
	}

	// Append text to the file
	_, err = c.file.WriteString(key + ":" + value + "\n")
	if err != nil {
		fmt.Println("Error appending text:", err)
		return
	}

	c.m[key] = offset
}

const Seconds = 5

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

		m := c.GetMapFromFile()

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

	m := c.GetMapFromFile()

	for k, v := range m {
		c.setRaw(k, v)
	}

	c.file.Seek(0, 0)
}

func (c *Engine) GetMapFromFile() map[string]string {
	m := make(map[string]string)
	_, err := c.file.Seek(0, 0)
	if err != nil {
		fmt.Println(err)
		return m
	}

	scanner := bufio.NewScanner(c.file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) >= 2 {
			m[parts[0]] = parts[1]
		}
	}

	return m
}

func (c *Engine) GetFileContent() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.file.Seek(0, 0)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	scanner := bufio.NewScanner(c.file)

	var content []string
	for scanner.Scan() {
		line := scanner.Text()
		content = append(content, line)
	}

	return strings.Join(content, "\n")
}

func (c *Engine) Close() {
	c.file.Close()
}
