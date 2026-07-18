package ui

import (
	"fmt"
	"image"
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/richardwooding/chipdeck/internal/chip8"
	"github.com/richardwooding/chipdeck/internal/roms"
)

const (
	pickListY = 200
	pickRowH  = 44
)

// pickerScene is the landing screen: bundled CC0 games plus drag-drop and a
// browse button for your own .ch8 files.
type pickerScene struct {
	selected  int
	errMsg    string
	autoTried bool
}

func newPickerScene() *pickerScene { return &pickerScene{autoTried: true} }

// newBootScene is the picker used at startup; it honors ?rom=N deep links.
func newBootScene() *pickerScene { return &pickerScene{} }

func justTaps() []image.Point {
	var pts []image.Point
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		pts = append(pts, image.Pt(x, y))
	}
	for _, id := range inpututil.AppendJustPressedTouchIDs(nil) {
		x, y := ebiten.TouchPosition(id)
		pts = append(pts, image.Pt(x, y))
	}
	return pts
}

func (s *pickerScene) browseBtn() (x0, y0, x1, y1 float64) {
	games := len(roms.Games())
	y := float64(pickListY + games*pickRowH + 16)
	return W/2 - 90, y, W/2 + 90, y + 40
}

func (s *pickerScene) Update(g *Game) error {
	games := roms.Games()

	if !s.autoTried {
		s.autoTried = true
		if n := autostartROM(); n >= 0 && n < len(games) {
			s.selected = n
			s.startBundled(g, games[n])
			return nil
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) && s.selected < len(games)-1 {
		s.selected++
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) && s.selected > 0 {
		s.selected--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		s.startBundled(g, games[s.selected])
		return nil
	}

	if taps := justTaps(); len(taps) > 0 {
		bx0, by0, bx1, by1 := s.browseBtn()
		for _, pt := range taps {
			if canPickFiles() && float64(pt.X) >= bx0 && float64(pt.X) < bx1 && float64(pt.Y) >= by0 && float64(pt.Y) < by1 {
				openFilePicker()
				continue
			}
			for i := range games {
				y := pickListY + i*pickRowH
				if pt.Y >= y && pt.Y < y+pickRowH && pt.X >= 120 && pt.X < W-120 {
					if s.selected == i {
						s.startBundled(g, games[i])
						return nil
					}
					s.selected = i
				}
			}
		}
	}

	if data, name, ok := takePickedFile(); ok {
		s.startCustom(g, data, name)
		return nil
	}
	if files := ebiten.DroppedFiles(); files != nil {
		if data, name, ok := firstFile(files); ok {
			s.startCustom(g, data, name)
			return nil
		}
	}
	return nil
}

// startBundled runs an archive game: Octo pacing (manifest tickrate, no
// display wait).
func (s *pickerScene) startBundled(g *Game, e roms.Entry) {
	q := chip8.DefaultQuirks()
	q.DisplayWait = false
	s.start(g, e.Data, e.Title, e.Controls, q, e.Tickrate)
}

// startCustom runs a dropped/browsed ROM with authentic VIP behavior.
func (s *pickerScene) startCustom(g *Game, data []byte, name string) {
	s.start(g, data, name, "keys: 1234/QWER/ASDF/ZXCV", chip8.DefaultQuirks(), 11)
}

func (s *pickerScene) start(g *Game, data []byte, title, controls string, q chip8.Quirks, tickrate int) {
	m := chip8.New(q)
	if err := m.LoadROM(data); err != nil {
		s.errMsg = fmt.Sprintf("%s: %v", title, err)
		return
	}
	g.scene = newPlayScene(m, title, controls, tickrate)
}

func firstFile(files fs.FS) (data []byte, name string, ok bool) {
	_ = fs.WalkDir(files, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || ok {
			return nil
		}
		if b, rerr := fs.ReadFile(files, path); rerr == nil {
			data, name, ok = b, d.Name(), true
		}
		return nil
	})
	return data, name, ok
}

func (s *pickerScene) Draw(dst *ebiten.Image) {
	title := "CHIPDECK"
	drawText(dst, title, (W-textWidth(title, 6))/2, 50, colAccent, 6)
	sub := "a CHIP-8 emulator with a live debugger — 100% Go"
	drawText(dst, sub, (W-textWidth(sub, 2))/2, 140, colDim, 2)

	games := roms.Games()
	for i, e := range games {
		y := float64(pickListY + i*pickRowH)
		if i == s.selected {
			vector.FillRect(dst, 120, float32(y)-4, W-240, pickRowH-6, colPanel, false)
			vector.StrokeRect(dst, 120, float32(y)-4, W-240, pickRowH-6, 1, colAccent, false)
			drawText(dst, "▶", 136, y+4, colAccent, 2)
		}
		drawText(dst, e.Title, 170, y+4, colText, 2)
		meta := fmt.Sprintf("%s · %s", e.Author, e.Year)
		drawText(dst, meta, W-140-textWidth(meta, 1), y+2, colDim, 1)
		drawText(dst, e.Desc, W-140-textWidth(e.Desc, 1), y+16, colDimmer, 1)
	}

	if canPickFiles() {
		x0, y0, x1, y1 := s.browseBtn()
		vector.FillRect(dst, float32(x0), float32(y0), float32(x1-x0), float32(y1-y0), colPanel, false)
		vector.StrokeRect(dst, float32(x0), float32(y0), float32(x1-x0), float32(y1-y0), 1, colAccent, false)
		lbl := "browse .ch8 files…"
		drawText(dst, lbl, x0+(x1-x0-textWidth(lbl, 1))/2, y0+(y1-y0-glyphH)/2, colAccent, 1)
	}

	if s.errMsg != "" {
		drawText(dst, s.errMsg, (W-textWidth(s.errMsg, 1))/2, H-64, colAmber, 1)
	}
	foot := "drop a .ch8 anywhere · games are CC0 from chip8Archive · tap/enter to play"
	drawText(dst, foot, (W-textWidth(foot, 1))/2, H-32, colDimmer, 1)
}
