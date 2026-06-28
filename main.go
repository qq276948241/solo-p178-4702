package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/nsf/termbox-go"
)

const (
	MapWidth  = 24
	MapHeight = 16

	TickNormal = 120 * time.Millisecond
	TickBoost  = 55 * time.Millisecond

	StarDuration     = 2000 * time.Millisecond
	CorpseDuration   = 3000 * time.Millisecond
	FoodSpawnChance  = 7
	StarSpawnChance  = 15
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
	State         GameState
	MapID         int
	Walls         []Point
	Player1       *Snake
	Player2       *Snake
	Foods         []Food
	Stars         []Star
	Corpses       []Corpse
	LastTick      time.Time
	SelectedMap   int
	MenuOptions   []string
}

var mapConfigs = [][]Point{
	{
		{6, 4}, {7, 4}, {8, 4},
		{15, 11}, {16, 11}, {17, 11},
		{11, 7}, {12, 7}, {11, 8}, {12, 8},
	},
	{
		{5, 2}, {5, 3}, {5, 4}, {5, 5},
		{18, 10}, {18, 11}, {18, 12}, {18, 13},
		{10, 3}, {11, 3}, {12, 3}, {13, 3},
		{10, 12}, {11, 12}, {12, 12}, {13, 12},
		{3, 8}, {20, 7},
	},
	{
		{0, 0}, {1, 0}, {MapWidth - 1, 0}, {MapWidth - 2, 0},
		{0, MapHeight - 1}, {1, MapHeight - 1}, {MapWidth - 1, MapHeight - 1}, {MapWidth - 2, MapHeight - 1},
		{7, 0}, {7, 1}, {7, 2},
		{16, MapHeight - 3}, {16, MapHeight - 2}, {16, MapHeight - 1},
		{11, 5}, {12, 5}, {13, 5},
		{11, 10}, {12, 10}, {13, 10},
		{3, 7}, {3, 8},
		{20, 7}, {20, 8},
	},
}

var mapNames = []string{
	"Cross Roads",
	"Cornered",
	"Fortress",
}

func NewSnake(startX, startY int, dir Direction, color, headColor termbox.Attribute) *Snake {
	body := make([]Point, 3)
	for i := 0; i < 3; i++ {
		switch dir {
		case Right:
			body[i] = Point{startX - i, startY}
		case Left:
			body[i] = Point{startX + i, startY}
		case Up:
			body[i] = Point{startX, startY + i}
		case Down:
			body[i] = Point{startX, startY - i}
		}
	}
	return &Snake{
		Body:      body,
		Dir:       dir,
		NextDir:   dir,
		Color:     color,
		HeadColor: headColor,
		Alive:     true,
	}
}

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

const BoostDuration = 250 * time.Millisecond

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
		}
	}
	for i := len(g.Stars) - 1; i >= 0; i-- {
		if g.Stars[i].Pos == head {
			s.StarEnd = time.Now().Add(StarDuration)
			g.Stars = append(g.Stars[:i], g.Stars[i+1:]...)
		}
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

func drawCell(x, y int, ch rune, fg, bg termbox.Attribute) {
	termbox.SetCell(x*2, y, ch, fg, bg)
	termbox.SetCell(x*2+1, y, ' ', fg, bg)
}

func (g *Game) Render() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	scoreLine := fmt.Sprintf(" P1(Red): %d  |  P2(Blue): %d ", g.Player1.Score, g.Player2.Score)
	ctrlLine := " P1: WASD + LShift  |  P2: Arrows + Space  |  P: Pause  R: Restart "
	for i, c := range scoreLine {
		termbox.SetCell(i, 0, c, termbox.ColorWhite|termbox.AttrBold, termbox.ColorBlack)
	}
	for i, c := range ctrlLine {
		termbox.SetCell(i, 1, c, termbox.ColorDarkGray, termbox.ColorBlack)
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
		"  Q (hold/tap)      Space (tap)   = Boost",
		"",
		"Rules:",
		"  ● = +1 pt    ◆ = +3 pts",
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
			if li >= 5 && li <= 9 {
				fg = termbox.ColorLightYellow
			}
			termbox.SetCell(lx+i, insY+li, c, fg, termbox.ColorBlack)
		}
	}
	_ = h
	termbox.Flush()
}

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
	case termbox.KeyCtrlA:
	case termbox.KeyCtrlR:
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
		} else if g.State == StatePaused {
			g.State = StatePlaying
		}
	case 'r', 'R':
		*g = *NewGame(g.MapID)
	case 'm', 'M':
		*g = *NewMenu()
	}
	return true
}

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
