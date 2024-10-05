package runtimejs

import (
	"fmt"
	"os"
	"strconv"

	"github.com/dop251/goja"
)

type File struct {
	runtime *RuntimeJS
}

func (f *File) SetGlobals() {
	f.runtime.vm.Set("readFile", func(call goja.FunctionCall) goja.Value {
		filename := call.Argument(0).String()
		callback := call.Argument(1)

		fn, ok := goja.AssertFunction(callback)
		if !ok {
			panic(f.runtime.vm.ToValue("TypeError: callback must be a function"))
		}

		go func() {
			fmt.Println("Starting file read operation")
			data, err := os.ReadFile(filename)

			fmt.Println("File read complete, queueing callback")
			f.runtime.queue <- func() {
				if err != nil {
					_, _ = fn(goja.Undefined(), f.runtime.vm.ToValue(err.Error()), goja.Undefined())
				} else {
					_, _ = fn(goja.Undefined(), goja.Undefined(), f.runtime.vm.ToValue(string(data)))
				}
			}
		}()
		return goja.Undefined()
	})

	f.runtime.vm.Set("writeFile", func(call goja.FunctionCall) goja.Value {
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
			panic(f.runtime.vm.ToValue("TypeError: callback must be a function"))
		}

		go func() {
			file, err := os.OpenFile(filename, flag, fileMode)
			if err != nil {
				panic(err)
			}
			defer file.Close()
			_, err = file.Write([]byte(data))
			if err != nil {
				f.runtime.queue <- func() {
					_, _ = fn(goja.Undefined(), f.runtime.vm.ToValue(err.Error()), goja.Undefined())
				}
			} else {
				f.runtime.queue <- func() {
					_, _ = fn(goja.Undefined(), goja.Undefined(), f.runtime.vm.ToValue(filename))
				}
			}
		}()
		return goja.Undefined()
	})

}
