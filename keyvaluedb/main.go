package main

func main() {
	// remove file
	//os.Remove("file.txt")
	cache := NewEngine("file.txt")
	defer cache.Close()
	// cache.Set("key4", "value1")
	// cache.Set("key6", "value2")
	// r := cache.Get("key1")
	// fmt.Println(r)

	go cache.CompactFile()

	exitSignal := make(chan struct{})
	<-exitSignal

}
