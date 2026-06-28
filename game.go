package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/nsf/termbox-go"
)

func NewGame(mapID int) *Game {
	g := &Game{
		State:       StatePlaying,
		MapID:       mapID,
		Walls:       make([]Point, len(mapConfigs[mapID])),
		Player1:     NewSnake(5, 8, Right, termbox.ColorRed, termbox.ColorLightRed),
		Player2:     NewSnake(18, 8, Left, termbox.ColorBlue, termbox.ColorLightBlue),
		Foods:       []Food{},
		Stars:       []Star{},
		Corpses:     []Corpse{},
		LastTick:    time.Now(),
		MenuOptions: mapNames,
	}
	copy(g.Walls, mapConfigs[mapID])
	g.SpawnFood()
	g.SpawnFood()
	if rand.Intn(DiamondSpawnChance) == 0 {
		g.SpawnDiamond()
	}
	return g
}

func NewMenu() *Game {
	return &Game{
		State:       StateMenu,
		SelectedMap: 0,
		MenuOptions: mapNames,
	}
}

func (g *Game) Occupied(p Point) bool {
	if p.X < 0 || p.X >= MapWidth || p.Y < 0 || p.Y >= MapHeight {
		return true
	}
	for _, w := range g.Walls {
		if w == p {
			return true
		}
	}
	if g.Player1.Alive {
		for _, b := range g.Player1.Body {
			if b == p {
				return true
			}
		}
	}
	if g.Player2.Alive {
		for _, b := range g.Player2.Body {
			if b == p {
				return true
			}
		}
	}
	for _, f := range g.Foods {
		if f.Pos == p {
			return true
		}
	}
	for _, s := range g.Stars {
		if s.Pos == p {
			return true
		}
	}
	if g.Diamond != nil && g.Diamond.Pos == p {
		return true
	}
	for _, c := range g.Corpses {
		if c.Pos == p {
			return true
		}
	}
	return false
}

func (g *Game) RandomEmpty() Point {
	for i := 0; i < 500; i++ {
		p := Point{rand.Intn(MapWidth), rand.Intn(MapHeight)}
		if !g.Occupied(p) {
			return p
		}
	}
	return Point{-1, -1}
}

func (g *Game) SpawnFood() {
	p := g.RandomEmpty()
	if p.X >= 0 {
		pts := 1
		if rand.Intn(5) == 0 {
			pts = 3
		}
		g.Foods = append(g.Foods, Food{Pos: p, Points: pts})
	}
}

func (g *Game) SpawnStar() {
	p := g.RandomEmpty()
	if p.X >= 0 {
		g.Stars = append(g.Stars, Star{Pos: p})
	}
}

func (g *Game) SpawnDiamond() {
	p := g.RandomEmpty()
	if p.X >= 0 {
		g.Diamond = &Diamond{Pos: p}
	}
}

func (g *Game) NextPos(s *Snake) Point {
	head := s.Head()
	switch s.NextDir {
	case Up:
		return Point{head.X, head.Y - 1}
	case Down:
		return Point{head.X, head.Y + 1}
	case Left:
		return Point{head.X - 1, head.Y}
	case Right:
		return Point{head.X + 1, head.Y}
	}
	return head
}

func (g *Game) CheckCollision(s *Snake, other *Snake, nextPos Point) bool {
	if !s.HasStar() {
		if g.IsWall(nextPos) {
			return true
		}
	}
	for i, b := range s.Body {
		if i == len(s.Body)-1 && s.GrowPending == 0 {
			continue
		}
		if b == nextPos {
			return true
		}
	}
	if other.Alive {
		for _, b := range other.Body {
			if b == nextPos {
				return true
			}
		}
	}
	for _, c := range g.Corpses {
		if c.Pos == nextPos {
			return true
		}
	}
	return false
}

