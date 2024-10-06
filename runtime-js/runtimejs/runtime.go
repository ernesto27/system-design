package runtimejs

import (
	"fmt"
	"os"
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

type Package interface {
	SetGlobals()
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
	packages := []Package{
		&File{runtime: runtimeJS},
		&OS{runtime: runtimeJS},
		&Http{runtime: runtimeJS},
	}

	for _, pkg := range packages {
		pkg.SetGlobals()
	}

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
