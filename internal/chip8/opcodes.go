package chip8

import "fmt"

// execute runs one decoded instruction. vF is always written last in the
// 8xxx arithmetic group so that vF-as-operand behaves like the original
// interpreter (the Timendus flags test checks exactly this).
func (m *Machine) execute(op uint16) error {
	x := int(op >> 8 & 0xF)
	y := int(op >> 4 & 0xF)
	nn := byte(op)
	nnn := op & 0x0FFF

	switch op >> 12 {
	case 0x0:
		return m.executeSys(op)
	case 0x1: // JP nnn
		m.PC = nnn
	case 0x2: // CALL nnn
		return m.call(nnn)
	case 0x3: // SE Vx, nn
		m.skipIf(m.V[x] == nn)
	case 0x4: // SNE Vx, nn
		m.skipIf(m.V[x] != nn)
	case 0x5: // SE Vx, Vy
		if op&0xF != 0 {
			return fmt.Errorf("unknown opcode %#04x", op)
		}
		m.skipIf(m.V[x] == m.V[y])
	case 0x6: // LD Vx, nn
		m.V[x] = nn
	case 0x7: // ADD Vx, nn (no carry)
		m.V[x] += nn
	case 0x8:
		return m.executeALU(op, x, y)
	case 0x9: // SNE Vx, Vy
		if op&0xF != 0 {
			return fmt.Errorf("unknown opcode %#04x", op)
		}
		m.skipIf(m.V[x] != m.V[y])
	case 0xA: // LD I, nnn
		m.I = nnn
	case 0xB: // JP v0+nnn / Bxnn quirk: xnn+vX
		m.PC = nnn + uint16(m.jumpOffset(x))
	case 0xC: // RND Vx, nn
		m.V[x] = m.randByte() & nn
	case 0xD:
		m.draw(x, y, int(op&0xF))
	case 0xE:
		return m.executeKeys(op, x, nn)
	case 0xF:
		return m.executeMisc(op, x, nn)
	}
	return nil
}

// executeSys handles the 0x0 group: CLS, RET, and ignored machine calls.
func (m *Machine) executeSys(op uint16) error {
	switch op {
	case 0x00E0: // CLS
		m.Display = [DisplayW * DisplayH]byte{}
	case 0x00EE: // RET
		if m.SP == 0 {
			return fmt.Errorf("stack underflow at %#04x", m.PC-2)
		}
		m.SP--
		m.PC = m.Stack[m.SP]
	default:
		// 0nnn machine-code call: ignored, like most interpreters.
	}
	return nil
}

// call pushes the return address and jumps.
func (m *Machine) call(nnn uint16) error {
	if int(m.SP) >= len(m.Stack) {
		return fmt.Errorf("stack overflow at %#04x", m.PC-2)
	}
	m.Stack[m.SP] = m.PC
	m.SP++
	m.PC = nnn
	return nil
}

// skipIf advances past the next instruction when cond holds.
func (m *Machine) skipIf(cond bool) {
	if cond {
		m.PC += 2
	}
}

// jumpOffset is Bnnn's register operand: v0 classically, vX under the quirk.
func (m *Machine) jumpOffset(x int) byte {
	if m.Quirks.JumpVX {
		return m.V[x]
	}
	return m.V[0]
}

// executeKeys handles Ex9E/ExA1 keyboard skips.
func (m *Machine) executeKeys(op uint16, x int, nn byte) error {
	switch nn {
	case 0x9E: // SKP Vx
		m.skipIf(m.Keys[m.V[x]&0xF])
	case 0xA1: // SKNP Vx
		m.skipIf(!m.Keys[m.V[x]&0xF])
	default:
		return fmt.Errorf("unknown opcode %#04x", op)
	}
	return nil
}

