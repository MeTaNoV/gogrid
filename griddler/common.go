package griddler

import "fmt"

type Tile struct {
	value int
}

type Solver interface {
	Solve() bool
	SetValue(square *Square, value int)
}

// utility struc Range
type Range struct {
	min, max int
}

func (r *Range) length() int {
	return r.max - r.min + 1
}

func (r *Range) print(prefix string) {
	fmt.Printf("%s-->Range(b:%d,e:%d)\n", prefix, r.min+1, r.max+1)
}
