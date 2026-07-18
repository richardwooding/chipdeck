package chip8

import "testing"

// exec loads one instruction at 0x200 and steps once.
func exec(t *testing.T, q Quirks, op uint16, setup func(*Machine)) *Machine {
	t.Helper()
	m := New(q)
	m.Mem[ProgramStart] = byte(op >> 8)
	m.Mem[ProgramStart+1] = byte(op)
	if setup != nil {
		setup(m)
	}
	if err := m.Step(); err != nil {
		t.Fatalf("Step(%04X): %v", op, err)
	}
	return m
}

func TestBasicOps(t *testing.T) {
	q := DefaultQuirks()

	t.Run("6xnn LD", func(t *testing.T) {
		m := exec(t, q, 0x6A42, nil)
		if m.V[0xA] != 0x42 {
			t.Errorf("VA = %02X, want 42", m.V[0xA])
		}
	})
	t.Run("7xnn ADD no carry", func(t *testing.T) {
		m := exec(t, q, 0x70FF, func(m *Machine) { m.V[0] = 2; m.V[0xF] = 7 })
		if m.V[0] != 1 {
			t.Errorf("V0 = %02X, want 01 (wrap)", m.V[0])
		}
		if m.V[0xF] != 7 {
			t.Errorf("VF = %02X; 7xnn must not touch VF", m.V[0xF])
		}
	})
	t.Run("1nnn JP", func(t *testing.T) {
		m := exec(t, q, 0x1234, nil)
		if m.PC != 0x234 {
			t.Errorf("PC = %03X, want 234", m.PC)
		}
	})
	t.Run("2nnn/00EE CALL and RET", func(t *testing.T) {
		m := exec(t, q, 0x2400, nil)
		if m.PC != 0x400 || m.SP != 1 || m.Stack[0] != 0x202 {
			t.Fatalf("after CALL: PC=%03X SP=%d Stack0=%03X", m.PC, m.SP, m.Stack[0])
		}
		m.Mem[0x400] = 0x00
		m.Mem[0x401] = 0xEE
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
		if m.PC != 0x202 || m.SP != 0 {
			t.Errorf("after RET: PC=%03X SP=%d", m.PC, m.SP)
		}
	})
	t.Run("stack underflow errors", func(t *testing.T) {
		m := New(q)
		m.Mem[ProgramStart] = 0x00
		m.Mem[ProgramStart+1] = 0xEE
		if err := m.Step(); err == nil {
			t.Error("RET on empty stack should error")
		}
	})
	t.Run("skips", func(t *testing.T) {
		for _, tc := range []struct {
			op    uint16
			setup func(*Machine)
			skip  bool
		}{
			{0x3042, func(m *Machine) { m.V[0] = 0x42 }, true},
			{0x3042, func(m *Machine) { m.V[0] = 0x41 }, false},
			{0x4042, func(m *Machine) { m.V[0] = 0x41 }, true},
			{0x4042, func(m *Machine) { m.V[0] = 0x42 }, false},
			{0x5010, func(m *Machine) { m.V[0], m.V[1] = 7, 7 }, true},
			{0x5010, func(m *Machine) { m.V[0], m.V[1] = 7, 8 }, false},
			{0x9010, func(m *Machine) { m.V[0], m.V[1] = 7, 8 }, true},
			{0x9010, func(m *Machine) { m.V[0], m.V[1] = 7, 7 }, false},
		} {
			m := exec(t, q, tc.op, tc.setup)
			want := uint16(0x202)
			if tc.skip {
				want = 0x204
			}
			if m.PC != want {
				t.Errorf("op %04X: PC = %03X, want %03X", tc.op, m.PC, want)
			}
		}
	})
	t.Run("Annn LD I", func(t *testing.T) {
		if m := exec(t, q, 0xA123, nil); m.I != 0x123 {
			t.Errorf("I = %03X", m.I)
		}
	})
	t.Run("Cxnn RND masks", func(t *testing.T) {
		m := exec(t, q, 0xC00F, func(m *Machine) { m.Rand = func() byte { return 0xAB } })
		if m.V[0] != 0x0B {
			t.Errorf("V0 = %02X, want 0B", m.V[0])
		}
	})
}

