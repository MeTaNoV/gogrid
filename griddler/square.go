package griddler

import (
	"fmt"
)

const (
	EMPTY = iota
	BLANK
	FILLED
)

// Square is the basic element of the grid, it inherits val from Tile
type Square struct {
	Tile
	x, y int
}

func NewSquare(x, y, v int) *Square {
	return &Square{
		Tile{v},
		x,
		y,
	}
}

func (s Square) show() {
	//fmt.Printf("(%d,%d,", s.x, s.y)
	switch s.value {
	case EMPTY:
		fmt.Printf(" ")
	case BLANK:
		fmt.Printf(".")
	case FILLED:
		fmt.Printf("X")
	}
	//fmt.Printf(")")
}

type PrioSquare struct {
	*Square
	pvalue   int
	priority int
}

type prioQueue [](*PrioSquare)

func (pq prioQueue) Len() int {
	return len(pq)
}

func (pq prioQueue) Less(i, j int) bool {
	return pq[i].priority > pq[j].priority
}

func (pq prioQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *prioQueue) Push(x interface{}) {
	item := x.(*PrioSquare)
	*pq = append(*pq, item)
}

func (pq *prioQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}
