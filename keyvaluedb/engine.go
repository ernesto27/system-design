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
	//file, err := os.Open("file.txt")
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

func (c *Engine) Get(key string) []byte {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.m[key]; !ok {
		return []byte{}
	}

	_, err := c.file.Seek(c.m[key]+int64(len(key))+1, 0)
	if err != nil {
		fmt.Println("Error seeking file:", err)
		return nil
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
	return content
}

func (c *Engine) Set(key string, value string) {
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

func (c *Engine) CompactFile() {
	for {
		time.Sleep(5 * time.Second)
		fmt.Println("Compacting file...")
		c.mu.Lock()

		backupFile, err := os.OpenFile("backup.txt", os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Println("Error creating backup file:", err)
			c.mu.Unlock()
			continue
		}

		fmt.Println(c.file)
		_, err = io.Copy(backupFile, c.file)
		if err != nil {
			fmt.Println("Error copying file contents to backup file:", err)
			c.mu.Unlock()
			backupFile.Close()
			continue
		}

		_, err = c.file.Seek(0, 0)
		if err != nil {
			fmt.Println(err)
			c.mu.Unlock()
			continue
		}

		scanner := bufio.NewScanner(c.file)

		m := make(map[string]string)
		for scanner.Scan() {
			line := scanner.Text()
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				m[parts[0]] = parts[1]
			}
		}

		err = c.file.Truncate(0)
		if err != nil {
			fmt.Println(err)
			c.mu.Unlock()
			continue
		}

		for k, v := range m {
			c.Set(k, v)
		}

		c.file.Seek(0, 0)
		c.mu.Unlock()
		backupFile.Close()

	}
}

func (c *Engine) Close() {
	c.file.Close()
}
