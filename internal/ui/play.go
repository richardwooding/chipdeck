package ui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/richardwooding/chipdeck/internal/chip8"
)

// playScene runs the machine: screen, keypad, debugger, transport keys.
type playScene struct {
	m        *chip8.Machine
	title    string
	controls string
	tickrate int // cycles per 60Hz frame

	screen  *screen
	pad     *keypad
	dbg     *debugger
	paused  bool
	showDbg bool
	err     error
	frame   int64
}

func newPlayScene(m *chip8.Machine, title, controls string, tickrate int) *playScene {
	return &playScene{
		m:        m,
		title:    title,
		controls: controls,
		tickrate: tickrate,
		screen:   newScreen(),
		pad:      newKeypad(24, 420, 36),
		dbg:      &debugger{},
		showDbg:  true,
	}
}

func (s *playScene) Update(g *Game) error {
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeySpace):
		s.paused = !s.paused
	case inpututil.IsKeyJustPressed(ebiten.KeyN):
		if s.paused && s.err == nil {
			s.dbg.Record(s.m.PC)
			if err := s.m.Step(); err != nil {
				s.err = err
			}
		}
	case inpututil.IsKeyJustPressed(ebiten.KeyB):
		s.m.Reset()
		s.err = nil
	case inpututil.IsKeyJustPressed(ebiten.KeyG):
		s.showDbg = !s.showDbg
	case inpututil.IsKeyJustPressed(ebiten.KeyP):
		s.screen.Persist = !s.screen.Persist
	case inpututil.IsKeyJustPressed(ebiten.KeyEqual), inpututil.IsKeyJustPressed(ebiten.KeyKPAdd):
		s.tickrate = min(s.tickrate*2, 2000)
	case inpututil.IsKeyJustPressed(ebiten.KeyMinus), inpututil.IsKeyJustPressed(ebiten.KeyKPSubtract):
		s.tickrate = max(s.tickrate/2, 1)
	case inpututil.IsKeyJustPressed(ebiten.KeyEscape):
		if b := g.beeper(); b != nil {
			b.Gate(false)
		}
		g.scene = newPickerScene()
		return nil
	}

	s.pad.Update(s.m)

	if !s.paused && s.err == nil {
		s.frame++
		s.dbg.Record(s.m.PC)
		s.m.TickTimers()
		if _, err := s.m.RunCycles(s.tickrate); err != nil {
			s.err = err
		}
	}
	if b := g.beeper(); b != nil {
		b.Gate(!s.paused && s.m.Beeping())
	}
	s.screen.Update(s.m)

	// New ROM dropped mid-game replaces the machine.
	if data, name, ok := takePickedFile(); ok {
		s.loadNew(data, name)
	} else if files := ebiten.DroppedFiles(); files != nil {
		if data, name, ok := firstFile(files); ok {
			s.loadNew(data, name)
		}
	}
	return nil
}

func (s *playScene) loadNew(data []byte, name string) {
	m := chip8.New(chip8.DefaultQuirks())
	if err := m.LoadROM(data); err != nil {
		s.err = err
		return
	}
	s.m = m
	s.title = name
	s.controls = "keys: 1234/QWER/ASDF/ZXCV"
	s.tickrate = 11
	s.err = nil
	s.screen = newScreen()
	s.dbg = &debugger{}
}

func (s *playScene) Draw(dst *ebiten.Image) {
	// Header
	drawText(dst, s.title, 24, 12, colText, 2)
	status := fmt.Sprintf("%d cyc/frame", s.tickrate)
	if s.paused {
		status += "  ⏸ PAUSED"
	}
	if s.m.Beeping() && !s.paused {
		status += "  ♪"
	}
	drawText(dst, status, 24, 40, colDim, 1)

	// Emulated screen: 640x320 at (24, 64)
	s.screen.Draw(dst, 24, 64, 640, 320)

	// Keypad below the screen
	s.pad.Draw(dst, s.m)

	// Controls hint next to the pad
	drawText(dst, s.controls, 210, 430, colDim, 1)
	help := "space pause · n step · b reset · g debugger · p phosphor · +/- speed · esc games"
	drawText(dst, help, 210, 450, colDimmer, 1)
	if s.err != nil {
		drawText(dst, fmt.Sprintf("halted: %v", s.err), 210, 476, colAmber, 1)
	}

	// Debugger panel on the right
	if s.showDbg {
		s.dbg.Draw(dst, s.m, 688, 12, W-688-12, H-24)
	}
}
