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
		l.isDone = true
	}
}

func (l *Line) incrementClues() {
	l.sumClues++
	if (l.sumBlanks+l.sumClues) == l.length && !l.isDone {
		l.isDone = true
	}
}

func (l *Line) updateCluesLimits(c *Clue, length int, reverse bool) {
	if reverse {
		if c.index == l.ce {
			for i := 0; i < length; i++ {
				l.g.SetValue(l.squares[c.end-i], BLANK)
			}
		}
		l.decrementCluesEnd(c, length)
	} else {
		if c.index == l.cb {
			for i := 0; i < length; i++ {
				l.g.SetValue(l.squares[c.begin+i], BLANK)
			}
		}
		l.incrementCluesBegin(c, length)
	}
}

func (l *Line) incrementCluesBegin(begC *Clue, n int) {
	index := begC.index
	for i := index; i < len(l.clues); i++ {
		switch {
		case i == index:
			l.clues[i].begin += n
			//l.clues[i].print("incrementCluesBegin")
		case i > index:
			// in case the current clue is already further, we will exit
			if l.clues[i-1].begin+l.clues[i-1].length+1 > l.clues[i].begin {
				l.clues[i].begin = l.clues[i-1].begin + l.clues[i-1].length + 1
				//l.clues[i].print("incrementCluesBegin")
			} else {
				return
			}
		}
		if l.clues[i].end-l.clues[i].begin < l.clues[i].length-1 {
			panic(&SolveError{&Square{x: l.index, y: l.index}, ErrInvalidClueSize})
		}
	}
}

func (l *Line) decrementCluesEnd(endC *Clue, n int) {
	index := endC.index
	for i := index; i >= 0; i-- {
		switch {
		case i == index:
			l.clues[i].end -= n
			//l.clues[i].print("decrementCluesEnd")
		case i < index:
			// in case the current clue is already further, we will exit
			if l.clues[i+1].end-l.clues[i+1].length-1 < l.clues[i].end {
				l.clues[i].end = l.clues[i+1].end - l.clues[i+1].length - 1
				//l.clues[i].print("decrementCluesEnd")
			} else {
				return
			}
		}
		if l.clues[i].end-l.clues[i].begin < l.clues[i].length-1 {
			panic(&SolveError{&Square{x: l.index, y: l.index}, ErrInvalidClueSize})
		}
	}
}

