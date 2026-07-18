// Package chip8 is a pure-Go CHIP-8 interpreter core: no rendering, no
// audio, no dependencies. The front end drives it by calling TickTimers at
// 60Hz and RunCycles for the CPU budget per frame, then reads Display, and
// feeds SetKey. Quirks reproduce original COSMAC VIP behavior by default.
package chip8

import (
	"fmt"
	"math/rand/v2"
)

const (
	DisplayW     = 64
	DisplayH     = 32
	MemSize      = 4096
	ProgramStart = 0x200
)

// Machine is one CHIP-8 interpreter instance.
//
// It is deliberately a value type with no interior pointers to mutable
// state: assigning one Machine to another is a complete snapshot (the rom
// slice and Rand override are shared but immutable), which is what makes
// History's rewind-by-replay exact.
type Machine struct {
	Mem     [MemSize]byte
	V       [16]byte
	I, PC   uint16
	SP      byte
	Stack   [16]uint16
	Delay   byte
	Sound   byte
	Keys    [16]bool
	Display [DisplayW * DisplayH]byte // 0 or 1 per pixel
	Quirks  Quirks
	Rand    func() byte // test override; nil uses the built-in seeded rng
	Cycles  uint64      // instructions executed since Reset — the rewind timeline

	rng uint64 // xorshift64* state; part of the snapshot, so replays match
	rom []byte

	// Fx0A state: original interpreters register a key on RELEASE.
	waitReg     int // register awaiting a key, -1 when not waiting
	waitPressed int // key seen down while waiting, -1 until one is pressed

	drewThisFrame bool // display-wait gate; cleared by TickTimers
}

// New returns a machine with the font loaded and PC at the program start.
func New(q Quirks) *Machine {
	m := &Machine{Quirks: q}
	m.Reset()
	return m
}

// LoadROM installs a program and resets the machine.
func (m *Machine) LoadROM(rom []byte) error {
	if len(rom) == 0 {
		return fmt.Errorf("empty ROM")
	}
	if len(rom) > MemSize-ProgramStart {
		return fmt.Errorf("ROM too large: %d bytes (max %d)", len(rom), MemSize-ProgramStart)
	}
	m.rom = append([]byte(nil), rom...)
	m.Reset()
	return nil
}

// Reset returns the machine to power-on state, keeping the loaded ROM. The
// rng is freshly seeded — a new run gets new randomness, while snapshots
// taken after Reset stay internally consistent.
func (m *Machine) Reset() {
	*m = Machine{Quirks: m.Quirks, Rand: m.Rand, rom: m.rom, rng: rand.Uint64() | 1}
	copy(m.Mem[FontStart:], font[:])
	copy(m.Mem[ProgramStart:], m.rom)
	m.PC = ProgramStart
	m.waitReg = -1
	m.waitPressed = -1
}

// randByte draws from the test override when set, else from the machine's
// own xorshift64* state (which snapshots carry, keeping replays exact).
func (m *Machine) randByte() byte {
	if m.Rand != nil {
		return m.Rand()
	}
	m.rng ^= m.rng << 13
	m.rng ^= m.rng >> 7
	m.rng ^= m.rng << 17
	return byte(m.rng * 0x2545F4914F6CDD1D >> 56)
}

// Step fetches, decodes, and executes one instruction.
func (m *Machine) Step() error {
	if int(m.PC)+1 >= MemSize {
		return fmt.Errorf("PC out of range: %#04x", m.PC)
	}
	op := uint16(m.Mem[m.PC])<<8 | uint16(m.Mem[m.PC+1])
	m.PC += 2
	if err := m.execute(op); err != nil {
		return err
	}
	m.Cycles++
	return nil
}

// Blocked reports whether the machine is parked on Fx0A waiting for a key.
func (m *Machine) Blocked() bool { return m.waitReg >= 0 }

// RunCycles executes up to n instructions, stopping early when blocked on
// Fx0A or when the display-wait quirk gates after a draw. It returns the
// number of instructions executed.
func (m *Machine) RunCycles(n int) (int, error) {
	for i := range n {
		if m.Blocked() {
			// Progress the key-wait state machine without burning cycles.
			if err := m.Step(); err != nil {
				return i, err
			}
			if m.Blocked() {
				return i, nil
			}
			continue
		}
		if m.Quirks.DisplayWait && m.drewThisFrame {
			return i, nil
		}
		if err := m.Step(); err != nil {
			return i, err
		}
	}
	return n, nil
}

// TickTimers advances the 60Hz timers and opens the display-wait gate.
func (m *Machine) TickTimers() {
	if m.Delay > 0 {
		m.Delay--
	}
	if m.Sound > 0 {
		m.Sound--
	}
	m.drewThisFrame = false
}

// SetKey updates one hex key (0-F).
func (m *Machine) SetKey(k byte, down bool) {
	if k < 16 {
		m.Keys[k] = down
	}
}

// Beeping reports whether the buzzer should sound.
func (m *Machine) Beeping() bool { return m.Sound > 0 }

// DisplayHash returns an FNV-1a hash of the framebuffer, for golden tests.
func (m *Machine) DisplayHash() uint64 {
	h := uint64(14695981039346656037)
	for _, p := range m.Display {
		h ^= uint64(p)
		h *= 1099511628211
	}
	return h
}

// DisplayASCII renders the framebuffer as text, for humans and test output.
func (m *Machine) DisplayASCII() string {
	buf := make([]byte, 0, (DisplayW+1)*DisplayH*3)
	for y := range DisplayH {
		for x := range DisplayW {
			if m.Display[y*DisplayW+x] != 0 {
				buf = append(buf, "█"...)
			} else {
				buf = append(buf, "·"...)
			}
		}
		buf = append(buf, '\n')
	}
	return string(buf)
}
