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

func (l *Line) print(prefix string) {
	fmt.Printf("%s-->Line: cb:%d, ce:%d\n", prefix, l.cb+1, l.ce+1)
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

func (l *Line) updateCluesLimits(c *Clue, length int, reverse bool) {
	if reverse {
		l.decrementCluesEnd(c, length)
	} else {
		l.incrementCluesBegin(c, length)
	}
}

func (l *Line) incrementCluesBegin(begC *Clue, n int) {
	index := begC.index
	for i := index; i < len(l.clues); i++ {
		switch {
		case i == index:
			l.clues[i].begin += n
			l.clues[i].print("incrementCluesBegin")
		case i > index:
			// in case the current clue is already further, we will exit
			if l.clues[i-1].begin+l.clues[i-1].length+1 > l.clues[i].begin {
				l.clues[i].begin = l.clues[i-1].begin + l.clues[i-1].length + 1
				l.clues[i].print("incrementCluesBegin")
			} else {
				return
			}
		}
	}
}

func (l *Line) decrementCluesEnd(endC *Clue, n int) {
	index := endC.index
	for i := index; i >= 0; i-- {
		switch {
		case i == index:
			l.clues[i].end -= n
			l.clues[i].print("decrementCluesEnd")
		case i < index:
			// in case the current clue is already further, we will exit
			if l.clues[i+1].end-l.clues[i+1].length-1 < l.clues[i].end {
				l.clues[i].end = l.clues[i+1].end - l.clues[i+1].length - 1
				l.clues[i].print("decrementCluesEnd")
			} else {
				return
			}
		}
	}
}

func (l *Line) updateCluesIndexes(c *Clue) {
	if c.index == l.cb {
		l.cb++
	}
	if c.index == l.ce {
		l.ce--
	}
	l.print("updateCluesIndexes")
}

func (l *Line) checkRangeForValue(value int, min, max int) bool {
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

func (l *Line) getRanges() [](*Range) {
	lastVal := EMPTY
	min, max := 0, 0
	result := make([](*Range), 0)

	// we can start from the first non solved clue +1 up to the last non solved one
	// alog 1 is taken care of the case when the first or last square is filled
	for i := l.clues[l.cb].begin; i <= l.clues[l.ce].end; i++ {
		s := l.squares[i]
		switch {
		case s.val == EMPTY, s.val == BLANK:
			if lastVal == FILLED {
				max = i - 1
				result = append(result, &Range{min: min, max: max})
			}
		case s.val == FILLED:
			if s.val != lastVal {
				// new one
				min = i
			}
			if i == l.clues[l.ce].end {
				max = i
				result = append(result, &Range{min: min, max: max})
			}
		}
		lastVal = s.val
	}

	return result
}

func (l *Line) unsolvedRanges() [](*Range) {
	lastVal := 0
	min, max := 0, 0
	result := make([](*Range), 0)

	// we can start from the first non solved clue +1 up to the last non solved one
	for i := l.clues[l.cb].begin; i <= l.clues[l.ce].end; i++ {
		s := l.squares[i]
		switch {
		case s.val == EMPTY, s.val == BLANK:
			if lastVal == FILLED {
				max = i - 1
				if l.squares[min-1].val != BLANK || l.squares[max+1].val != BLANK {
					result = append(result, &Range{min: min, max: max})
				}
			}
		case s.val == FILLED:
			if lastVal != s.val {
				// new one
				min = i
			}
			if i == l.clues[l.ce].end {
				max = i
				if l.squares[min-1].val != BLANK {
					result = append(result, &Range{min: min, max: max})
				}
			}
		}
		lastVal = s.val
	}

	return result
}

func (l *Line) solvedRanges() [](*Range) {
	lastVal := 0
	min, max := 0, 0
	result := make([](*Range), 0)

	// we can start from the first non solved clue +1 up to the last non solved one
	for i := l.clues[l.cb].begin; i <= l.clues[l.ce].end; i++ {
		s := l.squares[i]
		switch {
		case s.val == EMPTY:
		case s.val == BLANK:
			if lastVal == FILLED {
				max = i - 1
				if min == l.clues[l.cb].begin || l.squares[min-1].val == BLANK {
					result = append(result, &Range{min: min, max: max})
				}
			}
		case s.val == FILLED:
			if lastVal != s.val {
				// new one
				min = i
			}
			if i == l.clues[l.ce].end {
				max = i
				result = append(result, &Range{min: min, max: max})
			}
		}
		lastVal = s.val
	}

	return result
}

func (l *Line) updateCluesForRanges(rs [](*Range)) {
	// the presence of filled Range on a line introduce limit constraints om clues that
	// we are performing on a 2-pass phase from the beginning and from the end

	// From beginning
	iClue := l.cb
	iRange := 0
LoopBegin:
	for iRange < len(rs) {
		c := l.clues[iClue]
		r := rs[iRange]

		if !c.contains(r) {
			iClue++
			continue
		}

		r.print("updateCluesForRanges Begin")
		c.print("updateCluesForRanges Begin")

		switch {
		// if the clue does not fit, we can decrement its end  ......XX.. with (1,2)
		case c.length < r.length():
			if c.end > r.min+2 {
				l.decrementCluesEnd(c, c.end-r.min-2)
			}
		case c.length == r.length():
			// if it fits exactly, we can decrement its end
			if c.end > r.min+c.length-1 {
				l.decrementCluesEnd(c, c.end-(r.min+c.length-1))
			}
			iRange++
		case c.length > r.length():
			// we can decrement its end
			if c.end > r.min+c.length-1 {
				l.decrementCluesEnd(c, c.end-(r.min+c.length-1))
			}
			if iRange == len(rs)-1 {
				iRange++
				continue LoopBegin
			}
			// if the range is solved, it is impossible to fit, so we decrement its end and reset
			if l.squares[r.max+1].val == BLANK {
				if c.end > r.min+2 {
					l.decrementCluesEnd(c, c.end-r.min-2)
				}
				Pause()
				iClue = l.cb
				iRange = 0
				continue LoopBegin
			} else {
				// we try to find a set of ranges that fit
				for r.max <= c.end {
					if iRange == len(rs)-1 {
						iRange++
						continue LoopBegin
					}
					if l.squares[r.max+1].val == BLANK {
						if c.end > r.min+2 {
							l.decrementCluesEnd(c, c.end-r.min-2)
						}
						Pause()
						iClue = l.cb
						iRange = 0
						continue LoopBegin
					}
					iRange++
					if iRange < len(rs) {
						r = rs[iRange]
					} else {
						break
					}
				}
			}
		}
		iClue++
	}

	// From end
	iClue = l.ce
	iRange = len(rs) - 1
LoopEnd:
	for iRange >= 0 {
		c := l.clues[iClue]
		r := rs[iRange]

		if !c.contains(r) {
			iClue--
			continue
		}

		r.print("updateCluesForRanges End")
		c.print("updateCluesForRanges End")

		switch {
		// if the clue does not fit, we can decrement its end  ......XX.. with (1,2)
		case c.length < r.length():
			if c.begin < r.max-2 {
				l.incrementCluesBegin(c, r.max-2-c.begin)
			}
		case c.length == r.length():
			// if it fits exactly, we can decrement its end
			if c.begin < r.max-c.length+1 {
				l.incrementCluesBegin(c, r.max-c.length+1-c.begin)
			}
			iRange--
		case c.length > r.length():
			// we can decrement its end
			if c.begin < r.max-c.length+1 {
				l.incrementCluesBegin(c, r.max-c.length+1-c.begin)
			}
			if iRange == 0 {
				iRange--
				continue LoopEnd
			}
			// if the range is solved, it is impossible to fit, so we decrement its end and reset
			if l.squares[r.min-1].val == BLANK {
				if c.begin > r.max-2 {
					l.incrementCluesBegin(c, r.max-2-c.begin)
				}
				Pause()
				iClue = l.ce
				iRange = len(rs) - 1
				continue LoopEnd
			} else {
				// we try to find a set of ranges that fit
				for r.min >= c.begin {
					if iRange == 0 {
						iRange--
						continue LoopEnd
					}
					if l.squares[r.min-1].val == BLANK {
						if c.begin > r.max-2 {
							l.incrementCluesBegin(c, r.max-2-c.begin)
						}
						Pause()
						iClue = l.ce
						iRange = len(rs) - 1
						continue LoopEnd
					}
					iRange--
					if iRange >= 0 {
						r = rs[iRange]
					} else {
						break
					}
				}
			}
		}
		iClue--
	}
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
