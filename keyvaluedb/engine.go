package main

import (
	"fmt"
	"io"
	"os"
)

type Engine struct {
	m    map[string]int64
	file *os.File
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
	}
}

func (c *Engine) Get(key string) []byte {
	// TODO VALIDATE KEY EXISTS
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

func (c *Engine) Close() {
	c.file.Close()
}
