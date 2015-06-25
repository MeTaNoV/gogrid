package griddler

import (
	"fmt"
)

type Algorithm func(g *Griddler, l *Line)

func solveInitAlgo(g *Griddler, l *Line) {
	switch {
	// no clues are defined for the line, we can blank everything
	case l.totalClues == 0:
		for _, s := range l.squares {
			g.setValue(s, 1)
		}
	// the total of the clue is equal to the line length (i.e. one big clue is defined)
	case l.totalClues == l.length:
		for _, s := range l.squares {
			g.setValue(s, 2)
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

// algo 1, fill the beginning and end of line if possible
// XYY1...... with (3,...)
// .YX.0.... with (3,...)
func solveAlgo1(g *Griddler, l *Line) {
	// From the beginning
	for i := l.cb; i <= l.ce; i++ {
		c := l.clues[i]
		//fmt.Println("\nA1 From beginning:")
		//fmt.Printf("\nClue(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
		if !checkClueAlgo1(g, c, false) {
			return
		}
		if c.isDone {
			continue
		}
		break
	}

	// From the end
	for i := l.ce; i >= l.cb; i-- {
		c := l.clues[i]
		//fmt.Println("\nA1 From end:")
		//fmt.Printf("\nClue(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
		if !checkClueAlgo1(g, c, true) {
			return
		}
		if c.isDone {
			continue
		}
		break
	}
}

// TODO refactor if possible...
func checkClueAlgo1(g *Griddler, c *Clue, reverse bool) bool {
	emptyBefore, emptyAfter := 0, 0
	filledBefore, filledAfter := 0, 0
	l := c.l
	i := c.begin
	if reverse {
		i = c.end
	}
	//fmt.Printf("\nAlgo 1:")
	for {
		switch {
		case l.squares[i].val == 0:
			//fmt.Printf("(%d,%d, )", l.squares[i].x+1, l.squares[i].y+1)
			if filledBefore > 0 {
				// we can potentially fill the square
				if filledBefore+emptyBefore < c.length {
					g.setValue(l.squares[i], 2)
					filledBefore++
					break
				}
				// if we filled it completely, we can blank and potentially before or after
				if filledBefore == c.length {
					g.setValue(l.squares[i], 1)
					if reverse {
						if i < l.length-c.length-1 {
							g.setValue(l.squares[i+c.length+1], 1)
						}
					} else {
						if i > c.length+1 {
							g.setValue(l.squares[i-c.length-1], 1)
						}
					}
					// flag the clue
					c.isDone = true
					// update line clue indexes
					l.updateCluesIndexes(c, reverse)
					// no need to update the range of clues

					return true
				}
				// we increment the blank after and exit if necessary
				emptyAfter++
				if filledBefore+emptyAfter+filledAfter > c.length {
					return true
				}
			} else {
				// the number of empty square is > at the searched length, we can exit
				emptyBefore++
				if emptyBefore > c.length {
					return true
				}
			}
		case l.squares[i].val == 1:
			//fmt.Printf("(%d,%d,0)", l.squares[i].x+1, l.squares[i].y+1)
			if filledBefore > 0 {
				// we can pursue the filling going backward (ex: ..X..0)
				for j := filledBefore + emptyAfter + filledAfter; j < c.length; j++ {
					if reverse {
						g.setValue(l.squares[i+1+j], 2)
					} else {
						g.setValue(l.squares[i-1-j], 2)
					}
				}
				// update range
				if reverse {
					c.begin = i + 1
				} else {
					c.end = i - 1
				}

				if emptyAfter == 0 {
					// ending with blanks if we don't reach the border
					if reverse {
						if i+1+c.length < l.length {
							g.setValue(l.squares[i+1+c.length], 1)
						}
					} else {
						if i-1-c.length > 0 {
							g.setValue(l.squares[i-1-c.length], 1)
						}
					}
					// flag the clue
					c.isDone = true
					// update line clue status
					l.updateCluesIndexes(c, reverse)
					// TODO: be sure no need to update range
				}

				return true
			}
			if emptyBefore > 0 {
				// if no place for this clue, we can blank
				if emptyBefore < c.length {
					for j := 0; j < emptyBefore; j++ {
						if reverse {
							g.setValue(l.squares[i+1+j], 1)
						} else {
							g.setValue(l.squares[i-1-j], 1)
						}
					}
				} else {
					return true
				}
			}
			// if we encounter this first or only empties, we can propagate the update of clue's range
			l.updateCluesRanges(c, emptyBefore+1, reverse)
			emptyBefore = 0
		case l.squares[i].val == 2:
			//fmt.Printf("(%d,%d,X)", l.squares[i].x+1, l.squares[i].y+1)
			if emptyAfter == 0 {
				// if this is the first fill, we can update the clue range
				if filledBefore == 0 {
					if reverse {
						c.begin = i - c.length + 1
					} else {
						c.end = i + c.length - 1
					}
				}
				filledBefore++
				switch {
				case filledBefore < c.length:
					// if we joined existing checked square (e.g ..XX.X)
					if filledBefore+emptyBefore > c.length {
						if reverse {
							g.setValue(l.squares[i+c.length], 1)
						} else {
							g.setValue(l.squares[i-c.length], 1)
						}
						l.updateCluesRanges(c, 1, reverse)
						emptyBefore--
					}
				// here we found it all, we can blank the beginning and the end
				case filledBefore == c.length:
					if reverse {
						if i > 0 {
							g.setValue(l.squares[i-1], 1)
						}
						if i < l.length-c.length {
							g.setValue(l.squares[i+c.length], 1)
						}
					} else {
						if i < l.length-1 {
							g.setValue(l.squares[i+1], 1)
						}
						if i >= c.length {
							g.setValue(l.squares[i-c.length], 1)
						}
					}
					// if it was the last clue, blank until the end
					if reverse {
						if c.index == l.cb {
							for j := 0; j < i; j++ {
								g.setValue(l.squares[j], 1)
							}
						}
					} else {
						if c.index == l.ce {
							for j := c.end; j > i; j-- {
								g.setValue(l.squares[j], 1)
							}
						}
					}

					// flag the clue
					c.isDone = true
					// update line clue status
					l.updateCluesIndexes(c, reverse)
					// updqte clues ranges
					l.updateCluesRanges(c, emptyBefore, reverse)
					return true
				case filledBefore > c.length:
					fmt.Println("WWAARRNNIINNGG: filled > c.length")
					return false
				}
			} else {
				filledAfter++
			}
		}
		if reverse {
			i--
			if i < c.begin-1 || i < 0 {
				return true
			}
		} else {
			i++
			if i > c.end+1 || i > l.length-1 {
				return true
			}
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

// algo 2, update the valid range for each clue and update possible values with overlap
// 0110..Y..010 with (3,...)
func solveAlgo2(g *Griddler, l *Line) {
	for _, c := range l.clues[l.cb : l.ce+1] {
		if c.isDone {
			continue
		}
		// fmt.Println("\nA2 From beginning:")
		// fmt.Printf("Clue(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
		if !checkClueAlgo2(g, c, false) {
			return
		}
		// fmt.Println("\nA2 From end:")
		// fmt.Printf("Clue(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
		if !checkClueAlgo2(g, c, true) {
			return
		}

		// fmt.Println("\nSolve Overlap:")
		c.solveOverlap()

		// if first or last clue, we can run algo 1 to check if done
		if c.index == l.cb {
			checkClueAlgo1(g, c, false)
		}
		if c.index == l.ce {
			checkClueAlgo1(g, c, true)
		}
	}
}

func checkClueAlgo2(g *Griddler, c *Clue, reverse bool) bool {
	empty := 0
	filled := 0
	l := c.l
	i := c.begin
	if reverse {
		i = c.end
	}
	for {
		switch {
		case l.squares[i].val == 0:
			//fmt.Printf("(%d,%d) ", l.squares[i].x+1, l.squares[i].y+1)
			empty++
		case l.squares[i].val == 1:
			//fmt.Printf("(%d,%d).", l.squares[i].x+1, l.squares[i].y+1)
			if (empty + filled) < c.length {
				l.updateCluesRanges(c, empty+filled+1, reverse)
				empty = 0
				filled = 0
			}
		case l.squares[i].val == 2:
			//fmt.Printf("(%d,%d)X", l.squares[i].x+1, l.squares[i].y+1)
			filled++
		}
		if reverse {
			i--
			if i < c.begin {
				return true
			}
		} else {
			i++
			if i > c.end {
				return true
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

// algo 5, check finished ranged and try to find the corresponding clue from beginning or end to fill some blank
// e.g. ....0X0...0X0....X.X... for a (1,1,4) clue list will enable to blank the first 4 square
func solveAlgo5(g *Griddler, l *Line) {
	rsg := l.solvedRanges()
	// for _, r := range rsg {
	// 	fmt.Printf("\nRange(b:%d,e:%d):", r.min, r.max)
	// }
	if len(rsg) == 0 {
		return
	}

	// we want to map the solved group to clues and see if the first or last clue are among those for all possible mapping
	// if this is the case, we can blank and resolve those!
	// to verify that, we only have to check that the first/last clue is in the first/last range available
	// and that no other mapping is possible, i.e. if we find another mapping, it fails
	cbeg := l.clues[l.cb]
	// if the first clue contains the first range, we can proceed further
	if cbeg.begin <= rsg[0].min && cbeg.end >= rsg[0].max && cbeg.length == rsg[0].length() {
		isFound := false
		lastIndex := cbeg.index
		// we check if we found a possible mapping without the first clue
		for _, r := range rsg {
			isFound = false
			for _, c := range l.clues[lastIndex+1 : l.ce+1] {
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
			fmt.Println("A5 From beginning:")

			for i := 0; i < rsg[0].min-1; i++ {
				g.setValue(l.squares[i], 1)
			}
			cbeg.isDone = true
			l.updateCluesRanges(cbeg, rsg[0].min-cbeg.begin, false)
			l.updateCluesIndexes(cbeg, false)
		}
	}

	cend := l.clues[l.ce]
	// if the last clue contains the last range, we can proceed further
	if cend.begin <= rsg[len(rsg)-1].min && cend.end >= rsg[len(rsg)-1].max && cend.length == rsg[len(rsg)-1].length() {
		isFound := false
		lastIndex := cend.index
		// we check if we found a possible mapping without the first clue
		for i := len(rsg); i > 0; i-- {
			r := rsg[i-1]
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
			for i := l.length - 1; i > rsg[len(rsg)-1].max+1; i-- {
				g.setValue(l.squares[i], 1)
			}
			cend.isDone = true
			l.updateCluesRanges(cend, cend.end-rsg[len(rsg)-1].max, true)
			l.updateCluesIndexes(cend, true)
		}
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
