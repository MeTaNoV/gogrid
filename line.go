package griddler

import (
	"fmt"
)

type Line struct {
	g          *Griddler
	index      int
	length     int
	clues      [](*Clue)
	squares    [](*Square)
	sumBlanks  int
	sumClues   int // current sum of all clue values
	totalClues int // total sum of all clues evaluated
	cb, ce     int // indexes of the first and last non solved clue
	isDone     bool
}

type Stack [](*Line)

func (st *Stack) push(nl *Line) {
	for _, l := range *st {
		if l == nl {
			return
		}
	}
	*st = append(*st, nl)
}

func (st *Stack) pop() *Line {
	if len(*st) == 0 {
		return nil
	}
	ret := (*st)[len(*st)-1]
	*st = (*st)[0 : len(*st)-1]
	return ret
}

func NewLine(g *Griddler, index, length int) *Line {
	return &Line{
		g:          g,
		index:      index,
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
		fmt.Printf("Line/Column %d solved!!!\n", l.index+1)
		l.isDone = true
		l.g.incrementSolvedLines()
	}
}

func (l *Line) incrementClues() {
	l.sumClues++
	if (l.sumBlanks+l.sumClues) == l.length && !l.isDone {
		fmt.Printf("Line/Column %d solved!!!\n", l.index+1)
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
	if reverse {
		if c.index > l.cb {
			l.ce--
		}
	} else {
		if c.index < l.ce {
			l.cb++
		}
	}
}

func (l *Line) updateCluesRanges(c *Clue, length int, reverse bool) {
	if reverse {
		l.decrementCluesEnd(c.index, length)
	} else {
		l.incrementCluesBegin(c.index, length)
	}
}

func (l *Line) checkRange(value int, min, max int) bool {
	if min < 0 || max >= l.length {
		return false
	}
	for i := min; i <= max; i++ {
		if l.squares[i].val != value {
			return false
		}
	}
	return true
}

func (l *Line) unsolvedGroups() [](*Range) {
	lastVal := 0
	min, max := 0, 0
	result := make([](*Range), 0)

	// we can start from the first non solved clue +1 up to the last non solved one
	for i := l.clues[l.cb].begin + 1; i <= l.clues[l.ce].end; i++ {
		s := l.squares[i]
		switch {
		case s.val == 0, s.val == 1:
			if lastVal == 2 {
				max = i - 1
				if l.squares[min-1].val != 1 || l.squares[max+1].val != 1 {
					result = append(result, &Range{min: min, max: max})
				}
			}
		case s.val == 2:
			if lastVal != s.val {
				// new one
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

func (l *Line) getPotentialCluesForRange(r *Range) [](*Clue) {
	cs := make([](*Clue), 0)
	for _, c := range l.clues[l.cb : l.ce+1] {
		if c.begin <= r.min && c.end >= r.max && c.length >= r.length() {
			cs = append(cs, c)
		}
	}

	return cs
}

func (l *Line) getExactCluesForRange(r *Range) [](*Clue) {
	cs := make([](*Clue), 0)
	for _, c := range l.clues[l.cb : l.ce+1] {
		if c.begin <= r.min && c.end >= r.max && c.length == r.length() {
			cs = append(cs, c)
		}
	}

	return cs
}

func (l *Line) getStepToNextBlank(r *Range, reverse bool) (bool, int) {
	var i int

	if reverse {
		i = r.min - 1
	} else {
		i = r.max + 1
	}

	result := 0
	for i >= 0 && i < l.length {
		s := l.squares[i]
		switch {
		case s.val == 0:
			result++
		case s.val == 1:
			return true, result
		case s.val == 2:
			return false, result
		}
		i = IncOrDec(i, reverse)
	}
	return false, result
}

// TODO refactor this function...
func (l *Line) check1to1Mapping(rs [](*Range)) bool {
	// first we check the mapping in increasing order
	//fmt.Printf("\nA4 Checking begin:")
	c := l.clues[l.cb] // current clue
	mapping := make(map[*Clue](*Range), len(l.clues))
	for i := 0; i < len(rs); i++ {
		r := rs[i]
		// fmt.Printf("\nClue(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
		// fmt.Printf("\nRange(b:%d,e:%d):", r.min, r.max)
		if mapping[c] != nil {
			if c.length < r.max-mapping[c].min+1 {
				// if we didn't reach the last available clue
				if c.index < l.ce {
					c = l.clues[c.index+1]
					//fmt.Printf("\nNextClue(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
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
		//fmt.Printf("\nA4 Checking end:")
		c = l.clues[l.ce] // current clue
		mapping := make(map[*Clue](*Range), len(l.clues))
		for i := len(rs); i > 0; i-- {
			r := rs[i-1]
			// fmt.Printf("\nClue(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
			// fmt.Printf("\nRange(b:%d,e:%d):", r.min, r.max)
			if mapping[c] != nil {
				if c.length < mapping[c].max-r.min+1 {
					// if we didn't reach the last available clue
					if c.index > l.cb {
						c = l.clues[c.index-1]
						//fmt.Printf("\nNextClue(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
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

	//fmt.Printf("\nA4 Checking result (cb:%d, ce:%d, ci:%d): %t", l.cb, l.ce, c.index, c.index == l.cb)
	return c.index == l.cb
}
