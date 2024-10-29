package main

import (
	"syscall/js"

	"github.com/rilldata/rill/runtime/pkg/rilltime"
)

func main() {
	js.Global().Set("parseRillTime", parseRillTime())
	<-make(chan struct{})
}

func parseRillTime() js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 1 {
			return "Invalid no of arguments passed"
		}
		rt, err := rilltime.Parse(args[0].String())
		if err != nil {
			return err.Error()
		}
		return rt.String()
	})
	return jsonFunc
}
