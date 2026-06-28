package main

import (
	"time"

	"github.com/nsf/termbox-go"
)

const (
	MapWidth  = 24
	MapHeight = 16

	TickNormal = 120 * time.Millisecond
	TickBoost  = 55 * time.Millisecond

	StarDuration       = 2000 * time.Millisecond
	CorpseDuration     = 3000 * time.Millisecond
	DiamondSpawnChance = 8
	DiamondPoints      = 5
	DiamondGrow        = 3
	FlashDuration      = 1200 * time.Millisecond
	FoodSpawnChance    = 7
	StarSpawnChance    = 15
	BoostDuration      = 250 * time.Millisecond
)

type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

type Point struct {
	X, Y int
}

type Snake struct {
	Body        []Point
	Dir         Direction
	NextDir     Direction
	Color       termbox.Attribute
	HeadColor   termbox.Attribute
	Score       int
	Alive       bool
	BoostEnd    time.Time
	StarEnd     time.Time
	CorpseEnd   time.Time
	GrowPending int
}

type Food struct {
	Pos    Point
	Points int
}

type Star struct {
	Pos Point
}

type Diamond struct {
	Pos Point
}

type FlashMsg struct {
	Text      string
	Color     termbox.Attribute
	ExpiresAt time.Time
}

type Corpse struct {
	Pos       Point
	Color     termbox.Attribute
	ExpiresAt time.Time
}

type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StatePaused
	StateGameOver
)

type Game struct {
	State       GameState
	MapID       int
	Walls       []Point
	Player1     *Snake
	Player2     *Snake
	Foods       []Food
	Stars       []Star
	Diamond     *Diamond
	Corpses     []Corpse
	Flash       *FlashMsg
	LastTick    time.Time
	SelectedMap int
	MenuOptions []string
}

func (s *Snake) Head() Point {
	if len(s.Body) == 0 {
		return Point{-1, -1}
	}
	return s.Body[0]
}

func (s *Snake) HasStar() bool {
	return time.Now().Before(s.StarEnd)
}

func (s *Snake) HasBoost() bool {
	return s.Alive && time.Now().Before(s.BoostEnd)
}

func (g *Game) IsWall(p Point) bool {
	if p.X < 0 || p.X >= MapWidth || p.Y < 0 || p.Y >= MapHeight {
		return true
	}
	for _, w := range g.Walls {
		if w == p {
			return true
		}
	}
	return false
}
