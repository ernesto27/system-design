package runtimejs

import (
	"runtime"

	"github.com/dop251/goja"
)

type OS struct {
	runtime *RuntimeJS
}

func (o *OS) SetGlobals() {
	o.runtime.vm.Set("platform", func(call goja.FunctionCall) goja.Value {
		return o.runtime.vm.ToValue(runtime.GOOS)
	})

	o.runtime.vm.Set("arch", func(call goja.FunctionCall) goja.Value {
		return o.runtime.vm.ToValue(runtime.GOARCH)
	})
}