func TestALUFlags(t *testing.T) {
	q := DefaultQuirks()

	t.Run("8xy4 carry", func(t *testing.T) {
		m := exec(t, q, 0x8014, func(m *Machine) { m.V[0], m.V[1] = 0xFF, 0x02 })
		if m.V[0] != 0x01 || m.V[0xF] != 1 {
			t.Errorf("V0=%02X VF=%d, want 01/1", m.V[0], m.V[0xF])
		}
		m = exec(t, q, 0x8014, func(m *Machine) { m.V[0], m.V[1] = 1, 2; m.V[0xF] = 1 })
		if m.V[0] != 3 || m.V[0xF] != 0 {
			t.Errorf("V0=%02X VF=%d, want 03/0", m.V[0], m.V[0xF])
		}
	})
	t.Run("8xy4 vF as operand: flag written last", func(t *testing.T) {
		// VF += VF with carry: result must be the flag, not the sum.
		m := exec(t, q, 0x8FF4, func(m *Machine) { m.V[0xF] = 0xFF })
		if m.V[0xF] != 1 {
			t.Errorf("VF = %02X, want 1 (carry overwrites result)", m.V[0xF])
		}
	})
	t.Run("8xy5 not-borrow", func(t *testing.T) {
		m := exec(t, q, 0x8015, func(m *Machine) { m.V[0], m.V[1] = 5, 3 })
		if m.V[0] != 2 || m.V[0xF] != 1 {
			t.Errorf("V0=%02X VF=%d, want 02/1", m.V[0], m.V[0xF])
		}
		m = exec(t, q, 0x8015, func(m *Machine) { m.V[0], m.V[1] = 3, 5 })
		if m.V[0] != 0xFE || m.V[0xF] != 0 {
			t.Errorf("V0=%02X VF=%d, want FE/0", m.V[0], m.V[0xF])
		}
	})
	t.Run("8xy7 subn", func(t *testing.T) {
		m := exec(t, q, 0x8017, func(m *Machine) { m.V[0], m.V[1] = 3, 5 })
		if m.V[0] != 2 || m.V[0xF] != 1 {
			t.Errorf("V0=%02X VF=%d, want 02/1", m.V[0], m.V[0xF])
		}
	})
}

