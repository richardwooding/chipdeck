package chip8

// Quirks captures the behavioral differences between historical CHIP-8
// interpreters. The zero value is "modern lax"; DefaultQuirks returns the
// original COSMAC VIP behavior, which is what the Timendus quirks test
// expects in CHIP-8 mode.
type Quirks struct {
	VFReset     bool // 8xy1/8xy2/8xy3 zero vF afterwards
	MemoryIncrI bool // Fx55/Fx65 leave I = I + X + 1
	DisplayWait bool // Dxyn waits for the next vertical blank (max ~60 sprites/s)
	Clipping    bool // sprites clip at screen edges (origin still wraps)
	ShiftVX     bool // 8xy6/8xyE shift vX in place (false: shift vY into vX)
	JumpVX      bool // Bxnn jumps to xNN + vX (false: Bnnn jumps to nnn + v0)
}

// DefaultQuirks is original CHIP-8 (COSMAC VIP) behavior.
func DefaultQuirks() Quirks {
	return Quirks{
		VFReset:     true,
		MemoryIncrI: true,
		DisplayWait: true,
		Clipping:    true,
		ShiftVX:     false,
		JumpVX:      false,
	}
}
