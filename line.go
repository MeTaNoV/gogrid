package griddler

import (
	"fmt"
)

type Line struct {
	g          *Griddler
	length     int
	clues      [](*Clue)
	squares    [](*Square)
	sumBlanks  int
	sumClues   int // current sum of all clue values
	totalClues int // total sum of all clues evaluated
	cb, ce     int // indexes of the first and last non solved clue
	isDone     bool
}

func NewLine(g *Griddler, length int) *Line {
	return &Line{
		g:          g,
		length:     length,
		squares:    make([](*Square), length),
		sumBlanks:  0,
		sumClues:   0,
		totalClues: 0,
		isDone:     false,
	}
}

func (l *Line) addClues(cs [](*Clue)) {
	l.clues = cs
	for i, val := range cs {
		l.totalClues += val.length
		val.l = l
		val.index = i
	}
	l.cb = 0
	l.ce = len(cs) - 1
}

func (l *Line) incrementBlanks() {
	l.sumBlanks++
	if (l.sumBlanks+l.sumClues) == l.length && !l.isDone {
		l.isDone = true
		l.g.incrementSolvedLines()
	}
}

func (l *Line) incrementClues() {
	l.sumClues++
	if (l.sumBlanks+l.sumClues) == l.length && !l.isDone {
		l.isDone = true
		l.g.incrementSolvedLines()
	}
}

func (l *Line) incrementCluesBegin(index, n int) {
	for i := index; i < len(l.clues); i++ {
		switch {
		case i == index:
			l.clues[i].begin += n
		case i > index:
			// in case the current clue is already further, we will exit
			if l.clues[i-1].begin+l.clues[i-1].length+1 > l.clues[i].begin {
				l.clues[i].begin = l.clues[i-1].begin + l.clues[i-1].length + 1
			} else {
				return
			}
		}
	}
}

func (l *Line) decrementCluesEnd(index, n int) {
	for i := index; i >= 0; i-- {
		switch {
		case i == index:
			l.clues[i].end -= n
		case i < index:
			// in case the current clue is already further, we will exit
			if l.clues[i+1].end-l.clues[i+1].length-1 < l.clues[i].end {
				l.clues[i].end = l.clues[i+1].end - l.clues[i+1].length - 1
			} else {
				return
			}
		}
	}
}

func (l *Line) updateCluesIndexes(c *Clue, reverse bool) {
	fmt.Printf("Line clue range: cb:%d, ce:%d\n", l.cb, l.ce)
	if reverse {
		if c.index > l.cb {
			l.ce--
		}
	} else {
		if c.index < l.ce {
			l.cb++
		}
	}
	fmt.Printf("Line clue range: cb:%d, ce:%d\n", l.cb, l.ce)
	Pause()
}

func (l *Line) updateCluesRanges(c *Clue, length int, reverse bool) {
	if reverse {
		l.decrementCluesEnd(c.index, length)
		fmt.Printf("\nNewClue(n:%d,b:%d,e:%d,l:%d):\n", c.index+1, c.begin+1, c.end+1, c.length)
	} else {
		l.incrementCluesBegin(c.index, length)
		fmt.Printf("\nNewClue(n:%d,b:%d,e:%d,l:%d):\n", c.index+1, c.begin+1, c.end+1, c.length)
	}
}

func (l *Line) filledGroups() [](*Range) {
	lastVal := 0
	min, max := 0, 0
	result := make([](*Range), 0)

	// we can start from the first non solved clue up to the last non solved one
	for i := l.clues[l.cb].begin; i <= l.clues[l.ce].end; i++ {
		s := l.squares[i]
		switch {
		case s.val == 0, s.val == 1:
			if lastVal == 2 {
				max = i - 1
				result = append(result, &Range{min: min, max: max})
			}
		case s.val == 2:
			if lastVal != s.val {
				min = i
			}
		}
		lastVal = s.val
	}

	return result
}

func (l *Line) solvedGroups() [](*Range) {
	lastVal := 0
	min, max := 0, 0
	result := make([](*Range), 0)
	blankBefore := 0

	// we can start from the first non solved clue up to the last non solved one
	for i := l.clues[l.cb].begin; i <= l.clues[l.ce].end; i++ {
		// we are looking for a 0XXX0 pattern
		s := l.squares[i]
		switch {
		case s.val == 0:
			blankBefore = 0
			min, max = 0, 0
		case s.val == 1:
			if blankBefore == 0 {
				blankBefore = 1
				break
			}
			if lastVal == 2 {
				max = i - 1
				result = append(result, &Range{min: min, max: max})
			}
		case s.val == 2:
			if lastVal != s.val {
				min = i
			}
		}
		lastVal = s.val
	}

	return result
}

func (l *Line) checkMapping(rs [](*Range)) bool {
	fmt.Printf("\nA4 Checking begin:")
	// first we check the mapping in increasing order
	c := l.clues[l.cb] // current clue
	mapping := make(map[*Clue](*Range), len(l.clues))
	for i := 0; i < len(rs); i++ {
		r := rs[i]
		fmt.Printf("\nClue(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
		fmt.Printf("\nRange(b:%d,e:%d):", r.min, r.max)
		if mapping[c] != nil {
			if c.length < r.max-mapping[c].min+1 {
				// if we didn't reach the last available clue
				if c.index < l.ce {
					c = l.clues[c.index+1]
					fmt.Printf("\nNextClue(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
				} else {
					return false
				}
			} else {
				// TODO: think about how to deal with a non perfect mapping...
				return false
			}
		} else {
			mapping[c] = r
		}
	}

	// if it was successful, we check in decreasing order
	if c.index == l.ce {
		fmt.Printf("\nA4 Checking end:")
		c = l.clues[l.ce] // current clue
		mapping := make(map[*Clue](*Range), len(l.clues))
		for i := len(rs); i > 0; i-- {
			r := rs[i-1]
			fmt.Printf("\nClue(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
			fmt.Printf("\nRange(b:%d,e:%d):", r.min, r.max)
			if mapping[c] != nil {
				if c.length < mapping[c].max-r.min+1 {
					// if we didn't reach the last available clue
					if c.index > l.cb {
						c = l.clues[c.index-1]
						fmt.Printf("\nNextClue(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
					} else {
						return false
					}
				} else {
					// TODO: think about how to deal with a non perfect mapping...
					return false
				}
			} else {
				mapping[c] = r
			}
		}
	} else {
		return false
	}

	fmt.Printf("\nA4 Checking result (cb:%d, ce:%d, ci:%d): %t", l.cb, l.ce, c.index, c.index == l.cb)
	return c.index == l.cb
}
