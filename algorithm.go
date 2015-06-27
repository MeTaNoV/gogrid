package griddler

import (
//"fmt"
)

// user-defined function to define solving algorithm
type Algorithm func(g *Griddler, l *Line)

// algo to be used to solve basic case (empty/full) and initialize clue range
func solveInitAlgo(g *Griddler, l *Line) {
	switch {
	// no clues are defined for the line, we can blank everything
	case l.totalClues == 0:
		for _, s := range l.squares {
			g.SetValue(s, BLANK)
		}
	// the total of the clue is equal to the line length (i.e. one big clue is defined)
	case l.totalClues == l.length:
		for _, s := range l.squares {
			g.SetValue(s, FILLED)
		}
	// we initialized all clue ranges and solve the overlap
	default:
		for i, clue := range l.clues {
			sumBegin := 0
			for _, c := range l.clues[0:i] {
				sumBegin += c.length + 1
			}
			clue.begin = sumBegin
			sumEnd := 0
			for _, c := range l.clues[i+1:] {
				sumEnd += c.length + 1
			}
			clue.end = l.length - 1 - sumEnd
			clue.solveOverlap()
		}
	}
}

// for each range of filled block on the line, try to determine the associated clue and
// update relevant range information, then solveOverlap
func solveFilledRanges(g *Griddler, l *Line) {
	//l.print("solveFilledRanges")

	for _, c := range l.clues {
		//c.print("solveFilledRanges")
		c.solveConstraints(true)
		c.solveConstraints(false)
		c.solveOverlap()
		c.solveCompleteness()
	}

	rs := l.getAllRanges()
	l.updateCluesForRanges(rs)

	for _, r := range rs {
		//r.print("solveFilledRanges")

		cs := l.getPotentialCluesForRange(r)

		switch {
		case len(cs) == 0:
			// TODO throw error
		case len(cs) == 1:
			c := cs[0]
			//c.print("solveFilledRanges")
			if c.index == l.cb {
				// TODO, here we start at 0, because c.begin is already updated, else
				// we can move the blank process in updateCluesForRanges()
				for i := 0; i < r.max-c.length+1; i++ {
					g.SetValue(l.squares[i], BLANK)
				}
			}
			if c.begin < r.max-c.length+1 {
				l.incrementCluesBegin(c, r.max-c.length+1-c.begin)
			}
			if c.index == l.ce {
				// TODO, here we start at l.length-1, because c.end is already updated, else
				// we can move the blank process in updateCluesForRanges()
				for i := l.length - 1; i > r.min+c.length-1; i-- {
					g.SetValue(l.squares[i], BLANK)
				}
			}
			if c.end > r.min+c.length-1 {
				l.decrementCluesEnd(c, c.end-r.min-c.length+1)
			}
			//c.print("solveFilledRanges")
			c.solveConstraints(true)
			c.solveConstraints(false)
			c.solveOverlap()
			c.solveCompleteness()
		case len(cs) > 1:
			// for _, c := range cs {
			// 	c.print("solveFilledRanges")
			// }
			// if all potential clues are of the Range size, we can finish it
			if maxLength(cs) == r.length() {
				g.SetValue(l.squares[r.min-1], BLANK)
				g.SetValue(l.squares[r.max+1], BLANK)
			}
			// we can increment or decrement ??? It should be done in updateCluesForRanges() already...
			// TODO how can we blank beginning or end ?!?
		}
	}
}

func solveEmptyRanges(g *Griddler, l *Line) {
	//l.print("solveEmptyRanges")
	// first, we can handle the square not covered by any clue anymore
	i := 0
	for _, c := range l.clues {
		//c.print("solveEmptyRanges 1")
		if c.begin > i {
			//Pause()
			for j := i; j < c.begin; j++ {
				g.SetValue(l.squares[j], BLANK)
			}
		}
		i = c.end + 1
	}
	for j := i; j < l.length; j++ {
		g.SetValue(l.squares[j], BLANK)
	}

	// second, we can take a look at *real* empty ranges (e.g. ..0..0.. or border inclusive)
	rs := l.getEmptyRanges()

	for _, r := range rs {
		cs := l.getCluesForEmptyRange(r)

		// if there is no candidate, we can blank
		if len(cs) > 0 && minLength(cs) > r.length() {
			//r.print("solveEmptyRanges")
			//Pause()
			for i := r.min; i <= r.max; i++ {
				g.SetValue(l.squares[i], BLANK)
			}
		}
	}
}

