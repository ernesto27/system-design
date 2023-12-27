package main

import (
	"fmt"
	"strconv"
)

func main() {
	// remove file
	//os.Remove("file.txt")
	cache := NewEngine("file.txt")
	defer cache.Close()
	// cache.Set("key4", "value1")
	// cache.Set("key6", "value2")
	// fmt.Println(cache.Get("key4"))
	// return

	cache.Restore()
	fmt.Println(cache.Get("key7966"))
	return

	lorem := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec a diam lectus. Sed sit amet ipsum mauris. "

	for i := 0; i < 1000000; i++ {
		cache.Set("key"+strconv.Itoa(i), lorem)
	}

	return

	go cache.CompactFile()

	exitSignal := make(chan struct{})
	<-exitSignal

}
