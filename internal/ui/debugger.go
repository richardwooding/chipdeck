package ui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/richardwooding/chipdeck/internal/chip8"
)

// debugger renders the live machine-state panel: registers, stack, a
// disassembly window around PC, and a recent-PC trace.
type debugger struct {
	trace []uint16 // ring of recent PCs
}

func (d *debugger) Record(pc uint16) {
	d.trace = append(d.trace, pc)
	if len(d.trace) > 8 {
		d.trace = d.trace[len(d.trace)-8:]
	}
}

func (d *debugger) Draw(dst *ebiten.Image, m *chip8.Machine, x, y, w, h float64) {
	vector.FillRect(dst, float32(x), float32(y), float32(w), float32(h), colPanel, false)
	vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 1, colPanelEdge, false)

	cx, cy := x+10, y+8
	line := func(s string, clr interface {
		RGBA() (uint32, uint32, uint32, uint32)
	}) { drawText(dst, s, cx, cy, clr, 1); cy += glyphH + 1 }

	drawText(dst, "REGISTERS", cx, cy, colDimmer, 1)
	cy += glyphH + 3
	for row := range 4 {
		s := ""
		for col := range 4 {
			i := row*4 + col
			s += fmt.Sprintf("V%X %02X  ", i, m.V[i])
		}
		line(s, colText)
	}
	cy += 4
	line(fmt.Sprintf("I  %03X   PC %03X   SP %X", m.I, m.PC, m.SP), colAccent)
	line(fmt.Sprintf("DT %02X    ST %02X", m.Delay, m.Sound), colDim)

	if m.SP > 0 {
		s := "STACK "
		for i := range int(m.SP) {
			s += fmt.Sprintf("%03X ", m.Stack[i])
		}
		line(s, colDim)
	} else {
		line("STACK empty", colDimmer)
	}

	cy += 6
	drawText(dst, "DISASSEMBLY", cx, cy, colDimmer, 1)
	cy += glyphH + 3
	start := int(m.PC) - 4
	if start < 0 {
		start = 0
	}
	for addr := start; addr < start+20 && addr+1 < chip8.MemSize; addr += 2 {
		op := uint16(m.Mem[addr])<<8 | uint16(m.Mem[addr+1])
		clr := colDim
		prefix := "  "
		if addr == int(m.PC) {
			clr = colAccent
			prefix = "> "
		}
		line(fmt.Sprintf("%s%03X  %04X  %s", prefix, addr, op, chip8.Disassemble(op)), clr)
	}

	cy += 6
	drawText(dst, "TRACE", cx, cy, colDimmer, 1)
	cy += glyphH + 3
	for i := len(d.trace) - 1; i >= 0; i-- {
		pc := d.trace[i]
		if int(pc)+1 >= chip8.MemSize {
			continue
		}
		op := uint16(m.Mem[pc])<<8 | uint16(m.Mem[pc+1])
		line(fmt.Sprintf("  %03X  %s", pc, chip8.Disassemble(op)), colDimmer)
	}
}
