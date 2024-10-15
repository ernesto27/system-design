package runtimejs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
)

type RuntimeJS struct {
	vm             *goja.Runtime
	jsFileName     string
	queue          chan func()
	done           chan struct{}
	intervals      map[string]*intervalData
	doneInterval   chan struct{}
	intervalsMutex sync.Mutex
}

type intervalData struct {
	ticker *time.Ticker
	done   chan struct{}
}

type Options struct {
	Encoding string      `json:"encoding"`
	Mode     interface{} `json:"mode"`
	Flag     string      `json:"flag"`
}

type FetchOptions struct {
	Method  string
	Headers map[string]string
	Body    string
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
		vm:           vm,
		queue:        make(chan func(), 100),
		done:         make(chan struct{}),
		intervals:    make(map[string]*intervalData),
		doneInterval: make(chan struct{}),
		jsFileName:   fileName,
	}

	runtimeJS.setGlobals()
	packages := []Package{
		&File{runtime: runtimeJS},
		&OS{runtime: runtimeJS},
		&Http{runtime: runtimeJS},
		&Env{runtime: runtimeJS},
		&Require{runtime: runtimeJS},
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
			if arg.ExportType() != nil && arg.ExportType().Kind() == reflect.Map {
				jsonArg, err := json.Marshal(arg)
				if err != nil {
					fmt.Println("Error marshalling argument to JSON:", err)
					return goja.Undefined()
				}
				fmt.Print(string(jsonArg), " ")
			} else {
				fmt.Print(arg.String(), " ")
			}
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

	runtimeJS.intervals = make(map[string]*intervalData)

	runtimeJS.vm.Set("setInterval", func(call goja.FunctionCall) goja.Value {
		callback := call.Argument(0)
		interval := call.Argument(1).ToInteger()

		fn, ok := goja.AssertFunction(callback)
		if !ok {
			panic(runtimeJS.vm.ToValue("TypeError: callback must be a function"))
		}

		ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
		done := make(chan struct{})
		intervalID := fmt.Sprintf("interval_%p", ticker)
		runtimeJS.intervalsMutex.Lock()
		runtimeJS.intervals[intervalID] = &intervalData{ticker: ticker, done: done}
		runtimeJS.intervalsMutex.Unlock()
		fmt.Println(runtimeJS.intervals)
		fmt.Println(intervalID)

		go func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Recovered from panic in interval %s: %v\n", intervalID, r)
				}
			}()

			for {
				select {
				case <-ticker.C:
					_, err := fn(goja.Undefined())
					if err != nil {
						fmt.Printf("Error executing callback for interval %s: %v\n", intervalID, err)
					}
				case <-done:
					return
				}
			}
		}()

		return runtimeJS.vm.ToValue(intervalID)
	})

	runtimeJS.vm.Set("clearInterval", func(call goja.FunctionCall) goja.Value {
		intervalID := call.Argument(0).String()

		runtimeJS.intervalsMutex.Lock()
		defer runtimeJS.intervalsMutex.Unlock()

		if data, ok := runtimeJS.intervals[intervalID]; ok {
			data.ticker.Stop()
			close(data.done)
			delete(runtimeJS.intervals, intervalID)
			fmt.Printf("Interval %s cleared\n", intervalID)
		} else {
			fmt.Printf("Interval %s not found or already cleared\n", intervalID)
		}

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

	runtimeJS.vm.Set("fetch", func(call goja.FunctionCall) goja.Value {
		url := call.Argument(0).String()

		fetchOptions := FetchOptions{
			Method:  "GET",
			Headers: make(map[string]string),
		}
		if !goja.IsUndefined(call.Argument(1)) {
			optionsObj := call.Argument(1).Export()
			if optionsMap, ok := optionsObj.(map[string]interface{}); ok {
				if method, ok := optionsMap["method"].(string); ok {
					fetchOptions.Method = method
				}

				if headers, ok := optionsMap["headers"].(map[string]interface{}); ok {
					for key, value := range headers {
						if strValue, ok := value.(string); ok {
							fetchOptions.Headers[key] = strValue
						}
					}
				}

				if body, ok := optionsMap["body"].(string); ok {
					fetchOptions.Body = body
				}
			}
		}

		fmt.Println(fetchOptions)

		req, err := http.NewRequest(fetchOptions.Method, url, strings.NewReader(fetchOptions.Body))
		if err != nil {
			return runtimeJS.promise([]byte(err.Error()), true)
		}

		for key, value := range fetchOptions.Headers {
			req.Header.Set(key, value)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return runtimeJS.promise([]byte(err.Error()), true)
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return runtimeJS.vm.ToValue(err.Error())
		}

		return runtimeJS.promise(body, false)

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

func (r *RuntimeJS) promise(body []byte, catchError bool) goja.Value {
	return r.vm.ToValue(map[string]interface{}{
		"then": r.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			callback := call.Argument(0)

			fn, ok := goja.AssertFunction(callback)
			if !ok {
				panic("TypeError: callback must be a function")
			}

			_, err := fn(goja.Undefined(), r.vm.ToValue(string(body)))
			if err != nil {
				fmt.Println("Error executing callback:", err)
			}
			return r.promise(body, catchError)
		}),
		"catch": r.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if catchError {
				callback := call.Argument(0)
				fn, ok := goja.AssertFunction(callback)
				if !ok {
					panic("TypeError: callback must be a function")
				}

				_, err := fn(goja.Undefined(), r.vm.ToValue(string(body)))
				fmt.Println("catch error", err)
				if err != nil {
					panic("Error executing catch callback:")
				}
			}
			return goja.Undefined()
		}),
	})
}
