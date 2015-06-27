package griddler

import (
	"fmt"
)

type Clue struct {
	l          *Line
	index      int
	length     int
	begin, end int
	// isDone     bool
}

func NewClue(l int) *Clue {
	return &Clue{
		length: l,
	}
}

func (c *Clue) print(prefix string) {
	fmt.Printf("%s-->Clue(i:%d,b:%d,e:%d,l:%d)\n", prefix, c.index+1, c.begin+1, c.end+1, c.length)
}

func (c *Clue) solveOverlap() {
	diff := c.begin + c.length - (c.end + 1 - c.length)
	if diff > 0 {
		for j := 0; j < diff; j++ {
			c.l.g.SetValue(c.l.squares[c.end-c.length+1+j], FILLED)
		}
	}
}

func (c *Clue) solveConstraints(reverse bool) {
	empty := 0
	filled := 0
	l := c.l
	i := c.begin
	if reverse {
		i = c.end
	}
	for {
		switch {
		case l.squares[i].val == EMPTY:
			empty++
		case l.squares[i].val == BLANK:
			if (empty + filled) < c.length {
				l.updateCluesLimits(c, empty+filled+1, reverse)
				empty = 0
				filled = 0
			}
		case l.squares[i].val == FILLED:
			filled++
		}
		i = IncOrDec(i, reverse)
		if i < c.begin || i > c.end {
			return
		}
	}
}

func (c *Clue) solveCompleteness() {
	if c.end-c.begin == c.length-1 {
		if c.begin > 0 {
			c.l.g.SetValue(c.l.squares[c.begin-1], BLANK)
		}
		if c.end < c.l.length-1 {
			c.l.g.SetValue(c.l.squares[c.end+1], BLANK)
		}
		// update line clue indexes
		c.l.updateClueIndexes(c)
	}
}

func (c *Clue) contains(r *Range) bool {
	return c.begin <= r.min && c.end >= r.max
}
