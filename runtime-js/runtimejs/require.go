package runtimejs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
)

type Require struct {
	runtime *RuntimeJS
	modules map[string]interface{}
}

func (r *Require) SetGlobals() {
	r.modules = make(map[string]interface{})
	r.runtime.vm.Set("require", func(call goja.FunctionCall) goja.Value {
		moduleName := call.Argument(0).String()

		if _, ok := r.modules[moduleName]; ok {
			fmt.Println("call module from cache " + moduleName)
			return r.runtime.vm.ToValue(r.modules[moduleName])
		}

		fmt.Println("require called for module:", call.Argument(0).String())
		content, err := r.readModuleFile(moduleName)
		if err != nil {
			fmt.Println("Error reading module file:", err)
			return r.runtime.vm.ToValue(nil)
		}

		moduleRuntime := goja.New()

		moduleObj := moduleRuntime.NewObject()
		moduleRuntime.Set("module", moduleObj)
		moduleRuntime.Set("exports", moduleObj.Get("exports"))

		_, err = moduleRuntime.RunString(content)
		if err != nil {
			fmt.Println("Error executing module code:", err)
			return r.runtime.vm.ToValue(nil)
		}
		exports := moduleObj.Get("exports")

		exportValue := exports.Export()
		r.modules[moduleName] = exportValue

		return r.runtime.vm.ToValue(exportValue)
	})
}

func (r *Require) readModuleFile(moduleName string) (string, error) {
	if strings.HasPrefix(moduleName, "./") || strings.HasPrefix(moduleName, "../") {
		baseDir := filepath.Dir(moduleName)
		content, err := os.ReadFile(baseDir + "/" + moduleName)
		if err != nil {
			return "", err
		}
		return string(content), nil
	}

	return "", fmt.Errorf("module not found: %s", moduleName)
}
