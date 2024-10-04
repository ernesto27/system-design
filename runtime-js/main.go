package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/dop251/goja"
)

type RuntimeJS struct {
	vm    *goja.Runtime
	queue chan func()
	done  chan struct{}
}

type Options struct {
	Encoding string      `json:"encoding"`
	Mode     interface{} `json:"mode"`
	Flag     string      `json:"flag"`
}

func NewRuntimeJS(fileName string) (*RuntimeJS, error) {
	content, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", fileName, err)
		return nil, err
	}

	vm := goja.New()

	runtimeJS := &RuntimeJS{
		vm:    vm,
		queue: make(chan func(), 100),
		done:  make(chan struct{}),
	}

	runtimeJS.setGlobals()

	_, err = vm.RunString(string(content))
	if err != nil {
		return nil, err
	}

	return runtimeJS, nil
}

func (runtimeJS *RuntimeJS) setGlobals() {
	console := runtimeJS.vm.NewObject()
	console.Set("log", func(call goja.FunctionCall) goja.Value {
		for _, arg := range call.Arguments {
			fmt.Print(arg.String(), " ")
		}
		fmt.Println()
		return goja.Undefined()
	})
	runtimeJS.vm.Set("console", console)

	runtimeJS.vm.Set("setTimeout", func(call goja.FunctionCall) goja.Value {
		callback := call.Argument(0)
		delay := call.Argument(1).ToInteger()

		fn, ok := goja.AssertFunction(callback)
		if !ok {
			panic(runtimeJS.vm.ToValue("TypeError: callback must be a function"))
		}

		go func() {
			time.Sleep(time.Duration(delay) * time.Millisecond)
			_, err := fn(goja.Undefined(), goja.Undefined())
			if err != nil {
				fmt.Println("Error executing callback:", err)
			}
		}()

		return goja.Undefined()
	})

	runtimeJS.vm.Set("setInterval", func(call goja.FunctionCall) goja.Value {
		callback := call.Argument(0)
		interval := call.Argument(1).ToInteger()

		fn, ok := goja.AssertFunction(callback)
		if !ok {
			panic(runtimeJS.vm.ToValue("TypeError: callback must be a function"))
		}

		go func() {
			ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
			defer ticker.Stop()

			for range ticker.C {
				_, err := fn(goja.Undefined())
				if err != nil {
					fmt.Println("Error executing callback:", err)
				}
			}
		}()

		return goja.Undefined()
	})

	runtimeJS.vm.Set("readFile", func(call goja.FunctionCall) goja.Value {
		filename := call.Argument(0).String()
		callback := call.Argument(1)

		fn, ok := goja.AssertFunction(callback)
		if !ok {
			panic(runtimeJS.vm.ToValue("TypeError: callback must be a function"))
		}

		go func() {
			fmt.Println("Starting file read operation")
			data, err := os.ReadFile(filename)

			fmt.Println("File read complete, queueing callback")
			runtimeJS.queue <- func() {
				if err != nil {
					_, _ = fn(goja.Undefined(), runtimeJS.vm.ToValue(err.Error()), goja.Undefined())
				} else {
					_, _ = fn(goja.Undefined(), goja.Undefined(), runtimeJS.vm.ToValue(string(data)))
				}
			}
		}()
		return goja.Undefined()
	})

	runtimeJS.vm.Set("mkdir", func(call goja.FunctionCall) goja.Value {
		dirPath := call.Argument(0).String()
		callback := call.Argument(1)

		fn, ok := goja.AssertFunction(callback)
		if !ok {
			panic(runtimeJS.vm.ToValue("TypeError: callback must be a function"))
		}

		go func() {
			err := os.MkdirAll(dirPath, os.ModePerm)
			if err != nil {
				runtimeJS.queue <- func() {
					_, _ = fn(goja.Undefined(), runtimeJS.vm.ToValue(err.Error()), goja.Undefined())
				}
			} else {
				runtimeJS.queue <- func() {
					_, _ = fn(goja.Undefined(), goja.Undefined(), runtimeJS.vm.ToValue(dirPath))
				}
			}
		}()
		return goja.Undefined()
	})

	runtimeJS.vm.Set("writeFile", func(call goja.FunctionCall) goja.Value {
		filename := call.Argument(0).String()
		data := call.Argument(1).String()

		rawObj := call.Argument(2).Export()
		fmt.Printf("Raw object: %+v\n", rawObj)

		var options Options
		if obj, ok := rawObj.(map[string]interface{}); ok {
			if encoding, ok := obj["encoding"].(string); ok {
				options.Encoding = encoding
			}
			options.Mode = obj["mode"]
			if flag, ok := obj["flag"].(string); ok {
				options.Flag = flag
			}
		}
		fmt.Printf("Parsed options: %+v\n", options)

		// Handle Mode
		var fileMode os.FileMode = 0666 // Default value
		if modeVal := options.Mode; modeVal != nil {
			switch m := modeVal.(type) {
			case float64:
				fileMode = os.FileMode(int64(m))
			case int64:
				fileMode = os.FileMode(m)
			case string:
				if parsed, err := strconv.ParseInt(m, 0, 32); err == nil {
					fileMode = os.FileMode(parsed)
				}
			default:
				fmt.Printf("Unexpected type for Mode: %T\n", modeVal)
			}
		}

		// Handle Flag
		var flag int
		switch options.Flag {
		case "w":
			flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		case "a":
			flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
		default:
			flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		}

		fmt.Printf("File mode: %o\n", fileMode)
		fmt.Printf("Flag: %v\n", flag)
		fmt.Println("End")

		callback := call.Argument(3)
		fn, ok := goja.AssertFunction(callback)
		if !ok {
			panic(runtimeJS.vm.ToValue("TypeError: callback must be a function"))
		}

		go func() {
			file, err := os.OpenFile(filename, flag, fileMode)
			if err != nil {
				panic(err)
			}
			defer file.Close()
			_, err = file.Write([]byte(data))
			if err != nil {
				runtimeJS.queue <- func() {
					_, _ = fn(goja.Undefined(), runtimeJS.vm.ToValue(err.Error()), goja.Undefined())
				}
			} else {
				runtimeJS.queue <- func() {
					_, _ = fn(goja.Undefined(), goja.Undefined(), runtimeJS.vm.ToValue(filename))
				}
			}
		}()
		return goja.Undefined()
	})

	runtimeJS.vm.Set("eventLoop", runtimeJS.vm.NewObject())
	runtimeJS.vm.Get("eventLoop").(*goja.Object).Set("done", func() {
		go func() {
			runtimeJS.done <- struct{}{}
		}()
	})
}

func (el *RuntimeJS) RunEventLoop() {
	for {
		select {
		case f := <-el.queue:
			fmt.Println("Executing queued function")
			f()
		case <-el.done:
			fmt.Println("Received done signal, exiting event loop")
			return
		case <-time.After(time.Millisecond * 100):
			//fmt.Println("Timeout: no tasks in queue")
		}
	}
}

func main() {
	args := os.Args

	if len(args) < 2 {
		fmt.Println("Please provide a filename as an argument")
		os.Exit(1)
	}

	filename := args[1]

	runtimeJS, err := NewRuntimeJS(filename)
	if err != nil {
		fmt.Println("Error creating runtimeJS:", err)
		os.Exit(1)
	}
	runtimeJS.RunEventLoop()
}