func (l *Line) updateClueIndexes(c *Clue) {
	if l.cb != l.ce {
		//l.print("updateClueIndexes")
		if c.index == l.cb {
			l.cb++
		}
		if c.index == l.ce {
			l.ce--
		}
		//l.print("updateClueIndexes")
	}
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

func (l *Line) isSolved(r *Range) bool {
	if r.min == 0 {
		return l.squares[r.max+1].val == BLANK
	}
	if r.max == l.length-1 {
		return l.squares[r.min-1].val == BLANK
	}
	return l.squares[r.max+1].val == BLANK && l.squares[r.min-1].val == BLANK
}

func (l *Line) getAllRanges() [](*Range) {
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

func (l *Line) getUnsolvedRanges() [](*Range) {
	lastVal := EMPTY
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

func (l *Line) getSolvedRanges() [](*Range) {
	lastVal := EMPTY
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

func (l *Line) getEmptyRanges() [](*Range) {
	lastVal := -1
	min, max := 0, 0
	result := make([](*Range), 0)

	// we can start from the first non solved clue +1 up to the last non solved one
	for i := l.clues[l.cb].begin; i <= l.clues[l.ce].end; i++ {
		s := l.squares[i]
		switch {
		case s.val == EMPTY:
			if lastVal != s.val {
				// new one
				min = i
			}
			if i == l.clues[l.ce].end && (min == l.clues[l.cb].begin || l.squares[min-1].val == BLANK) {
				max = i
				result = append(result, &Range{min: min, max: max})
			}
		case s.val == BLANK:
			if lastVal == EMPTY {
				max = i - 1
				if min == l.clues[l.cb].begin || l.squares[min-1].val == BLANK {
					result = append(result, &Range{min: min, max: max})
				}
			}
		case s.val == FILLED:
		}
		lastVal = s.val
	}

	return result
}

func (l *Line) updateCluesForRanges(rs [](*Range)) {
	// the presence of filled Range on a line introduce limit constraints om clues that
	// we are performing on a 2-pass phase from the beginning and from the end

	//l.print("updateCluesForRanges")

	// From beginning
	iClue := l.cb
	iRange := 0
LoopBegin:
	for iRange < len(rs) {
		// if we didn't mapped all clue by this time, this is a puzzle issue
		if iClue > l.ce {
			panic(&SolveError{&Square{x: l.index, y: l.index}, ErrInvalidClueRange})
		}

		c := l.clues[iClue]
		r := rs[iRange]

		if !c.contains(r) {
			iClue++
			continue
		}

		//r.print("updateCluesForRanges Begin")

		switch {
		// if the clue does not fit, we can decrement its end  ......XX.. with (1,2)
		case c.length < r.length():
			//c.print("updateCluesForRanges Begin case 1")
			if c.end > r.min-2 {
				l.decrementCluesEnd(c, c.end-r.min+2)
			}
		case c.length == r.length():
			//c.print("updateCluesForRanges Begin case 2")
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
			// if the range is solved, it is impossible to fit, so we decrement its end and reset
			if l.isSolved(r) {
				//c.print("updateCluesForRanges Begin case 4")
				if c.end > r.min-2 {
					l.decrementCluesEnd(c, c.end-r.min+2)
				}
				//Pause()
				iClue = l.cb
				iRange = 0
				continue LoopBegin
			} else {
				//c.print("updateCluesForRanges Begin case 5")
				// we try to find a set of ranges that fit
				var concat *Range
				iRange++
				if iRange < len(rs) {
					concat = &Range{min: r.min, max: rs[iRange].max}
					r = rs[iRange]
				} else {
					break
				}
				if concat.max == c.end {
					iRange++
					if iRange < len(rs) {
						r = rs[iRange]
					}
					iClue++
					continue LoopBegin
				}
				for concat.max < c.end {
					if l.isSolved(concat) {
						//concat.print("updateCluesForRanges Begin case 6")
						if c.end > concat.min-2 {
							l.decrementCluesEnd(c, c.end-concat.min+2)
						}
						iClue = l.cb
						iRange = 0
						continue LoopBegin
					}
					//concat.print("updateCluesForRanges Begin case 7")
					iRange++
					if iRange < len(rs) {
						concat = &Range{min: r.min, max: rs[iRange].max}
						r = rs[iRange]
					} else {
						break
					}
					if concat.max == c.end {
						iRange++
						if iRange < len(rs) {
							r = rs[iRange]
						}
						iClue++
						continue LoopBegin
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
		// if we didn't mapped all clue by this time, this is a puzzle issue
		if iClue < l.cb {
			panic(&SolveError{&Square{x: l.index, y: l.index}, ErrInvalidClueRange})
		}

		c := l.clues[iClue]
		r := rs[iRange]

		if !c.contains(r) {
			iClue--
			continue
		}

		//r.print("updateCluesForRanges End")

		switch {
		// if the clue does not fit, we can decrement its end  ......XX.. with (1,2)
		case c.length < r.length():
			//c.print("updateCluesForRanges End case 1")
			if c.begin < r.max+2 {
				l.incrementCluesBegin(c, r.max+2-c.begin)
			}
		case c.length == r.length():
			//c.print("updateCluesForRanges End case 2")
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
			// if the range is solved, it is impossible to fit, so we decrement its end and reset
			if l.isSolved(r) {
				//c.print("updateCluesForRanges End case 4")
				if c.begin < r.max+2 {
					l.incrementCluesBegin(c, r.max+2-c.begin)
				}
				//Pause()
				iClue = l.ce
				iRange = len(rs) - 1
				continue LoopEnd
			} else {
				//c.print("updateCluesForRanges End case 5")
				// we try to find a set of ranges that fit
				var concat *Range
				iRange--
				if iRange >= 0 {
					concat = &Range{min: rs[iRange].min, max: r.max}
					r = rs[iRange]
				} else {
					break
				}
				if concat.min == c.begin {
					iRange--
					if iRange >= 0 {
						r = rs[iRange]
					}
					iClue--
					continue LoopEnd
				}
				for concat.min > c.begin {
					if l.isSolved(concat) {
						//concat.print("updateCluesForRanges End case 6")
						if c.begin < concat.max+2 {
							l.incrementCluesBegin(c, concat.max+2-c.begin)
						}
						iClue = l.ce
						iRange = len(rs) - 1
						continue LoopEnd
					}
					//concat.print("updateCluesForRanges End case 7")
					iRange--
					if iRange >= 0 {
						concat = &Range{min: rs[iRange].min, max: r.max}
						r = rs[iRange]
					} else {
						break
					}
					if concat.min == c.begin {
						iRange--
						if iRange >= 0 {
							r = rs[iRange]
						}
						iClue--
						continue LoopEnd
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

func (l *Line) getCluesForEmptyRange(r *Range) [](*Clue) {
	cs := make([](*Clue), 0)
	for _, c := range l.clues[l.cb : l.ce+1] {
		if c.begin <= r.min && c.end > r.min {
			cs = append(cs, c)
		}
		if c.begin < r.max && c.end >= r.max {
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
