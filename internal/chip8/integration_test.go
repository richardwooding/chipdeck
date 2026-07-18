package chip8

import (
	"os"
	"path/filepath"
	"testing"
)

// runROM loads a Timendus test ROM, applies memory pokes (the suite reads a
// selector byte at 0x1FF to skip its menus), then runs 60Hz frames until the
// display has been stable for 60 consecutive frames or maxFrames elapse.
func runROM(t *testing.T, name string, pokes map[uint16]byte, q Quirks, maxFrames int) *Machine {
	t.Helper()
	rom, err := os.ReadFile(filepath.Join("testdata", "timendus", name))
	if err != nil {
		t.Skipf("test ROM unavailable: %v", err)
	}
	m := New(q)
	m.Rand = func() byte { return 0x5A } // determinism for goldens
	if err := m.LoadROM(rom); err != nil {
		t.Fatal(err)
	}
	for addr, v := range pokes {
		m.Mem[addr] = v
	}
	stable := 0
	last := uint64(0)
	for range maxFrames {
		m.TickTimers()
		if _, err := m.RunCycles(11); err != nil {
			t.Fatalf("run: %v\n%s", err, m.DisplayASCII())
		}
		h := m.DisplayHash()
		if h == last {
			stable++
			if stable >= 60 {
				break
			}
		} else {
			stable = 0
			last = h
		}
	}
	return m
}

// goldens maps test ROM -> human-verified framebuffer hash. Regenerate with:
//
//	go test ./internal/chip8/ -run TestTimendus -v -chip8.dump
//
// then eyeball the ASCII art (all opcodes ticked, "all quirks correct") and
// update the constants.
var goldens = map[string]uint64{
	"1-chip8-logo": 10173898340286673873, // suite splash renders
	"2-ibm-logo":   1985184697246681613,  // classic striped IBM logo
	"3-corax+":     12080005269903082134, // all 24 opcode groups show the pass mark
	"4-flags":      15737658800505001742, // all carry/borrow groups show the pass mark
	"5-quirks":     5657534232238378441,  // all six quirks correct in CHIP-8 mode
}

func checkGolden(t *testing.T, name string, m *Machine) {
	t.Helper()
	if *dumpDisplays {
		t.Logf("%s display (hash %d):\n%s", name, m.DisplayHash(), m.DisplayASCII())
		return
	}
	want, ok := goldens[name]
	if !ok {
		t.Fatalf("no golden for %s — run with -chip8.dump and record it", name)
	}
	if got := m.DisplayHash(); got != want {
		t.Errorf("%s: display hash %d, want %d\n%s", name, got, want, m.DisplayASCII())
	}
}

func TestTimendusLogo(t *testing.T) {
	checkGolden(t, "1-chip8-logo", runROM(t, "1-chip8-logo.ch8", nil, DefaultQuirks(), 600))
}

func TestTimendusIBM(t *testing.T) {
	checkGolden(t, "2-ibm-logo", runROM(t, "2-ibm-logo.ch8", nil, DefaultQuirks(), 600))
}

func TestTimendusCorax(t *testing.T) {
	checkGolden(t, "3-corax+", runROM(t, "3-corax+.ch8", nil, DefaultQuirks(), 1200))
}

func TestTimendusFlags(t *testing.T) {
	checkGolden(t, "4-flags", runROM(t, "4-flags.ch8", nil, DefaultQuirks(), 1200))
}

func TestTimendusQuirks(t *testing.T) {
	// 0x1FF=1 auto-selects CHIP-8 mode, skipping the menu.
	m := runROM(t, "5-quirks.ch8", map[uint16]byte{0x1FF: 1}, DefaultQuirks(), 3600)
	checkGolden(t, "5-quirks", m)
}

func TestTimendusQuirksDiscriminates(t *testing.T) {
	// Flipping a quirk must change the verdict screen — proves the golden
	// actually verifies behavior rather than accepting anything.
	if *dumpDisplays {
		t.Skip("dump mode")
	}
	wrong := DefaultQuirks()
	wrong.ShiftVX = true
	m := runROM(t, "5-quirks.ch8", map[uint16]byte{0x1FF: 1}, wrong, 3600)
	if m.DisplayHash() == goldens["5-quirks"] {
		t.Fatal("quirks test passed with a wrong quirk setting — golden is not discriminating")
	}
}

func TestTimendusBeep(t *testing.T) {
	rom, err := os.ReadFile(filepath.Join("testdata", "timendus", "7-beep.ch8"))
	if err != nil {
		t.Skipf("test ROM unavailable: %v", err)
	}
	m := New(DefaultQuirks())
	if err := m.LoadROM(rom); err != nil {
		t.Fatal(err)
	}
	beeped, silent := false, false
	for range 600 {
		m.TickTimers()
		if _, err := m.RunCycles(11); err != nil {
			t.Fatal(err)
		}
		if m.Beeping() {
			beeped = true
		} else if beeped {
			silent = true
		}
	}
	if !beeped || !silent {
		t.Errorf("beeped=%v wentSilentAfter=%v; want a beeping pattern", beeped, silent)
	}
}
