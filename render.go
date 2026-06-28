package main

import (
	"fmt"
	"time"

	"github.com/nsf/termbox-go"
)

func drawCell(x, y int, ch rune, fg, bg termbox.Attribute) {
	termbox.SetCell(x*2, y, ch, fg, bg)
	termbox.SetCell(x*2+1, y, ' ', fg, bg)
}

func (g *Game) Render() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	scoreFg := termbox.ColorWhite | termbox.AttrBold
	scoreBg := termbox.ColorBlack
	flashActive := g.Flash != nil && time.Now().Before(g.Flash.ExpiresAt)
	if flashActive {
		elapsed := time.Since(g.Flash.ExpiresAt.Add(-FlashDuration))
		if elapsed.Milliseconds()/120%2 == 0 {
			scoreBg = g.Flash.Color
			scoreFg = termbox.ColorWhite | termbox.AttrBold
		}
	}
	scoreLine := fmt.Sprintf(" P1(Red): %d  |  P2(Blue): %d ", g.Player1.Score, g.Player2.Score)
	ctrlLine := " P1: WASD+Q Boost  |  P2: Arrows+Space Boost  |  P: Pause  R: Restart "
	for i, c := range scoreLine {
		termbox.SetCell(i, 0, c, scoreFg, scoreBg)
	}
	for i, c := range ctrlLine {
		termbox.SetCell(i, 1, c, termbox.ColorDarkGray, termbox.ColorBlack)
	}
	if flashActive {
		flashLine := g.Flash.Text
		startX := len(scoreLine) + 2
		for i, c := range flashLine {
			termbox.SetCell(startX+i, 0, c, g.Flash.Color|termbox.AttrBold, termbox.ColorBlack)
		}
	}
	offsetY := 3
	for x := 0; x < MapWidth; x++ {
		drawCell(x, offsetY-1, '─', termbox.ColorDarkGray, termbox.ColorBlack)
		drawCell(x, offsetY+MapHeight, '─', termbox.ColorDarkGray, termbox.ColorBlack)
	}
	for y := 0; y < MapHeight; y++ {
		termbox.SetCell(0, offsetY+y, '│', termbox.ColorDarkGray, termbox.ColorBlack)
		termbox.SetCell(MapWidth*2+1, offsetY+y, '│', termbox.ColorDarkGray, termbox.ColorBlack)
	}
	termbox.SetCell(0, offsetY-1, '┌', termbox.ColorDarkGray, termbox.ColorBlack)
	termbox.SetCell(MapWidth*2+1, offsetY-1, '┐', termbox.ColorDarkGray, termbox.ColorBlack)
	termbox.SetCell(0, offsetY+MapHeight, '└', termbox.ColorDarkGray, termbox.ColorBlack)
	termbox.SetCell(MapWidth*2+1, offsetY+MapHeight, '┘', termbox.ColorDarkGray, termbox.ColorBlack)
	for _, w := range g.Walls {
		drawCell(w.X, w.Y+offsetY, '█', termbox.ColorDarkGray, termbox.ColorBlack)
	}
	for _, c := range g.Corpses {
		drawCell(c.Pos.X, c.Pos.Y+offsetY, 'x', c.Color, termbox.ColorBlack)
	}
	for _, f := range g.Foods {
		ch := '●'
		if f.Points >= 3 {
			ch = '◆'
		}
		drawCell(f.Pos.X, f.Pos.Y+offsetY, ch, termbox.ColorGreen|termbox.AttrBold, termbox.ColorBlack)
	}
	for _, s := range g.Stars {
		drawCell(s.Pos.X, s.Pos.Y+offsetY, '★', termbox.ColorYellow|termbox.AttrBold, termbox.ColorBlack)
	}
	if g.Diamond != nil {
		drawCell(g.Diamond.Pos.X, g.Diamond.Pos.Y+offsetY, '♦', termbox.ColorYellow|termbox.AttrBold, termbox.ColorBlack)
	}
	g.renderSnake(g.Player1, offsetY)
	g.renderSnake(g.Player2, offsetY)
	if g.State == StatePaused {
		g.renderOverlay("PAUSED - Press P to resume")
	}
	if g.State == StateGameOver {
		g.renderGameOver()
	}
	termbox.Flush()
}

func (g *Game) renderSnake(s *Snake, offsetY int) {
	if !s.Alive {
		return
	}
	starActive := s.HasStar()
	for i, b := range s.Body {
		ch := '■'
		fg := s.Color
		if i == 0 {
			ch = '●'
			fg = s.HeadColor
			if starActive {
				fg = fg | termbox.AttrBold
				ch = '◎'
			}
		}
		bg := termbox.ColorBlack
		if starActive && i%2 == 0 {
			bg = termbox.ColorMagenta
		}
		drawCell(b.X, b.Y+offsetY, ch, fg, bg)
	}
}

