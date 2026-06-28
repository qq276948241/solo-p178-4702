package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/nsf/termbox-go"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	err := termbox.Init()
	if err != nil {
		fmt.Fprintln(os.Stderr, "termbox init error:", err)
		os.Exit(1)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc | termbox.InputAlt)
	game := NewMenu()
	eventQueue := make(chan termbox.Event, 64)
	go func() {
		for {
			ev := termbox.PollEvent()
			eventQueue <- ev
		}
	}()
	ticker := time.NewTicker(30 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case ev := <-eventQueue:
			if ev.Type == termbox.EventError {
				return
			}
			if ev.Type != termbox.EventKey {
				continue
			}
			var cont bool
			if game.State == StateMenu {
				cont = game.HandleKeyMenu(ev)
			} else {
				cont = game.HandleKeyPlaying(ev)
			}
			if !cont {
				return
			}
		case <-ticker.C:
			if game.State == StateMenu {
				game.RenderMenu()
				continue
			}
			now := time.Now()
			tr := game.GetTickRate()
			if now.Sub(game.LastTick) >= tr {
				if game.State == StatePlaying {
					game.Tick()
				}
				game.LastTick = now
			}
			game.Render()
		}
	}
}
