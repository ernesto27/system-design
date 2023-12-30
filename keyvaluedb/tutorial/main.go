package main

import "fmt"

func main() {
	e, _ := NewEngine()
	e.Set("foo", "bar")
	value, err := e.Get("foo")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(value)

}
