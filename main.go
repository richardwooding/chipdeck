// chipdeck is a CHIP-8 emulator with a live debugger: a pure-Go interpreter
// core (COSMAC VIP quirks, Timendus-suite verified) under an Ebitengine
// front end. One codebase runs native for development and WebAssembly in
// the browser.
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/richardwooding/chipdeck/internal/ui"
)

func main() {
	ebiten.SetWindowSize(ui.W, ui.H)
	ebiten.SetWindowTitle("chipdeck")
	if err := ebiten.RunGame(ui.NewGame()); err != nil {
		log.Fatal(err)
	}
}