// ......X.10... with (2,2) or ......XX..110.... with (4,4)
// the goal is to look if a candidate fit in the gap up to a blank
// that candidate being the current clue or the next/previous one
// and if we don't find one, we can blank
func solveAlgo6(g *Griddler, l *Line) {
	//l.print("solveAlgo6")
	rsg := l.getUnsolvedRanges()

	for _, r := range rsg {
		cs := l.getPotentialCluesForRange(r)

		longest := 0
		isFound, step := l.getStepToNextBlank(r, false)
		if isFound {
			isFound = false
			for _, c := range cs {
				if c.index < l.ce {
					next := l.clues[c.index+1]
					if step <= c.length-r.length() || step > next.length {
						isFound = true
						break
					}
				}
				longest = max(longest, c.length-r.length())
			}
			//if we didn't find anyone, we can blank taking into account the longest trail
			if len(cs) > 1 && !isFound {
				for i := longest + 1; i <= step; i++ {
					g.SetValue(l.squares[r.max+i], BLANK)
				}
			}
		}

		longest = 0
		isFound, step = l.getStepToNextBlank(r, true)
		if isFound {
			isFound = false
			for _, c := range cs {
				if c.index > l.cb {
					previous := l.clues[c.index-1]
					if step <= c.length-r.length() || step > previous.length {
						isFound = true
						break
					}
				}
				longest = max(longest, c.length-r.length())
			}
			//if we didn't find anyone, we can blank taking into account the longest trail
			if len(cs) > 1 && !isFound {
				for i := longest + 1; i <= step; i++ {
					g.SetValue(l.squares[r.min-i], BLANK)
				}
			}
		}
	}
}

// ......YX0..YX0.. with (2,2,...) -> we can fill because of minimum size
// .......YXX.0.... with (4,5,...)
// for each filled group, we check the minimum size of all potential clue and fill
// i.e. if we find one that is currently the size or smaller that range size plus the gap, we can't do anything
// if not we take the shortest we found to do the fill
func solveAlgo7(g *Griddler, l *Line) {
	//l.print("solveAlgo7")
	rsg := l.getUnsolvedRanges()

	for _, r := range rsg {
		cs := l.getPotentialCluesForRange(r)

		shortest := l.length
		isFound, step := l.getStepToNextBlank(r, false)
		if isFound {
			isFound = false
			for _, c := range cs {
				if c.length <= step+r.length() {
					//fmt.Printf("Canceling... Clue(n:%d,b:%d,e:%d,l:%d)\n", c.index+1, c.begin+1, c.end+1, c.length)
					isFound = true
					break
				}
				shortest = min(shortest, c.length-step-r.length())
			}
			//if we didn't find anyone, we can fill taking into account the shortest trail
			if !isFound {
				for i := 0; i < shortest; i++ {
					g.SetValue(l.squares[r.min-i-1], 2)
				}
			}
		}

		shortest = l.length
		isFound, step = l.getStepToNextBlank(r, true)
		if isFound {
			isFound = false
			for _, c := range cs {
				if c.length <= step+r.length() {
					//fmt.Printf("Canceling... Clue(n:%d,b:%d,e:%d,l:%d)\n", c.index+1, c.begin+1, c.end+1, c.length)
					isFound = true
					break
				}
				shortest = min(shortest, c.length-step-r.length())
			}
			//if we didn't find anyone, we can fill taking into account the shortest trail
			if !isFound {
				for i := 0; i < shortest; i++ {
					g.SetValue(l.squares[r.max+i+1], FILLED)
				}
			}
		}
	}
}

// Algo 8: check possible border constraints following the pattern:
// |..X... -> .0X... (1,Z>1,...) or ...XX... -> ..0XX... with (2,Z>2,...)
func solveAlgo8(g *Griddler, l *Line) {
	//l.print("solveAlgo8")
	// From the beginning
	cb := l.clues[l.cb]
	if l.checkRangeForValue(2, cb.begin+cb.length+1, cb.begin+2*cb.length) {
		// fmt.Println("\nA8 Checking border min size constraints:")
		// fmt.Println("\nA8 Found at the beginning!")
		g.SetValue(l.squares[cb.begin+cb.length], BLANK)
	}

	// From the end
	ce := l.clues[l.ce]
	if l.checkRangeForValue(2, ce.end-2*ce.length, ce.end-ce.length-1) {
		// fmt.Println("\nA8 Checking border min size constraints:")
		// fmt.Println("\nA8 Found at the end!")
		g.SetValue(l.squares[ce.end-ce.length], BLANK)
	}
}
