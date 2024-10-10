package runtimejs

import (
	"os"
	"strings"

	"github.com/dop251/goja"
)

type Env struct {
	runtime *RuntimeJS
}

func (e *Env) SetGlobals() {
	e.runtime.vm.Set("process", e.runtime.vm.NewObject())
	process := e.runtime.vm.Get("process").(*goja.Object)
	process.Set("env", e.runtime.vm.NewObject())
	env := process.Get("env").(*goja.Object)

	for _, envVar := range os.Environ() {
		pair := strings.SplitN(envVar, "=", 2)
		if len(pair) == 2 && pair[0] != "" && pair[1] != "" {
			env.Set(pair[0], pair[1])
		}
	}
}
