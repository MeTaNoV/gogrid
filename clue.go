package griddler

type Clue struct {
	l          *Line
	index      int
	length     int
	begin, end int
	isDone     bool
}

func NewClue(l int) *Clue {
	return &Clue{
		length: l,
		isDone: false,
	}
}

func (c *Clue) solveOverlap() {
	diff := c.begin + c.length - (c.end + 1 - c.length)
	if diff > 0 {
		for j := 0; j < diff; j++ {
			c.l.g.setValue(c.l.squares[c.end-c.length+1+j], 2)
		}
	}
}
