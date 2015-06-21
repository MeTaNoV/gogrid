package griddler

import (
	"fmt"
)

// Square is the basic element of the grid
type Square struct {
	x, y int
	val  int
	g    *Griddler
}

func NewSquare(x, y, v int, g *Griddler) *Square {
	return &Square{
		x:   x,
		y:   y,
		val: v,
		g:   g,
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

type Stack [](*Square)

func (st *Stack) push(sq *Square) {
	*st = append(*st, sq)
}

func (s *Stack) pop() *Square {
	if len(*s) == 0 {
		return nil
	}
	ret := (*s)[len(*s)-1]
	*s = (*s)[0 : len(*s)-1]
	return ret
}
