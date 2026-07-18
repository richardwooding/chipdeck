//go:build js && wasm

package ui

import (
	"sync"
	"syscall/js"
)

var (
	pickedMu   sync.Mutex
	pickedData []byte
	pickedName string
)

func canPickFiles() bool { return true }

// openFilePicker shows the browser's file chooser (must follow a user
// gesture; Ebiten handles taps on the next frame, within the window).
func openFilePicker() {
	doc := js.Global().Get("document")
	input := doc.Call("createElement", "input")
	input.Set("type", "file")
	input.Set("accept", ".ch8,.c8,.rom")

	var onChange js.Func
	onChange = js.FuncOf(func(this js.Value, _ []js.Value) any {
		defer onChange.Release()
		files := input.Get("files")
		if files.Get("length").Int() == 0 {
			return nil
		}
		file := files.Call("item", 0)
		name := file.Get("name").String()
		var then js.Func
		then = js.FuncOf(func(this js.Value, args []js.Value) any {
			defer then.Release()
			u8 := js.Global().Get("Uint8Array").New(args[0])
			data := make([]byte, u8.Get("length").Int())
			js.CopyBytesToGo(data, u8)
			pickedMu.Lock()
			pickedData, pickedName = data, name
			pickedMu.Unlock()
			return nil
		})
		file.Call("arrayBuffer").Call("then", then)
		return nil
	})
	input.Call("addEventListener", "change", onChange)
	input.Call("click")
}

func takePickedFile() ([]byte, string, bool) {
	pickedMu.Lock()
	defer pickedMu.Unlock()
	if pickedData == nil {
		return nil, "", false
	}
	d, n := pickedData, pickedName
	pickedData, pickedName = nil, ""
	return d, n, true
}
