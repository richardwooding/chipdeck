# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

chipdeck is a CHIP-8 emulator: a pure-Go interpreter core under an Ebitengine
front end with a live debugger panel. One codebase runs native (`go run .`,
the dev loop) and as WebAssembly on GitHub Pages
(https://richardwooding.github.io/chipdeck/play/).

## Commands

```sh
go test ./internal/chip8/                    # core: opcode tables + Timendus goldens, fully headless
go test ./internal/chip8/ -run TestTimendusQuirks -v   # one test
go test ./internal/chip8/ -run TestTimendus -v -chip8.dump   # print framebuffers instead of asserting
go run .                                     # native window (macOS)
GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o docs/play/chipdeck.wasm .
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" docs/play/
(cd docs/play && python3 -m http.server 8080)   # then open ?rom=N to deep-link a bundled game
go vet ./... && gofmt -l .                   # CI also runs golangci-lint (latest) — errcheck and unused are strict
```

CI runs `go test` under `xvfb-run` because importing Ebitengine panics at init
on a display-less Linux runner. This matters the moment `internal/ui` gains a
test file; `internal/chip8` must never need it (see below).

## Architecture and invariants

**`internal/chip8` is a pure interpreter with zero Ebitengine imports.** Keep
it that way — every test in it runs headlessly, and the golden tests depend on
that. The front end drives it with the two-call contract: `TickTimers()` at
60Hz (decrements timers, opens the display-wait gate), then
`RunCycles(tickrate)` for that frame's CPU budget. `RunCycles` parks early
when `Fx0A` blocks or when the display-wait quirk gates after a `DRW`.

Core behavioral invariants (all covered by tests; breaking them fails the
Timendus goldens):

- **vF is written last** in the 8xxx ALU group, so `VF` as an operand keeps
  original semantics (the flags test checks `8FF4` style cases).
- **`Fx0A` waits for key RELEASE**, not press (COSMAC VIP semantics), tracked
  by `waitReg`/`waitPressed` with PC rewind while blocked.
- **Rewind determinism**: `History` snapshots are taken AFTER `TickTimers`
  and before the frame's `RunCycles`; `SeekCycle` replays with `Step` only
  and never re-ticks, so timers must be frame-stable during replay. The
  RNG is xorshift state inside `Machine` (snapshot-carried); the `Rand`
  func field is a test-only override that breaks replay determinism.
- **Quirks default to original COSMAC VIP** (`DefaultQuirks()`): vF reset ON,
  memory-increment ON, display-wait ON, clipping ON, shift-vX OFF, jump-vX
  OFF. Sprite *origins* wrap even when the *body* clips.

**Golden tests** (`integration_test.go`): Timendus test ROMs run until the
framebuffer is stable for 60 frames, then `DisplayHash()` is compared to a
committed constant. To regenerate after an intentional core change: run with
`-chip8.dump`, read the ASCII framebuffers (every opcode group must show the
same pass glyph; the quirks screen must show all six correct), then update the
`goldens` map. `TestTimendusQuirksDiscriminates` flips a quirk and asserts the
hash *changes* — never delete it; it's what makes the goldens meaningful. The
quirks ROM's menu is skipped by poking `0x1FF = 1` (CHIP-8 mode).

**ROM licensing is a hard boundary.** `internal/roms` embeds only CC0 games
from JohnEarnest/chip8Archive (manifest in `roms.go`, attribution in
NOTICE.md). The Timendus suite is **GPL-3.0** and lives exclusively in
`internal/chip8/testdata/timendus/` as CI fixtures with its LICENSE alongside
— it must never be `go:embed`ed or otherwise shipped in the binary/site.

**Two pacing modes.** Bundled archive games are Octo-era: they run with
`DisplayWait=false` and the per-game `Tickrate` from the manifest (7 to 1000
cycles/frame — Danm8ku needs the high end). Dropped/browsed ROMs get authentic
VIP behavior: `DefaultQuirks()` at 11 cycles/frame (~700 ips).

**`internal/ui` details worth knowing:**

- `screen.go`'s phosphor persistence (decay to 55% per frame, toggled by `P`)
  is presentation only — it must never leak into the core or affect goldens.
- The beeper is an endless square-wave `io.Reader` gated by an atomic bool;
  the audio player is created lazily on first scene use (browser autoplay
  needs a prior user gesture) and is never stopped — silence is written when
  gated off.
- `filepicker_js.go` / `filepicker_native.go` are build-tagged; the JS picker
  must be invoked soon after a user tap (transient activation).
- `?rom=N` deep-links a bundled game (also how headless screenshots reach the
  play scene: headless Chrome needs `--use-angle=swiftshader
  --enable-unsafe-swiftshader` for WebGL2).

## Deploy

Push to `main` touching `docs/**`, `internal/**`, or `main.go` triggers
`pages.yml`: it builds the wasm into `docs/play/` (the wasm and `wasm_exec.js`
are gitignored, never committed) and deploys `docs/` to GitHub Pages. The
landing page is gloam-styled with vendored `gloam.css`/`gloam.js`, kept in
sync by the weekly `gloam-sync.yml` PR workflow.