func (g *Game) renderOverlay(text string) {
	w, h := termbox.Size()
	startX := (w - len(text)) / 2
	startY := h / 2
	bg := termbox.ColorMagenta
	for i := 0; i < len(text)+4; i++ {
		termbox.SetCell(startX-2+i, startY-1, ' ', termbox.ColorWhite, bg)
		termbox.SetCell(startX-2+i, startY, ' ', termbox.ColorWhite, bg)
		termbox.SetCell(startX-2+i, startY+1, ' ', termbox.ColorWhite, bg)
	}
	for i, c := range text {
		termbox.SetCell(startX+i, startY, c, termbox.ColorWhite|termbox.AttrBold, bg)
	}
	_ = h
}

func (g *Game) renderGameOver() {
	w, h := termbox.Size()
	winner := "DRAW"
	if g.Player1.Score > g.Player2.Score {
		winner = "PLAYER 1 (RED) WINS!"
	} else if g.Player2.Score > g.Player1.Score {
		winner = "PLAYER 2 (BLUE) WINS!"
	}
	lines := []string{
		"GAME OVER",
		winner,
		fmt.Sprintf("P1: %d  |  P2: %d", g.Player1.Score, g.Player2.Score),
		"Press R to restart, M for menu",
	}
	maxLen := 0
	for _, l := range lines {
		if len(l) > maxLen {
			maxLen = len(l)
		}
	}
	startX := (w - maxLen) / 2
	startY := h/2 - len(lines)/2
	bg := termbox.ColorRed
	for li, line := range lines {
		for i := 0; i < maxLen+4; i++ {
			termbox.SetCell(startX-2+i, startY+li, ' ', termbox.ColorWhite, bg)
		}
		pad := (maxLen - len(line)) / 2
		for i, c := range line {
			termbox.SetCell(startX+pad+i, startY+li, c, termbox.ColorWhite|termbox.AttrBold, bg)
		}
	}
}

func (g *Game) RenderMenu() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	w, h := termbox.Size()
	title := "  === TWO PLAYER SNAKE BATTLE ===  "
	sub := " Select a map with Arrow Keys, Enter to start "
	instructions := []string{
		"",
		"PLAYER 1 (RED)    PLAYER 2 (BLUE)",
		"  W A S D          Arrow Keys     = Move",
		"  Q (tap)          Space (tap)    = Boost",
		"",
		"Rules:",
		"  ● = +1 pt    ◆ = +3 pts",
		"  ♦ = +5 pts & grow 3! (respawns instantly)",
		"  ★ = Wall-pass for 2 seconds",
		"  Die: wall, self, opponent, or 'x' corpse (3s)",
		"  P = Pause    R = Restart    M = Menu    Esc = Quit",
		"",
	}
	startX := (w - len(title)) / 2
	startY := 2
	for i, c := range title {
		termbox.SetCell(startX+i, startY, c, termbox.ColorCyan|termbox.AttrBold, termbox.ColorBlack)
	}
	startX = (w - len(sub)) / 2
	for i, c := range sub {
		termbox.SetCell(startX+i, startY+1, c, termbox.ColorWhite, termbox.ColorBlack)
	}
	menuStartY := startY + 4
	for i, name := range g.MenuOptions {
		line := fmt.Sprintf("  %d. %s  ", i+1, name)
		selected := i == g.SelectedMap
		bg := termbox.ColorBlack
		fg := termbox.ColorWhite
		prefix := "   "
		if selected {
			bg = termbox.ColorCyan
			fg = termbox.ColorBlack | termbox.AttrBold
			prefix = " > "
		}
		fullLine := prefix + line
		lx := (w - len(fullLine)) / 2
		for j := 0; j < len(fullLine)+4; j++ {
			termbox.SetCell(lx-2+j, menuStartY+i, ' ', fg, bg)
		}
		for j, c := range fullLine {
			termbox.SetCell(lx+j, menuStartY+i, c, fg, bg)
		}
	}
	insY := menuStartY + len(g.MenuOptions) + 2
	for li, line := range instructions {
		lx := (w - len(line)) / 2
		for i, c := range line {
			fg := termbox.ColorDarkGray
			if li == 1 || li == 2 || li == 3 {
				fg = termbox.ColorLightGray
			}
			if li >= 5 && li <= 10 {
				fg = termbox.ColorLightYellow
			}
			termbox.SetCell(lx+i, insY+li, c, fg, termbox.ColorBlack)
		}
	}
	_ = h
	termbox.Flush()
}
