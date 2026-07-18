# chipdeck

A CHIP-8 emulator with a **live debugger**, written entirely in Go — a
pure interpreter core under an Ebitengine front end, running native and in
the browser via WebAssembly.

**▶ Run it: https://richardwooding.github.io/chipdeck/play/**

Seven CC0 games are bundled (Cave Explorer, Outlaw, Br8kout, Flight Runner,
Glitch Ghost, Danm8ku, Snek — all from
[chip8Archive](https://github.com/JohnEarnest/chip8Archive)), or drop any
`.ch8` file onto the page. The debugger panel shows registers, the stack,
a disassembly window around PC, and an instruction trace — pause with
`space`, single-step with `n`, and watch the machine think.

## Accuracy

The core implements original **COSMAC VIP** behavior (toggleable quirks):
vF reset on logic ops, `I` increment on register load/store, display-wait
vblank gating, sprite clipping, vY-sourced shifts, v0 jumps, and
key-release semantics for `Fx0A`. It passes the
[Timendus test suite](https://github.com/Timendus/chip8-test-suite) —
corax+ opcodes, flags, and all six quirks in CHIP-8 mode — verified in CI
as golden framebuffer hashes, with a negative test proving the goldens
discriminate. Bundled archive games run Octo-style (manifest tickrate, no
display wait); dropped ROMs default to authentic VIP pacing.

## Controls

| Key | Action |
| --- | --- |
| `1234` / `QWER` / `ASDF` / `ZXCV` | the COSMAC hex keypad (`123C/456D/789E/A0BF`) |
| `space` | pause / resume |
| `n` | step one instruction (while paused) |
| `b` | reset the ROM |
| `g` | toggle the debugger panel |
| `p` | toggle phosphor persistence |
| `+` / `-` | double / halve the emulation speed |
| `esc` | back to the game picker |

Deep-link a bundled game with `?rom=0` … `?rom=6`.

## Development

```sh
go test ./internal/chip8/    # opcode tables + Timendus goldens, fully headless
go run .                     # native window
GOOS=js GOARCH=wasm go build -o docs/play/chipdeck.wasm .
```

Regenerate golden hashes after intentional core changes with
`go test ./internal/chip8/ -run TestTimendus -v -chip8.dump` and eyeball
the ASCII framebuffers.

Game credits in [NOTICE.md](NOTICE.md). MIT.
