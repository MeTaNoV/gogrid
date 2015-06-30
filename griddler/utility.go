package griddler

import (
	"fmt"
	"math"
	"os"
)

type Tile struct {
	value int
}

type Solver interface {
	Solve() bool
	SetValue(square *Square, value int)
}

// utility function to pause and wait for the user to press enter
func Pause() {
	var b []byte = make([]byte, 2)
	os.Stdin.Read(b)
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func maxLength(cs [](*Clue)) int {
	result := 0
	for _, c := range cs {
		result = max(result, c.length)
	}
	return result
}

func minLength(cs [](*Clue)) int {
	result := math.MaxUint8 // support for lines of max length 256
	for _, c := range cs {
		result = min(result, c.length)
	}
	return result
}

func IncOrDec(i int, reverse bool) int {
	if reverse {
		i--
	} else {
		i++
	}
	return i
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

// Special stack implementation where elements are unique
type lStack [](*Line)

func (st *lStack) push(nste *Line) {
	for _, ste := range *st {
		if ste == nste {
			return
		}
	}
	*st = append(*st, nste)
}

func (st *lStack) pop() *Line {
	if len(*st) == 0 {
		return nil
	}
	ret := (*st)[len(*st)-1]
	*st = (*st)[0 : len(*st)-1]
	return ret
}
