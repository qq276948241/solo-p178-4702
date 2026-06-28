package main

import (
	"github.com/nsf/termbox-go"
)

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