func TestQuirkSensitiveOps(t *testing.T) {
	vip := DefaultQuirks()
	modern := Quirks{ShiftVX: true, JumpVX: true} // everything else off

	t.Run("8xy1 OR resets VF on VIP", func(t *testing.T) {
		m := exec(t, vip, 0x8011, func(m *Machine) { m.V[0], m.V[1], m.V[0xF] = 0xF0, 0x0F, 7 })
		if m.V[0] != 0xFF || m.V[0xF] != 0 {
			t.Errorf("V0=%02X VF=%d, want FF/0", m.V[0], m.V[0xF])
		}
		m = exec(t, modern, 0x8011, func(m *Machine) { m.V[0], m.V[1], m.V[0xF] = 0xF0, 0x0F, 7 })
		if m.V[0xF] != 7 {
			t.Errorf("modern: VF=%d, want 7 (untouched)", m.V[0xF])
		}
	})
	t.Run("8xy6 SHR source register", func(t *testing.T) {
		m := exec(t, vip, 0x8016, func(m *Machine) { m.V[0], m.V[1] = 0xFF, 0x03 })
		if m.V[0] != 0x01 || m.V[0xF] != 1 {
			t.Errorf("VIP: V0=%02X VF=%d, want 01/1 (shifted VY)", m.V[0], m.V[0xF])
		}
		m = exec(t, modern, 0x8016, func(m *Machine) { m.V[0], m.V[1] = 0x03, 0xFF })
		if m.V[0] != 0x01 || m.V[0xF] != 1 {
			t.Errorf("modern: V0=%02X VF=%d, want 01/1 (shifted VX)", m.V[0], m.V[0xF])
		}
	})
	t.Run("8xyE SHL source register", func(t *testing.T) {
		m := exec(t, vip, 0x801E, func(m *Machine) { m.V[0], m.V[1] = 0x00, 0x81 })
		if m.V[0] != 0x02 || m.V[0xF] != 1 {
			t.Errorf("VIP: V0=%02X VF=%d, want 02/1", m.V[0], m.V[0xF])
		}
	})
	t.Run("Bnnn vs Bxnn", func(t *testing.T) {
		m := exec(t, vip, 0xB300, func(m *Machine) { m.V[0], m.V[3] = 0x10, 0x20 })
		if m.PC != 0x310 {
			t.Errorf("VIP: PC=%03X, want 310 (nnn+V0)", m.PC)
		}
		m = exec(t, modern, 0xB300, func(m *Machine) { m.V[0], m.V[3] = 0x10, 0x20 })
		if m.PC != 0x320 {
			t.Errorf("modern: PC=%03X, want 320 (xnn+VX)", m.PC)
		}
	})
	t.Run("Fx55 memory increment", func(t *testing.T) {
		m := exec(t, vip, 0xF255, func(m *Machine) {
			m.I = 0x300
			m.V[0], m.V[1], m.V[2] = 1, 2, 3
		})
		if m.Mem[0x300] != 1 || m.Mem[0x302] != 3 {
			t.Error("stored values wrong")
		}
		if m.I != 0x303 {
			t.Errorf("VIP: I=%03X, want 303", m.I)
		}
		m = exec(t, modern, 0xF255, func(m *Machine) { m.I = 0x300 })
		if m.I != 0x300 {
			t.Errorf("modern: I=%03X, want 300 (unchanged)", m.I)
		}
	})
	t.Run("Fx65 load + increment", func(t *testing.T) {
		m := exec(t, vip, 0xF165, func(m *Machine) {
			m.I = 0x300
			m.Mem[0x300], m.Mem[0x301] = 9, 8
		})
		if m.V[0] != 9 || m.V[1] != 8 || m.I != 0x302 {
			t.Errorf("V0=%d V1=%d I=%03X", m.V[0], m.V[1], m.I)
		}
	})
}

func TestDraw(t *testing.T) {
	vip := DefaultQuirks()
	wrap := vip
	wrap.Clipping = false

	sprite := func(m *Machine) {
		m.I = 0x300
		m.Mem[0x300] = 0b11000000
		m.Mem[0x301] = 0b11000000
	}

	t.Run("draw and collide", func(t *testing.T) {
		m := exec(t, vip, 0xD012, func(m *Machine) { sprite(m); m.V[0], m.V[1] = 0, 0 })
		if m.Display[0] != 1 || m.Display[1] != 1 || m.Display[DisplayW] != 1 {
			t.Fatal("sprite not drawn")
		}
		if m.V[0xF] != 0 {
			t.Errorf("VF=%d, want 0 (no collision)", m.V[0xF])
		}
		// Draw again in place: everything erases, collision reported.
		m.PC = ProgramStart
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
		if m.V[0xF] != 1 {
			t.Errorf("VF=%d, want 1 (collision)", m.V[0xF])
		}
		if m.Display[0] != 0 {
			t.Error("XOR should have erased the pixel")
		}
	})
	t.Run("origin wraps", func(t *testing.T) {
		m := exec(t, vip, 0xD012, func(m *Machine) { sprite(m); m.V[0], m.V[1] = 64, 32 })
		if m.Display[0] != 1 {
			t.Error("origin (64,32) should wrap to (0,0)")
		}
	})
	t.Run("body clips on VIP", func(t *testing.T) {
		m := exec(t, vip, 0xD012, func(m *Machine) { sprite(m); m.V[0], m.V[1] = 63, 31 })
		if m.Display[31*DisplayW+63] != 1 {
			t.Error("corner pixel should draw")
		}
		if m.Display[31*DisplayW] != 0 || m.Display[63] != 0 {
			t.Error("overflow pixels should clip, not wrap")
		}
	})
	t.Run("body wraps when clipping off", func(t *testing.T) {
		m := exec(t, wrap, 0xD012, func(m *Machine) { sprite(m); m.V[0], m.V[1] = 63, 31 })
		if m.Display[31*DisplayW] != 1 || m.Display[63] != 1 || m.Display[0] != 1 {
			t.Error("overflow pixels should wrap")
		}
	})
}

