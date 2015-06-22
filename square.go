package griddler

import (
	"fmt"
)

// Square is the basic element of the grid
type Square struct {
	x, y int
	val  int
}

func NewSquare(x, y, v int, g *Griddler) *Square {
	return &Square{
		x:   x,
		y:   y,
		val: v,
	}
}

func (s Square) show() {
	//fmt.Printf("(%d,%d,", s.x, s.y)
	switch s.val {
	case 0:
		fmt.Printf(" ")
	case 1:
		fmt.Printf(".")
	case 2:
		fmt.Printf("X")
	}
	//fmt.Printf(")")
}
