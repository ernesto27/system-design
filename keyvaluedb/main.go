package main

import (
	"fmt"
	"os"
)

func main() {
	// remove file
	os.Remove("file.txt")
	cache := NewEngine("file.txt")
	defer cache.Close()
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	r := cache.Get("key1")
	fmt.Println(r)

}
