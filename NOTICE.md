# Third-party notices

## Bundled games

All bundled games are from John Earnest's
[chip8Archive](https://github.com/JohnEarnest/chip8Archive), released under
**CC0** ("No Rights Reserved"). Attribution is a courtesy, given gladly:

- **Cave Explorer** — John Earnest, 2014
- **Outlaw** — John Earnest, 2014
- **Br8kout** — SharpenedSpoon, 2014
- **Flight Runner** — TodPunk, 2014
- **Glitch Ghost** — Jackie Kircher, 2014
- **Danm8ku** — buffi, 2015
- **Snek** — John Earnest, 2021

## Test fixtures (not distributed)

CI runs [Timendus' chip8-test-suite](https://github.com/Timendus/chip8-test-suite)
(GPL-3.0) as golden-image fixtures under `internal/chip8/testdata/` — see the
license copy there. The test ROMs are never embedded into the chipdeck binary
or website.

## Libraries

- **github.com/hajimehoshi/ebiten/v2** (Ebitengine) — Apache License 2.0.
- **github.com/hajimehoshi/bitmapfont/v4** — package Apache License 2.0;
  bundled glyphs derive from M+ FONTS, GNU Unifont (dual OFL/GPLv2+ with
  font-embedding exception), and public-domain sources.
