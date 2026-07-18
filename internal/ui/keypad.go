package ui

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/richardwooding/chipdeck/internal/chip8"
)

// keyMap is the standard COSMAC-to-QWERTY mapping:
//
//	1 2 3 C        1 2 3 4
//	4 5 6 D   <-   Q W E R
//	7 8 9 E        A S D F
//	A 0 B F        Z X C V
var keyMap = map[ebiten.Key]byte{
	ebiten.KeyDigit1: 0x1, ebiten.KeyDigit2: 0x2, ebiten.KeyDigit3: 0x3, ebiten.KeyDigit4: 0xC,
	ebiten.KeyQ: 0x4, ebiten.KeyW: 0x5, ebiten.KeyE: 0x6, ebiten.KeyR: 0xD,
	ebiten.KeyA: 0x7, ebiten.KeyS: 0x8, ebiten.KeyD: 0x9, ebiten.KeyF: 0xE,
	ebiten.KeyZ: 0xA, ebiten.KeyX: 0x0, ebiten.KeyC: 0xB, ebiten.KeyV: 0xF,
}

// padLayout is the on-screen 4x4 grid in COSMAC order.
var padLayout = [16]byte{
	0x1, 0x2, 0x3, 0xC,
	0x4, 0x5, 0x6, 0xD,
	0x7, 0x8, 0x9, 0xE,
	0xA, 0x0, 0xB, 0xF,
}

// keypad feeds keyboard and touch input into the machine and draws the
// on-screen pad (which doubles as the mobile input).
type keypad struct {
	x, y, cell float64
	touchHeld  map[ebiten.TouchID]byte
}

func newKeypad(x, y, cell float64) *keypad {
	return &keypad{x: x, y: y, cell: cell, touchHeld: map[ebiten.TouchID]byte{}}
}

func (k *keypad) cellRect(i int) (x0, y0, x1, y1 float64) {
	const gap = 4
	col, row := i%4, i/4
	x0 = k.x + float64(col)*(k.cell+gap)
	y0 = k.y + float64(row)*(k.cell+gap)
	return x0, y0, x0 + k.cell, y0 + k.cell
}

func (k *keypad) keyAt(p image.Point) (byte, bool) {
	for i, key := range padLayout {
		x0, y0, x1, y1 := k.cellRect(i)
		if float64(p.X) >= x0 && float64(p.X) < x1 && float64(p.Y) >= y0 && float64(p.Y) < y1 {
			return key, true
		}
	}
	return 0, false
}

// Update pushes current input state into the machine.
func (k *keypad) Update(m *chip8.Machine) {
	var down [16]bool
	for key, hex := range keyMap {
		if ebiten.IsKeyPressed(key) {
			down[hex] = true
		}
	}
	// touches: track which pad cell each active touch is on
	for id := range k.touchHeld {
		delete(k.touchHeld, id)
	}
	for _, id := range ebiten.AppendTouchIDs(nil) {
		x, y := ebiten.TouchPosition(id)
		if hex, ok := k.keyAt(image.Pt(x, y)); ok {
			k.touchHeld[id] = hex
			down[hex] = true
		}
	}
	// mouse press on the pad (desktop clicking)
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if hex, ok := k.keyAt(image.Pt(x, y)); ok {
			down[hex] = true
		}
	}
	for hex, d := range down {
		m.SetKey(byte(hex), d)
	}
}

// Draw renders the pad, lighting pressed keys.
func (k *keypad) Draw(dst *ebiten.Image, m *chip8.Machine) {
	for i, key := range padLayout {
		x0, y0, _, _ := k.cellRect(i)
		fill := colPanel
		txt := colDim
		if m.Keys[key] {
			fill = colAccentDim
			txt = colText
		}
		vector.FillRect(dst, float32(x0), float32(y0), float32(k.cell), float32(k.cell), fill, false)
		vector.StrokeRect(dst, float32(x0), float32(y0), float32(k.cell), float32(k.cell), 1, colPanelEdge, false)
		label := fmt.Sprintf("%X", key)
		drawText(dst, label, x0+(k.cell-textWidth(label, 1))/2, y0+(k.cell-glyphH)/2, txt, 1)
	}
}