func (m *Machine) executeALU(op uint16, x, y int) error {
	var flag byte
	hasFlag := false
	switch op & 0xF {
	case 0x0: // LD Vx, Vy
		m.V[x] = m.V[y]
	case 0x1: // OR
		m.V[x] |= m.V[y]
		if m.Quirks.VFReset {
			m.V[0xF] = 0
		}
	case 0x2: // AND
		m.V[x] &= m.V[y]
		if m.Quirks.VFReset {
			m.V[0xF] = 0
		}
	case 0x3: // XOR
		m.V[x] ^= m.V[y]
		if m.Quirks.VFReset {
			m.V[0xF] = 0
		}
	case 0x4: // ADD with carry
		sum := uint16(m.V[x]) + uint16(m.V[y])
		m.V[x] = byte(sum)
		flag, hasFlag = byte(sum>>8), true
	case 0x5: // SUB Vx -= Vy, vF = NOT borrow
		notBorrow := byte(0)
		if m.V[x] >= m.V[y] {
			notBorrow = 1
		}
		m.V[x] -= m.V[y]
		flag, hasFlag = notBorrow, true
	case 0x6: // SHR
		if !m.Quirks.ShiftVX {
			m.V[x] = m.V[y]
		}
		flag, hasFlag = m.V[x]&1, true
		m.V[x] >>= 1
	case 0x7: // SUBN Vx = Vy - Vx
		notBorrow := byte(0)
		if m.V[y] >= m.V[x] {
			notBorrow = 1
		}
		m.V[x] = m.V[y] - m.V[x]
		flag, hasFlag = notBorrow, true
	case 0xE: // SHL
		if !m.Quirks.ShiftVX {
			m.V[x] = m.V[y]
		}
		flag, hasFlag = m.V[x]>>7, true
		m.V[x] <<= 1
	default:
		return fmt.Errorf("unknown opcode %#04x", op)
	}
	if hasFlag {
		m.V[0xF] = flag // written last: vF-as-operand keeps original semantics
	}
	return nil
}

func (m *Machine) executeMisc(op uint16, x int, nn byte) error {
	switch nn {
	case 0x07: // LD Vx, DT
		m.V[x] = m.Delay
	case 0x0A: // LD Vx, K — waits for a key press then RELEASE (VIP semantics)
		if m.waitReg < 0 {
			m.waitReg = x
			m.waitPressed = -1
		}
		if m.waitPressed < 0 {
			for k := range 16 {
				if m.Keys[k] {
					m.waitPressed = k
					break
				}
			}
		} else if !m.Keys[m.waitPressed] {
			m.V[m.waitReg] = byte(m.waitPressed)
			m.waitReg = -1
			m.waitPressed = -1
			return nil // completed: PC stays advanced
		}
		m.PC -= 2 // still waiting: re-execute this instruction
	case 0x15: // LD DT, Vx
		m.Delay = m.V[x]
	case 0x18: // LD ST, Vx
		m.Sound = m.V[x]
	case 0x1E: // ADD I, Vx
		m.I += uint16(m.V[x])
	case 0x29: // LD F, Vx — font sprite address
		m.I = FontStart + uint16(m.V[x]&0xF)*5
	case 0x33: // BCD
		if int(m.I)+2 >= MemSize {
			return fmt.Errorf("BCD write out of range: I=%#04x", m.I)
		}
		v := m.V[x]
		m.Mem[m.I] = v / 100
		m.Mem[m.I+1] = v / 10 % 10
		m.Mem[m.I+2] = v % 10
	case 0x55: // LD [I], V0..Vx
		if int(m.I)+x >= MemSize {
			return fmt.Errorf("register store out of range: I=%#04x", m.I)
		}
		copy(m.Mem[m.I:], m.V[:x+1])
		if m.Quirks.MemoryIncrI {
			m.I += uint16(x) + 1
		}
	case 0x65: // LD V0..Vx, [I]
		if int(m.I)+x >= MemSize {
			return fmt.Errorf("register load out of range: I=%#04x", m.I)
		}
		copy(m.V[:x+1], m.Mem[m.I:])
		if m.Quirks.MemoryIncrI {
			m.I += uint16(x) + 1
		}
	default:
		return fmt.Errorf("unknown opcode %#04x", op)
	}
	return nil
}

// draw XORs an n-row sprite at (vX, vY). The origin wraps; the body clips or
// wraps per the quirk. vF reports collision.
func (m *Machine) draw(x, y, n int) {
	x0 := int(m.V[x]) % DisplayW
	y0 := int(m.V[y]) % DisplayH
	m.V[0xF] = 0
	for r := range n {
		addr := int(m.I) + r
		if addr >= MemSize {
			break
		}
		py := y0 + r
		if py >= DisplayH {
			if m.Quirks.Clipping {
				continue
			}
			py %= DisplayH
		}
		b := m.Mem[addr]
		for c := range 8 {
			if b&(0x80>>c) == 0 {
				continue
			}
			px := x0 + c
			if px >= DisplayW {
				if m.Quirks.Clipping {
					continue
				}
				px %= DisplayW
			}
			idx := py*DisplayW + px
			if m.Display[idx] != 0 {
				m.V[0xF] = 1
			}
			m.Display[idx] ^= 1
		}
	}
	m.drewThisFrame = true
}
