package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/richardwooding/chipdeck/internal/chip8"
)

// screen renders the 64x32 framebuffer with optional phosphor persistence,
// which tames the XOR flicker inherent to CHIP-8 drawing.
type screen struct {
	img     *ebiten.Image
	pix     []byte    // RGBA staging buffer
	decay   []float32 // per-pixel phosphor level 0..1
	Persist bool
}

func newScreen() *screen {
	return &screen{
		img:     ebiten.NewImage(chip8.DisplayW, chip8.DisplayH),
		pix:     make([]byte, chip8.DisplayW*chip8.DisplayH*4),
		decay:   make([]float32, chip8.DisplayW*chip8.DisplayH),
		Persist: true,
	}
}

// Update folds the machine's framebuffer into the phosphor state.
func (s *screen) Update(m *chip8.Machine) {
	for i, p := range m.Display {
		v := float32(p)
		if s.Persist {
			if v < s.decay[i] {
				v = s.decay[i] * 0.55 // fast-ish decay: ghosts, not smears
			}
		}
		s.decay[i] = v

		on, off := colPixOn, colPixOff
		r := float32(off.R) + (float32(on.R)-float32(off.R))*v
		g := float32(off.G) + (float32(on.G)-float32(off.G))*v
		b := float32(off.B) + (float32(on.B)-float32(off.B))*v
		s.pix[i*4+0] = byte(r)
		s.pix[i*4+1] = byte(g)
		s.pix[i*4+2] = byte(b)
		s.pix[i*4+3] = 0xff
	}
	s.img.WritePixels(s.pix)
}

// Draw blits the screen scaled into the rect at (x, y, w, h) with a border.
func (s *screen) Draw(dst *ebiten.Image, x, y, w, h float64) {
	vector.StrokeRect(dst, float32(x)-2, float32(y)-2, float32(w)+4, float32(h)+4, 1, colPanelEdge, false)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(w/chip8.DisplayW, h/chip8.DisplayH)
	op.GeoM.Translate(x, y)
	dst.DrawImage(s.img, op)
}