func TestKeysAndTimers(t *testing.T) {
	q := DefaultQuirks()

	t.Run("Ex9E/ExA1", func(t *testing.T) {
		m := exec(t, q, 0xE09E, func(m *Machine) { m.V[0] = 5; m.Keys[5] = true })
		if m.PC != 0x204 {
			t.Errorf("SKP with key down: PC=%03X, want 204", m.PC)
		}
		m = exec(t, q, 0xE0A1, func(m *Machine) { m.V[0] = 5; m.Keys[5] = true })
		if m.PC != 0x202 {
			t.Errorf("SKNP with key down: PC=%03X, want 202", m.PC)
		}
	})
	t.Run("Fx0A waits for release", func(t *testing.T) {
		m := New(q)
		m.Mem[ProgramStart] = 0xF3
		m.Mem[ProgramStart+1] = 0x0A
		// no key: blocked, PC parked
		for range 3 {
			if err := m.Step(); err != nil {
				t.Fatal(err)
			}
		}
		if !m.Blocked() || m.PC != ProgramStart {
			t.Fatalf("should be blocked at %03X, PC=%03X", ProgramStart, m.PC)
		}
		// key down: still blocked (release semantics)
		m.SetKey(7, true)
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
		if !m.Blocked() {
			t.Fatal("must wait for RELEASE, not press")
		}
		// release: completes
		m.SetKey(7, false)
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
		if m.Blocked() || m.V[3] != 7 || m.PC != ProgramStart+2 {
			t.Errorf("blocked=%v V3=%d PC=%03X", m.Blocked(), m.V[3], m.PC)
		}
	})
	t.Run("timers and beep", func(t *testing.T) {
		m := exec(t, q, 0xF018, func(m *Machine) { m.V[0] = 2 })
		if !m.Beeping() {
			t.Fatal("sound timer set: should beep")
		}
		m.TickTimers()
		m.TickTimers()
		if m.Beeping() {
			t.Error("sound timer expired: should stop")
		}
	})
	t.Run("Fx33 BCD", func(t *testing.T) {
		m := exec(t, q, 0xF033, func(m *Machine) { m.V[0] = 234; m.I = 0x300 })
		if m.Mem[0x300] != 2 || m.Mem[0x301] != 3 || m.Mem[0x302] != 4 {
			t.Errorf("BCD = %d %d %d", m.Mem[0x300], m.Mem[0x301], m.Mem[0x302])
		}
	})
	t.Run("Fx29 font address", func(t *testing.T) {
		m := exec(t, q, 0xF029, func(m *Machine) { m.V[0] = 0xA })
		if m.I != FontStart+10*5 {
			t.Errorf("I = %03X", m.I)
		}
	})
}

func TestDisplayWaitGate(t *testing.T) {
	m := New(DefaultQuirks())
	// program: draw, then spin
	rom := []byte{
		0xA3, 0x00, // LD I, 300
		0xD0, 0x11, // DRW V0,V0,1
		0xD0, 0x11, // DRW again
		0x12, 0x04, // JP 204
	}
	if err := m.LoadROM(rom); err != nil {
		t.Fatal(err)
	}
	n, err := m.RunCycles(100)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("ran %d cycles, want 2 (gated after first draw)", n)
	}
	m.TickTimers() // vblank opens the gate
	n, err = m.RunCycles(100)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("ran %d cycles after vblank, want 1 (second draw gates)", n)
	}
}

func TestRunCyclesStopsWhenBlocked(t *testing.T) {
	m := New(DefaultQuirks())
	if err := m.LoadROM([]byte{0xF0, 0x0A}); err != nil { // LD V0, K
		t.Fatal(err)
	}
	n, err := m.RunCycles(50)
	if err != nil {
		t.Fatal(err)
	}
	if n >= 50 || !m.Blocked() {
		t.Errorf("n=%d blocked=%v; RunCycles should park on Fx0A", n, m.Blocked())
	}
}
