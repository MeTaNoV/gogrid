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
func solveFilledRanges(g *Griddler, l *Line) {
	l.print("solveFilledRanges")

	rs := l.getRanges()
	l.updateCluesForRanges(rs)

	for _, r := range rs {
		r.print("solveFilledRanges")

		cs := l.getPotentialCluesForRange(r)

		switch {
		case len(cs) == 0:
			// TODO throw error
		case len(cs) == 1:
			c := cs[0]
			c.print("solveFilledRanges")
			if c.index == l.cb {
				// TODO, here we start at 0, because c.begin is already updated, else
				// we can move the blank process in updateCluesForRanges()
				for i := 0; i < r.max-c.length+1; i++ {
					g.setValue(l.squares[i], BLANK)
				}
			}
			if c.begin < r.max-c.length+1 {
				l.incrementCluesBegin(c, r.max-c.length+1-c.begin)
			}
			if c.index == l.ce {
				// TODO, here we start at l.length-1, because c.end is already updated, else
				// we can move the blank process in updateCluesForRanges()
				for i := l.length - 1; i > r.min+c.length-1; i-- {
					g.setValue(l.squares[i], BLANK)
				}
			}
			if c.end > r.min+c.length-1 {
				l.decrementCluesEnd(c, c.end-r.min-c.length+1)
			}
			c.solveConstraints(true)
			c.solveConstraints(false)
			c.solveOverlap()
			c.solveCompleteness()
		case len(cs) > 1:
			for _, c := range cs {
				c.print("solveFilledRanges")
			}
			// if all potential clues are of the Range size, we can finish it
			if maxLength(cs) == r.length() {
				g.setValue(l.squares[r.min-1], BLANK)
				g.setValue(l.squares[r.max+1], BLANK)
			}
			// we can increment or decrement ??? It should be done in updateCluesForRanges() already...
			// TODO how can we blank beginning or end ?!?
		}
	}
}

// algo 5, check finished ranged and try to find the corresponding clue from beginning or end to fill some blank
// e.g. ....0X0...0X0....X.X... for a (1,1,4) clue list will enable to blank the first 4 square
func solveFinishedRanges(g *Griddler, l *Line) {
	rs := l.solvedRanges()
	if len(rs) == 0 {
		return
	}

	for _, r := range rs {
		r.print("solveFinishedRanges")
	}

	//Pause()

	// we want to map the solved group to clues and see if the first or last clue are among those for all possible mapping
	// if this is the case, we can blank and resolve those!
	// to verify that, we only have to check that the first/last clue is in the first/last range available
	// and that no other mapping is possible, i.e. if we find another mapping, it fails
	cbeg := l.clues[l.cb]
	// if the first clue contains the first range, we can proceed further
	if cbeg.begin <= rs[0].min && cbeg.end >= rs[0].max && cbeg.length == rs[0].length() {
		cbeg.print("cbeg")
		isFound := false
		lastIndex := cbeg.index
		// we check if we found a possible mapping without the first clue
		for _, r := range rs {
			isFound = false
			r.print("r")
			for _, c := range l.clues[lastIndex+1 : l.ce+1] {
				if c.begin <= r.min && c.end >= r.max && c.length == r.length() {
					c.print("c")
					isFound = true
					lastIndex = c.index
					break
				}
			}
		}
		// if not found, the first range is the first clue!
		if !isFound {
			fmt.Println("A5 Checking solved group mapping:")
			fmt.Println("A5 From beginning:")

			for i := 0; i < rs[0].min-1; i++ {
				g.setValue(l.squares[i], 1)
			}
			cbeg.isDone = true
			l.incrementCluesBegin(cbeg, rs[0].min-cbeg.begin)
			l.updateCluesIndexes(cbeg)
		}
	}

	//Pause()

	cend := l.clues[l.ce]
	// if the last clue contains the last range, we can proceed further
	if cend.begin <= rs[len(rs)-1].min && cend.end >= rs[len(rs)-1].max && cend.length == rs[len(rs)-1].length() {
		isFound := false
		lastIndex := cend.index
		// we check if we found a possible mapping without the first clue
		for i := len(rs); i > 0; i-- {
			r := rs[i-1]
			isFound = false
			for j := lastIndex - 1; j >= l.cb; j-- {
				c := l.clues[j]
				if c.begin <= r.min && c.end >= r.max && c.length == r.length() {
					isFound = true
					lastIndex = c.index
					break
				}
			}
		}
		// if not found, the first range is the first clue!
		if !isFound {
			fmt.Println("A5 Checking solved group mapping:")
			fmt.Println("A5 From end:")
			for i := l.length - 1; i > rs[len(rs)-1].max+1; i-- {
				g.setValue(l.squares[i], 1)
			}
			cend.isDone = true
			l.decrementCluesEnd(cend, cend.end-rs[len(rs)-1].max)
			l.updateCluesIndexes(cend)
		}
	}

	//Pause()
}

func solveEmptyRanges(g *Griddler, l *Line) {

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
