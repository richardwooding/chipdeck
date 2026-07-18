package chip8

import "testing"

// runFrames drives m like the front end does: TickTimers, then snapshot,
// then the frame's cycles. Snapshotting AFTER the tick matters: SeekCycle
// replays with Step only, so every replayed instruction must observe the
// same timer values live execution saw.
func runFrames(t *testing.T, m *Machine, h *History, frames, cyclesPerFrame int) {
	t.Helper()
	for range frames {
		m.TickTimers()
		if h != nil {
			h.Push(m)
		}
		if _, err := m.RunCycles(cyclesPerFrame); err != nil {
			t.Fatal(err)
		}
	}
}

// rndROM exercises randomness, drawing, AND timer reads — the three things
// replay determinism depends on.
func rndROM() []byte {
	return []byte{
		0xA2, 0x00, // LD I, 200
		0xC0, 0x3F, // RND V0, 3F
		0xC1, 0x1F, // RND V1, 1F
		0xD0, 0x11, // DRW V0, V1, 1
		0xF2, 0x07, // LD V2, DT   (timer read: diverges if replay re-ticks)
		0x32, 0x00, // SE V2, 00
		0xF3, 0x15, // LD DT, V3
		0x73, 0x01, // ADD V3, 01
		0x12, 0x02, // JP 202
	}
}

func TestDeterministicReplay(t *testing.T) {
	m := New(DefaultQuirks())
	if err := m.LoadROM(rndROM()); err != nil {
		t.Fatal(err)
	}
	h := NewHistory(64)
	runFrames(t, m, h, 30, 11)

	// Capture the live end state, rewind to an earlier cycle, replay back
	// up: the states must match exactly.
	endCycles := m.Cycles
	endHash := m.DisplayHash()
	endV := m.V
	endRng := m.rng

	if !h.SeekCycle(m, endCycles/2) {
		t.Fatal("seek to midpoint failed")
	}
	if m.Cycles != endCycles/2 {
		t.Fatalf("Cycles = %d, want %d", m.Cycles, endCycles/2)
	}
	for m.Cycles < endCycles {
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
	}
	if m.DisplayHash() != endHash || m.V != endV || m.rng != endRng {
		t.Fatal("replay from snapshot diverged from live execution")
	}
}

func TestSeekMatchesLiveIntermediateState(t *testing.T) {
	// Record the true state at a specific cycle during a live run, then
	// rewind to it and compare field-for-field.
	m := New(DefaultQuirks())
	if err := m.LoadROM(rndROM()); err != nil {
		t.Fatal(err)
	}
	h := NewHistory(64)

	var want Machine
	const targetFrame = 17
	for f := range 30 {
		h.Push(m)
		m.TickTimers()
		if _, err := m.RunCycles(11); err != nil {
			t.Fatal(err)
		}
		if f == targetFrame {
			want = *m // live truth at this cycle
		}
	}

	if !h.SeekCycle(m, want.Cycles) {
		t.Fatal("seek failed")
	}
	if m.V != want.V || m.I != want.I || m.PC != want.PC || m.rng != want.rng ||
		m.Display != want.Display || m.Cycles != want.Cycles {
		t.Fatal("rewound state differs from live state at the same cycle")
	}
}

func TestHistoryHorizonAndWrap(t *testing.T) {
	m := New(DefaultQuirks())
	if err := m.LoadROM(rndROM()); err != nil {
		t.Fatal(err)
	}
	h := NewHistory(8) // tiny ring: old frames must fall off
	runFrames(t, m, h, 30, 11)

	if h.Len() != 8 {
		t.Fatalf("Len = %d, want 8", h.Len())
	}
	horizon, ok := h.Oldest()
	if !ok || horizon == 0 {
		t.Fatalf("horizon = %d ok=%v; want a wrapped, nonzero horizon", horizon, ok)
	}
	if h.SeekCycle(m, 0) {
		t.Fatal("seek before the horizon should fail")
	}
	if !h.SeekCycle(m, horizon) {
		t.Fatal("seek to the horizon itself should succeed")
	}
}

func TestSeekRewritesHistory(t *testing.T) {
	m := New(DefaultQuirks())
	if err := m.LoadROM(rndROM()); err != nil {
		t.Fatal(err)
	}
	h := NewHistory(64)
	runFrames(t, m, h, 20, 11)

	mid := m.Cycles / 2
	if !h.SeekCycle(m, mid) {
		t.Fatal("seek failed")
	}
	// All remaining snapshots must be <= mid: the future was discarded.
	for i := 1; i <= h.Len(); i++ {
		idx := (h.head - i + len(h.ring)) % len(h.ring)
		if h.ring[idx].Cycles > mid {
			t.Fatalf("snapshot with Cycles=%d survived a seek to %d", h.ring[idx].Cycles, mid)
		}
	}
}
