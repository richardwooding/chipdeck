// Package roms embeds the bundled CHIP-8 games. Everything here is CC0 from
// John Earnest's chip8Archive (github.com/JohnEarnest/chip8Archive);
// attribution is a courtesy, given gladly here and in NOTICE.md. Tickrates
// come from the archive's programs.json (Octo semantics: cycles per 60Hz
// frame). Archive games are Octo-era, so they run without the display-wait
// quirk; the tickrate paces them instead.
package roms

import _ "embed"

// Entry is one bundled game.
type Entry struct {
	Title    string
	Author   string
	Year     string
	Desc     string
	Controls string
	Tickrate int // cycles per 60Hz frame
	Data     []byte
}

var (
	//go:embed caveexplorer.ch8
	caveExplorer []byte
	//go:embed outlaw.ch8
	outlaw []byte
	//go:embed br8kout.ch8
	br8kout []byte
	//go:embed flightrunner.ch8
	flightRunner []byte
	//go:embed glitchGhost.ch8
	glitchGhost []byte
	//go:embed danm8ku.ch8
	danm8ku []byte
	//go:embed snek.ch8
	snek []byte
)

// Games lists the bundled titles, shown in picker order.
func Games() []Entry {
	return []Entry{
		{"Cave Explorer", "John Earnest", "2014", "Explore a sprawling cave", "ASWD move", 20, caveExplorer},
		{"Outlaw", "John Earnest", "2014", "Wild-west shootout", "ASWD move · E shoot", 15, outlaw},
		{"Br8kout", "SharpenedSpoon", "2014", "Breakout in 199 bytes", "A/D paddle", 7, br8kout},
		{"Flight Runner", "TodPunk", "2014", "Dodge the walls", "W/S climb and dive", 15, flightRunner},
		{"Glitch Ghost", "Jackie Kircher", "2014", "Haunt the glitches", "ASWD move · E flip", 200, glitchGhost},
		{"Danm8ku", "buffi", "2015", "Bullet-hell chaos", "ASWD dodge", 1000, danm8ku},
		{"Snek", "John Earnest", "2021", "Snake in 65 bytes", "ASWD steer", 1000, snek},
	}
}
