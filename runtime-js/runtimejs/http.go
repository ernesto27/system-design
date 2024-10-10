package runtimejs

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dop251/goja"
)

type Http struct {
	runtime *RuntimeJS
}

func (h *Http) SetGlobals() {
	h.runtime.vm.Set("createServer", func(call goja.FunctionCall) goja.Value {
		// Extract the callback function
		callback := call.Argument(0)
		fn, ok := goja.AssertFunction(callback)
		if !ok {
			panic(h.runtime.vm.ToValue("TypeError: callback must be a function"))
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			jsReq := h.runtime.vm.NewObject()
			jsRes := h.runtime.vm.NewObject()

			jsReq.Set("method", r.Method)
			jsReq.Set("url", r.URL.Path)

			jsRes.Set("writeHead", func(call goja.FunctionCall) goja.Value {
				statusCode := int(call.Argument(0).ToInteger())
				headers := call.Argument(1).Export()

				fmt.Println("statusCode:", statusCode)
				fmt.Println("headers:", headers)

				if headersMap, ok := headers.(map[string]interface{}); ok {
					for key, value := range headersMap {
						w.Header().Set(key, fmt.Sprint(value))
					}
				}

				w.WriteHeader(statusCode)
				return goja.Undefined()
			})

			jsRes.Set("end", func(call goja.FunctionCall) goja.Value {
				content := call.Argument(0).String()
				_, err := w.Write([]byte(content))
				if err != nil {
					http.Error(w, "Error writing response", http.StatusInternalServerError)
				}
				return goja.Undefined()
			})

			jsRes.Set("json", func(call goja.FunctionCall) goja.Value {
				jsonObj := call.Argument(0).Export()

				jsonData, err := json.Marshal(jsonObj)
				if err != nil {
					http.Error(w, "Error marshaling JSON", http.StatusInternalServerError)
					return goja.Undefined()
				}

				w.Header().Set("Content-Type", "application/json")
				_, err = w.Write(jsonData)
				if err != nil {
					http.Error(w, "Error writing JSON response", http.StatusInternalServerError)
				}
				return goja.Undefined()
			})

			// Call the JavaScript callback with request and response objects
			_, err := fn(goja.Undefined(), h.runtime.vm.ToValue(jsReq), h.runtime.vm.ToValue(jsRes))
			if err != nil {
				fmt.Println("Error calling callback:", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		})

		serverObj := h.runtime.vm.NewObject()
		serverObj.Set("listen", func(call goja.FunctionCall) goja.Value {
			port := call.Argument(0).ToInteger()
			host := call.Argument(1).String()

			// Start the server in a goroutine
			go func() {
				addr := fmt.Sprintf("%s:%d", host, port)
				fmt.Printf("Listening on %s\n", addr)
				// portString := strconv.Itoa(int(port))
				err := http.ListenAndServe(addr, handler)
				if err != nil {
					fmt.Printf("Server error: %v\n", err)
				}
			}()

			return goja.Undefined()
		})

		return serverObj
	})
}
