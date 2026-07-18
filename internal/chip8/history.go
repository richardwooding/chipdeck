package chip8

// History is a ring of whole-machine snapshots enabling time travel. The
// front end pushes one snapshot per 60Hz frame (before running that frame's
// cycles); stepping backwards restores the nearest earlier snapshot and
// replays forward. Replay is exact because a Machine snapshot carries
// everything execution depends on — including the rng state — and inputs
// only change at frame boundaries.
type History struct {
	ring []Machine
	head int // next write position
	size int
}

// NewHistory returns a ring holding up to capacity snapshots. At one
// snapshot per frame, 600 ≈ ten seconds of rewind for ~4MB.
func NewHistory(capacity int) *History {
	if capacity < 1 {
		capacity = 1
	}
	return &History{ring: make([]Machine, capacity)}
}

// Push records a snapshot of m.
func (h *History) Push(m *Machine) {
	h.ring[h.head] = *m
	h.head = (h.head + 1) % len(h.ring)
	if h.size < len(h.ring) {
		h.size++
	}
}

// Len reports how many snapshots are held.
func (h *History) Len() int { return h.size }

// Oldest returns the cycle count of the oldest snapshot — the rewind
// horizon — and false when the ring is empty.
func (h *History) Oldest() (uint64, bool) {
	if h.size == 0 {
		return 0, false
	}
	idx := (h.head - h.size + len(h.ring)) % len(h.ring)
	return h.ring[idx].Cycles, true
}

// before returns the newest snapshot with Cycles <= target, or nil.
func (h *History) before(target uint64) *Machine {
	for i := 1; i <= h.size; i++ {
		idx := (h.head - i + len(h.ring)) % len(h.ring)
		if h.ring[idx].Cycles <= target {
			return &h.ring[idx]
		}
	}
	return nil
}

// SeekCycle rewinds m to exactly the state it had when Cycles == target, by
// restoring the nearest earlier snapshot and replaying forward. Snapshots
// newer than the target are discarded (history is rewritten from here).
// Reports false when target is behind the rewind horizon.
func (h *History) SeekCycle(m *Machine, target uint64) bool {
	snap := h.before(target)
	if snap == nil {
		return false
	}
	*m = *snap
	for m.Cycles < target {
		before := m.Cycles
		if err := m.Step(); err != nil {
			return false
		}
		if m.Cycles == before {
			break // parked on Fx0A with no key events to replay
		}
	}
	h.dropAfter(target)
	return true
}

// dropAfter forgets snapshots newer than target so a rewind followed by new
// execution doesn't leave future states in the ring.
func (h *History) dropAfter(target uint64) {
	for h.size > 0 {
		idx := (h.head - 1 + len(h.ring)) % len(h.ring)
		if h.ring[idx].Cycles <= target {
			return
		}
		h.head = idx
		h.size--
	}
}
