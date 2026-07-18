// Package ui is the Ebitengine front end: a ROM picker and the play scene
// with the emulated screen, touch keypad, and live debugger panel.
package ui

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

const (
	W = 960
	H = 600

	audioBufferSize = 50 * time.Millisecond
)

type scene interface {
	Update(g *Game) error
	Draw(dst *ebiten.Image)
}

// Game is the ebiten.Game; it owns the audio context and current scene.
type Game struct {
	audioCtx *audio.Context
	beep     *beeper
	scene    scene
}

func NewGame() *Game {
	g := &Game{audioCtx: audio.NewContext(beepRate)}
	g.scene = newBootScene()
	return g
}

// beeper lazily starts on first use so the browser audio context has had a
// user gesture by then.
func (g *Game) beeper() *beeper {
	if g.beep == nil {
		if b, err := newBeeper(g.audioCtx); err == nil {
			g.beep = b
		}
	}
	return g.beep
}

func (g *Game) Update() error              { return g.scene.Update(g) }
func (g *Game) Draw(dst *ebiten.Image)     { dst.Fill(colBG); g.scene.Draw(dst) }
func (g *Game) Layout(_, _ int) (int, int) { return W, H }
