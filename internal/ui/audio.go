package ui

import (
	"sync/atomic"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const beepRate = 44100

// beeper is an endless square-wave source; Gate turns the tone on and off.
// The player pulls continuously (silence when gated off), which sidesteps
// start/stop latency and browser autoplay re-suspension.
type beeper struct {
	on    atomic.Bool
	phase int
}

// Read produces s16le stereo: a 440Hz square when on, silence otherwise.
func (b *beeper) Read(p []byte) (int, error) {
	const period = beepRate / 440
	n := len(p) / 4 * 4
	for i := 0; i < n; i += 4 {
		var sample int16
		if b.on.Load() {
			if b.phase < period/2 {
				sample = 6000
			} else {
				sample = -6000
			}
		}
		b.phase = (b.phase + 1) % period
		p[i] = byte(sample)
		p[i+1] = byte(sample >> 8)
		p[i+2] = byte(sample)
		p[i+3] = byte(sample >> 8)
	}
	return n, nil
}

// Gate sets whether the buzzer sounds.
func (b *beeper) Gate(on bool) { b.on.Store(on) }

// newBeeper wires the square wave into an audio player and starts it.
func newBeeper(ctx *audio.Context) (*beeper, error) {
	b := &beeper{}
	p, err := ctx.NewPlayer(b)
	if err != nil {
		return nil, err
	}
	p.SetBufferSize(audioBufferSize)
	p.Play()
	return b, nil
}