func (g *Game) MoveSnake(s *Snake, other *Snake) {
	if !s.Alive {
		return
	}
	s.Dir = s.NextDir
	np := g.NextPos(s)
	if s.HasStar() {
		if np.X < 0 {
			np.X = MapWidth - 1
		} else if np.X >= MapWidth {
			np.X = 0
		}
		if np.Y < 0 {
			np.Y = MapHeight - 1
		} else if np.Y >= MapHeight {
			np.Y = 0
		}
		for i := range g.Walls {
			if g.Walls[i] == np {
				switch s.Dir {
				case Up:
					np = Point{np.X, MapHeight - 1}
				case Down:
					np = Point{np.X, 0}
				case Left:
					np = Point{MapWidth - 1, np.Y}
				case Right:
					np = Point{0, np.Y}
				}
				break
			}
		}
	}
	if g.CheckCollision(s, other, np) {
		s.Alive = false
		s.CorpseEnd = time.Now().Add(CorpseDuration)
		for _, b := range s.Body {
			g.Corpses = append(g.Corpses, Corpse{
				Pos:       b,
				Color:     s.Color,
				ExpiresAt: s.CorpseEnd,
			})
		}
		return
	}
	s.Body = append([]Point{np}, s.Body...)
	if s.GrowPending > 0 {
		s.GrowPending--
	} else {
		s.Body = s.Body[:len(s.Body)-1]
	}
	g.CheckPickups(s)
}

func (g *Game) CheckPickups(s *Snake) {
	head := s.Head()
	for i := len(g.Foods) - 1; i >= 0; i-- {
		if g.Foods[i].Pos == head {
			s.Score += g.Foods[i].Points
			s.GrowPending += g.Foods[i].Points
			g.Foods = append(g.Foods[:i], g.Foods[i+1:]...)
			g.SpawnFood()
			if rand.Intn(FoodSpawnChance) == 0 {
				g.SpawnFood()
			}
			if rand.Intn(StarSpawnChance) == 0 {
				g.SpawnStar()
			}
			if rand.Intn(DiamondSpawnChance) == 0 && g.Diamond == nil {
				g.SpawnDiamond()
			}
		}
	}
	for i := len(g.Stars) - 1; i >= 0; i-- {
		if g.Stars[i].Pos == head {
			s.StarEnd = time.Now().Add(StarDuration)
			g.Stars = append(g.Stars[:i], g.Stars[i+1:]...)
		}
	}
	if g.Diamond != nil && g.Diamond.Pos == head {
		s.Score += DiamondPoints
		s.GrowPending += DiamondGrow
		name := "P1"
		color := termbox.ColorLightRed
		if s == g.Player2 {
			name = "P2"
			color = termbox.ColorLightBlue
		}
		g.Flash = &FlashMsg{
			Text:      fmt.Sprintf(" %s got ♦ +5! ", name),
			Color:     color,
			ExpiresAt: time.Now().Add(FlashDuration),
		}
		g.Diamond = nil
		g.SpawnDiamond()
	}
}

func (g *Game) Tick() {
	now := time.Now()
	g.MoveSnake(g.Player1, g.Player2)
	g.MoveSnake(g.Player2, g.Player1)
	for i := len(g.Corpses) - 1; i >= 0; i-- {
		if now.After(g.Corpses[i].ExpiresAt) {
			g.Corpses = append(g.Corpses[:i], g.Corpses[i+1:]...)
		}
	}
	if g.Flash != nil && now.After(g.Flash.ExpiresAt) {
		g.Flash = nil
	}
	if !g.Player1.Alive && !g.Player2.Alive {
		g.State = StateGameOver
	}
	if !g.Player1.Alive && time.Now().After(g.Player1.CorpseEnd) {
		g.Player1.CorpseEnd = time.Now().Add(10 * time.Hour)
	}
	if !g.Player2.Alive && time.Now().After(g.Player2.CorpseEnd) {
		g.Player2.CorpseEnd = time.Now().Add(10 * time.Hour)
	}
}

func (g *Game) GetTickRate() time.Duration {
	p1b := g.Player1.HasBoost()
	p2b := g.Player2.HasBoost()
	if p1b && p2b {
		return TickBoost / 2
	}
	if p1b || p2b {
		return TickBoost
	}
	return TickNormal
}
