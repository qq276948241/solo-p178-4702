package main

import (
	"time"

	"github.com/nsf/termbox-go"
)

func (g *Game) HandleKeyMenu(ev termbox.Event) bool {
	switch ev.Key {
	case termbox.KeyArrowUp:
		g.SelectedMap--
		if g.SelectedMap < 0 {
			g.SelectedMap = len(g.MenuOptions) - 1
		}
	case termbox.KeyArrowDown:
		g.SelectedMap++
		if g.SelectedMap >= len(g.MenuOptions) {
			g.SelectedMap = 0
		}
	case termbox.KeyEnter:
		*g = *NewGame(g.SelectedMap)
	case termbox.KeyEsc:
		return false
	case termbox.KeySpace:
		*g = *NewGame(g.SelectedMap)
	}
	return true
}

func (g *Game) HandleKeyPlaying(ev termbox.Event) bool {
	switch ev.Key {
	case termbox.KeyEsc:
		return false
	case termbox.KeyArrowUp:
		if g.Player2.Dir != Down {
			g.Player2.NextDir = Up
		}
	case termbox.KeyArrowDown:
		if g.Player2.Dir != Up {
			g.Player2.NextDir = Down
		}
	case termbox.KeyArrowLeft:
		if g.Player2.Dir != Right {
			g.Player2.NextDir = Left
		}
	case termbox.KeyArrowRight:
		if g.Player2.Dir != Left {
			g.Player2.NextDir = Right
		}
	case termbox.KeySpace:
		if g.Player2.Alive {
			g.Player2.BoostEnd = time.Now().Add(BoostDuration)
		}
	}
	switch ev.Ch {
	case 'w', 'W':
		if g.Player1.Dir != Down {
			g.Player1.NextDir = Up
		}
	case 's', 'S':
		if g.Player1.Dir != Up {
			g.Player1.NextDir = Down
		}
	case 'a', 'A':
		if g.Player1.Dir != Right {
			g.Player1.NextDir = Left
		}
	case 'd', 'D':
		if g.Player1.Dir != Left {
			g.Player1.NextDir = Right
		}
	case 'q', 'Q':
		if g.Player1.Alive {
			g.Player1.BoostEnd = time.Now().Add(BoostDuration)
		}
	case 'p', 'P':
		if g.State == StatePlaying {
			g.State = StatePaused
			g.PausedAt = time.Now()
		} else if g.State == StatePaused {
			pauseDur := time.Since(g.PausedAt)
			g.shiftTimers(pauseDur)
			g.State = StatePlaying
		}
	case 'r', 'R':
		*g = *NewGame(g.MapID)
	case 'm', 'M':
		*g = *NewMenu()
	}
	return true
}
