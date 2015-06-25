package griddler

import (
	"fmt"
)

// user-defined function to define solving algorithm
type Algorithm func(g *Griddler, l *Line)

// algo to be used to solve basic case (empty/full) and initialize clue range
func solveInitAlgo(g *Griddler, l *Line) {
	switch {
	// no clues are defined for the line, we can blank everything
	case l.totalClues == 0:
		for _, s := range l.squares {
			g.setValue(s, BLANK)
		}
	// the total of the clue is equal to the line length (i.e. one big clue is defined)
	case l.totalClues == l.length:
		for _, s := range l.squares {
			g.setValue(s, FILLED)
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
func solveFilledRange(g *Griddler, l *Line) {
	l.print("solveFilledRange")

	rs := l.getRanges()
	for _, r := range rs {
		r.print("solveFilledRange")
	}

	for _, r := range rs {
		c := l.searchClueForRange(r, rs)

		if c != nil {
			if c.index == l.cb {
				for i := c.begin; i < r.max-c.length+1; i++ {
					g.setValue(l.squares[i], BLANK)
				}
			}
			if c.begin < r.max-c.length+1 {
				l.updateCluesRanges(c, r.max-c.length+1-c.begin, false)
			}
			if c.index == l.ce {
				for i := c.end; i > r.min+c.length-1; i-- {
					g.setValue(l.squares[i], BLANK)
				}
			}
			if c.end > r.min+c.length-1 {
				l.updateCluesRanges(c, c.end-r.min-c.length+1, true)
			}
			c.solveConstraints(true)
			c.solveConstraints(false)
			c.solveOverlap()
			c.solveCompleteness()
		}
	}
}

// When a range has only one clue associated, update ranges accordingly
// and blank if first or last: ( ......XXX....)
func solveAlgo9(g *Griddler, l *Line) {
	rsg := l.getRanges()

	for _, r := range rsg {
		cs := l.getPotentialCluesForRange(r)

		if len(cs) == 1 {
			c := cs[0]
			if c.index == l.cb {
				for i := c.begin; i < r.max-c.length+1; i++ {
					g.setValue(l.squares[i], 1)
				}
			}
			if c.begin < r.max-c.length+1 {
				l.updateCluesRanges(c, r.max-c.length+1-c.begin, false)
			}
			if c.index == l.ce {
				for i := c.end; i > r.min+c.length-1; i-- {
					g.setValue(l.squares[i], 1)
				}
			}
			if c.end > r.min+c.length-1 {
				l.updateCluesRanges(c, c.end-r.min-c.length+1, true)
			}
		}
	}
}

// algo 3, check the clues with regards to maximal sizes on found squares
func solveAlgo3(g *Griddler, l *Line) {
	rsg := l.getRanges()

	for _, r := range rsg {

		cs := l.getPotentialCluesForRange(r)
		canceled := false

		for _, c := range cs {
			cs := make([](*Clue), 0)
			switch {
			case c.length < r.length():
			case c.length == r.length():
				//fmt.Printf("Clue(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
				cs = append(cs, c)
			case c.length > r.length():
				//fmt.Printf("Canceling...Clue(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
				canceled = true
			}
		}
		if !canceled && len(cs) > 0 {
			fmt.Println("\nA3 Checking group size:")
			fmt.Printf("Line clue range: cb:%d, ce:%d\n", l.cb, l.ce)
			fmt.Printf("Range(b:%d,e:%d):\n", r.min+1, r.max+1)
			g.setValue(l.squares[r.min-1], 1)
			g.setValue(l.squares[r.max+1], 1)

			// if we found only one candidate, if it is at the beginning or end, we can blank more
			if len(cs) == 1 {
				c := cs[0]
				if c.index == l.cb {
					for i := 0; i < r.min-1; i++ {
						g.setValue(c.l.squares[i], 1)
					}
					c.l.incrementCluesBegin(c.index, r.min-c.begin)
					//fmt.Printf("\nNewClue(A3b)(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
				}

				if c.index == l.ce {
					for i := l.length - 1; i > r.max+1; i-- {
						g.setValue(c.l.squares[i], 1)
					}
					c.l.decrementCluesEnd(c.index, c.end-r.max)
					//fmt.Printf("\nNewClue(A3e)(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
				}
			}
		}
	}
}

// algo 4, check the clue with regards to existing found squares and ownership
func solveAlgo4(g *Griddler, l *Line) {
	rsg := l.getRanges()

	if len(rsg) > l.ce-l.cb {

		if l.check1to1Mapping(rsg) {
			fmt.Println("\nA4 Checking filled group 1-1 mapping:")
			for i, c := range l.clues[l.cb : l.ce+1] {
				// fmt.Printf("\nClue(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
				// fmt.Printf("\nRange(b:%d,e:%d):", rsg[ri].min+1, rsg[ri].max+1)
				checkClueAlgo4(g, c, i, rsg)
			}
		}
	}
}

func checkClueAlgo4(g *Griddler, c *Clue, ri int, rsg [](*Range)) {
	// if first, blank up to potential beginning of clue
	if c.index == c.l.cb {
		for i := 0; i < rsg[ri].max-c.length+1; i++ {
			g.setValue(c.l.squares[i], 1)
		}
		c.l.incrementCluesBegin(c.index, rsg[ri].max-c.length+1-c.begin)
		//fmt.Printf("\nNewClue(A4b)(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
	}
	// take into account left neighbour
	if c.index > c.l.cb {
		for i := rsg[ri].max - c.length; i > rsg[ri-1].max+c.l.clues[c.index-1].length; i-- {
			g.setValue(c.l.squares[i], 1)
		}
	}
	// take into account right neighbour
	if c.index < c.l.ce {
		for i := rsg[ri].min + c.length; i < rsg[ri+1].min-c.l.clues[c.index+1].length; i++ {
			g.setValue(c.l.squares[i], 1)
		}
	}

	// if last, blank down to potential ending of clue
	if c.index == c.l.ce {
		for i := c.l.length - 1; i > rsg[ri].min+c.length-1; i-- {
			g.setValue(c.l.squares[i], 1)
		}
		c.l.decrementCluesEnd(c.index, c.end-rsg[ri].min-c.length+1)
		//fmt.Printf("\nNewClue(A4e)(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
	}
}

// ......X.10... with (2,2) or ......XX..110.... with (4,4)
// the goal is to look if a candidate fit in the gap up to a blank
// that candidate being the current clue or the next/previous one
// and if we don't find one, we can blank
func solveAlgo6(g *Griddler, l *Line) {
	rsg := l.unsolvedRanges()

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
				fmt.Println("\nA6 Checking filled/blank constraints:")
				fmt.Printf("Range(b:%d,e:%d):\n", r.min+1, r.max+1)
				fmt.Printf("A6 Forward: step:%d, longest:%d\n", step, longest)
				for i := longest + 1; i <= step; i++ {
					g.setValue(l.squares[r.max+i], 1)
					//Pause()
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
				fmt.Println("\nA6 Checking filled/blank constraints:")
				fmt.Printf("Range(b:%d,e:%d):\n", r.min+1, r.max+1)
				fmt.Printf("A6 Backward: step:%d, longest:%d\n", step, longest)
				for i := longest + 1; i <= step; i++ {
					g.setValue(l.squares[r.min-i], 1)
					//Pause()
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
	rsg := l.unsolvedRanges()

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
				fmt.Println("\nA7 Checking filled/size constraints:")
				fmt.Printf("Range(b:%d,e:%d):\n", r.min+1, r.max+1)
				fmt.Printf("A7 Backward fill: step:%d, shortest:%d\n", step, shortest)
				for i := 0; i < shortest; i++ {
					g.setValue(l.squares[r.min-i-1], 2)
					//Pause()
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
				fmt.Println("\nA7 Checking filled/size constraints:")
				fmt.Printf("Range(b:%d,e:%d):\n", r.min+1, r.max+1)
				fmt.Printf("A7 Forward: step:%d, shortest:%d\n", step, shortest)
				for i := 0; i < shortest; i++ {
					g.setValue(l.squares[r.max+i+1], 2)
					//Pause()
				}
			}
		}
	}
}

// Algo 8: check possible border constraints following the pattern:
// |..X... -> .0X... (1,Z>1,...) or ...XX... -> ..0XX... with (2,Z>2,...)
func solveAlgo8(g *Griddler, l *Line) {
	// From the beginning
	cb := l.clues[l.cb]
	if l.checkRangeForValue(2, cb.begin+cb.length+1, cb.begin+2*cb.length) {
		fmt.Println("\nA8 Checking border min size constraints:")
		fmt.Println("\nA8 Found at the beginning!")
		g.setValue(l.squares[cb.begin+cb.length], 1)
	}

	// From the end
	ce := l.clues[l.ce]
	if l.checkRangeForValue(2, ce.end-2*ce.length, ce.end-ce.length-1) {
		fmt.Println("\nA8 Checking border min size constraints:")
		fmt.Println("\nA8 Found at the end!")
		g.setValue(l.squares[ce.end-ce.length], 1)
	}
}
