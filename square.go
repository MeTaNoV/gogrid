package griddler

import (
	"fmt"
)

const (
	EMPTY = iota
	BLANK
	FILLED
)

// Square is the basic element of the grid
type Square struct {
	x, y int
	val  int
}

func NewSquare(x, y, v int) *Square {
	return &Square{
		x:   x,
		y:   y,
		val: v,
	}
}

func (s Square) show() {
	//fmt.Printf("(%d,%d,", s.x, s.y)
	switch s.val {
	case EMPTY:
		fmt.Printf(" ")
	case BLANK:
		fmt.Printf(".")
	case FILLED:
		fmt.Printf("X")
	}
	//fmt.Printf(")")
}
