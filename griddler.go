package griddler

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Griddler struct {
	width          int
	height         int
	lines          [](*Line)
	columns        [](*Line)
	sumSolvedLines int
	isDone         bool
	lStack         Stack
	cStack         Stack
	solveQueue     chan (*Square)
}

func NewGriddler() *Griddler {
	g := &Griddler{
		sumSolvedLines: 0,
		isDone:         false,
		lStack:         Stack{},
		cStack:         Stack{},
	}
	return g
}

func (g *Griddler) error(e error, l int) error {
	return &ParseError{
		line: l,
		err:  e,
	}
}

func (g *Griddler) initBoard() {
	g.lines = make([](*Line), g.height)
	for i := 0; i < g.height; i++ {
		g.lines[i] = NewLine(g, i, g.width)
		for j := 0; j < g.width; j++ {
			g.lines[i].squares[j] = NewSquare(i, j, 0, g)
		}
	}
	g.columns = make([](*Line), g.width)
	for i := 0; i < g.width; i++ {
		g.columns[i] = NewLine(g, i, g.height)
		for j := 0; j < g.height; j++ {
			g.columns[i].squares[j] = g.lines[j].squares[i]
		}
	}
	g.solveQueue = make(chan (*Square), g.width*g.height)
}

func (g *Griddler) incrementSolvedLines() {
	g.sumSolvedLines++
	if g.sumSolvedLines == g.width+g.height {
		g.isDone = true
	}
}

func (g *Griddler) Load(filename string) error {
	gFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer gFile.Close()

	gScanner := bufio.NewScanner(gFile)

	// Reading the griddler size on the first line
	gScanner.Scan()
	firstLine := gScanner.Text()
	firstLineSize := strings.Split(firstLine, "x")
	if len(firstLineSize) != 2 {
		return g.error(ErrInvalidGridSizeFormat, 1)
	}
	width, err := strconv.Atoi(firstLineSize[0])
	if err != nil {
		return g.error(ErrInvalidGridSizeValue, 1)
	}
	g.width = width
	height, err := strconv.Atoi(firstLineSize[1])
	if err != nil {
		return g.error(ErrInvalidGridSizeValue, 1)
	}
	g.height = height
	g.initBoard()

	// Reading the clue line until the end of the file
	line := 1
	for gScanner.Scan() {
		line++
		gLine := gScanner.Text()
		gLineTokens := strings.Split(gLine, ";")
		if len(gLineTokens) != 2 {
			return g.error(ErrMissingSemiColon, line)
		}

		gLineInfos := strings.Split(gLineTokens[0], ":")
		index, err := strconv.Atoi(gLineInfos[1])
		if err != nil {
			return g.error(ErrInvalidIntLine, line)
		}

		gLineStrings := strings.Split(gLineTokens[1], ",")
		gLineNumbers := make([](*Clue), len(gLineStrings))
		for i, val := range gLineStrings {
			conv, err := strconv.Atoi(val)
			gLineNumbers[i] = NewClue(conv)
			if err != nil {
				return g.error(ErrInvalidIntValue, line)
			}
		}

		switch gLineInfos[0] {
		case "H":
			if index > height {

			}
			g.lines[index-1].addClues(gLineNumbers)
		case "V":
			if index > width {

			}
			g.columns[index-1].addClues(gLineNumbers)
		default:
			return g.error(ErrInvalidTokenLine, line)
		}
	}
	return nil
}

func (g *Griddler) setValue(s *Square, val int) {
	if s.val == 0 {
		s.val = val
		g.lStack.push(g.lines[s.x])
		g.cStack.push(g.columns[s.y])
		if val == 2 {
			g.lines[s.x].incrementClues()
			g.columns[s.y].incrementClues()
		} else {
			g.lines[s.x].incrementBlanks()
			g.columns[s.y].incrementBlanks()
		}
		fmt.Printf("FOUND (%d,%d)\n", s.x+1, s.y+1)
		g.solveQueue <- s
		g.Show()
		//Pause()
	}
}

func (g *Griddler) Show() {
	for i := 0; i < g.height+2; i++ {
		if i == 0 || i == g.height+1 {
			fmt.Printf("+")
			for j := 0; j < g.width; j++ {
				fmt.Printf("-")
			}
			fmt.Println("+")
		} else {
			fmt.Printf("|")
			for j := 0; j < g.width; j++ {
				g.lines[i-1].squares[j].show()
			}
			fmt.Printf("| %d\n", i)
			//fmt.Printf("(%d,%d,%t)\n", g.lines[i-1].sum, g.lines[i-1].sumClues, g.lines[i-1].isDone)
		}
	}
	fmt.Printf(" ")
	for i := 0; i < g.width; i++ {
		fmt.Printf("%d", (i+1)%10)
		//fmt.Printf("(%d,%d,%t)\n", g.columns[i].sum, g.columns[i].sumClues, g.columns[i].isDone)
	}
	//fmt.Printf("\nLines completed: %d", g.sum)
	fmt.Println()
}

func (g *Griddler) Solve() {
	g.solveInit()
	l := g.lStack.pop()
	c := g.cStack.pop()
	for l != nil || c != nil {
		if l != nil && !l.isDone {
			fmt.Printf("\n=================== checking line %d ===================\n", l.index)
			Pause()
			g.checkLine(l)
		}
		if c != nil && !c.isDone {
			fmt.Printf("\n=================== checking column %d ===================\n", c.index)
			Pause()
			g.checkLine(c)
		}
		l = g.lStack.pop()
		c = g.cStack.pop()
	}
	if g.isDone {
		fmt.Println("Griddler completed!!!")
	} else {
		fmt.Println("Griddler uncompleted, find new search algorithm!")
	}
}

func (g *Griddler) solveInit() {
	for _, line := range g.lines {
		g.solveInitLine(line)
	}
	for _, col := range g.columns {
		g.solveInitLine(col)
	}
}

func (g *Griddler) solveInitLine(line *Line) {
	switch {
	case line.totalClues == 0:
		for _, s := range line.squares {
			g.setValue(s, 1)
		}
	case line.totalClues == g.width:
		for _, s := range line.squares {
			g.setValue(s, 2)
		}
	default:
		for i, clue := range line.clues {
			sumBegin := 0
			for _, c := range line.clues[0:i] {
				sumBegin += c.length + 1
			}
			clue.begin = sumBegin
			sumEnd := 0
			for _, c := range line.clues[i+1:] {
				sumEnd += c.length + 1
			}
			clue.end = line.length - 1 - sumEnd
			diff := clue.begin + clue.length - (clue.end + 1 - clue.length)
			if diff > 0 {
				for j := 0; j < diff; j++ {
					g.setValue(line.squares[clue.end-clue.length+1+j], 2)
				}
			}
		}
	}
}

func (g *Griddler) checkLine(l *Line) bool {
	// if we found all clues, we can blank all remaining square
	if l.sumClues == l.totalClues {
		for _, s := range l.squares {
			g.setValue(s, 1)
		}
		return true
	}
	// if we found all blanks, we can set the remaining clues
	if l.sumBlanks == l.length-l.totalClues {
		for _, s := range l.squares {
			g.setValue(s, 2)
		}
		return true
	}
	// algo 1, fill the beginning and end of line if possible
	res := g.checkLineAlgo1(l)
	if !res {
		return false
	}

	// algo 2, update the valid range for each clue and update possible values
	if !l.isDone {
		res = g.checkLineAlgo2(l)
		if !res {
			return false
		}
	}

	// algo 3, check the clues with regards to maximal sizes on found squares
	if !l.isDone {
		res = g.checkLineAlgo3(l)
		if !res {
			return false
		}
	}

	// algo 4, check the clue with regards to existing found squares and ownership
	if !l.isDone {
		res = g.checkLineAlgo4(l)
		if !res {
			return false
		}
	}
	// algo 5, check finished ranged and try to find the corresponding clue from beginning or end to fill some blank
	// e.g. ....0X0...0X0....X.X... for a (1,1,4) clue list will enable to blank the first 4 square
	if !l.isDone {
		res = g.checkLineAlgo5(l)
		if !res {
			return false
		}
	}

	return true
}

func (g *Griddler) checkLineAlgo1(l *Line) bool {
	// From the beginning
	for i := l.cb; i <= l.ce; i++ {
		c := l.clues[i]
		if c.isDone {
			continue
		}
		fmt.Println("\nA1 From beginning:")
		fmt.Printf("\nClue(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
		if !g.checkClueAlgo1(c, false) {
			return false
		}
		if c.isDone {
			continue
		}
		break
	}

	// From the end
	for i := l.ce; i >= l.cb; i-- {
		c := l.clues[i]
		if c.isDone {
			continue
		}
		fmt.Println("\nA1 From end:")
		fmt.Printf("\nClue(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
		if !g.checkClueAlgo1(c, true) {
			return false
		}
		if c.isDone {
			continue
		}
		break
	}
	return true
}

func (g *Griddler) checkClueAlgo1(c *Clue, reverse bool) bool {
	emptyBefore, emptyAfter := 0, 0
	filledBefore, filledAfter := 0, 0
	l := c.l
	i := c.begin
	if reverse {
		i = c.end
	}
	fmt.Printf("\nAlgo 1:")
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

					l.updateCluesRanges(c, emptyBefore, reverse)
					// flag the clue
					c.isDone = true
					// update line clue status
					l.updateCluesIndexes(c, reverse)
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

func (g *Griddler) checkLineAlgo2(l *Line) bool {
	for _, c := range l.clues[l.cb:l.ce] {
		if c.isDone {
			continue
		}
		fmt.Println("\nA2 From beginning:")
		fmt.Printf("\nClue(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
		if !g.checkClueAlgo2(c, false) {
			return false
		}
		fmt.Println("\nA2 From end:")
		fmt.Printf("\nClue(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
		if !g.checkClueAlgo2(c, true) {
			return false
		}
		c.solveOverlap()
	}
	return true
}

func (g *Griddler) checkClueAlgo2(c *Clue, reverse bool) bool {
	empty := 0
	filled := 0
	l := c.l
	i := c.begin
	if reverse {
		i = c.end
	}
	fmt.Printf("\nAlgo 2:")
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

func (g *Griddler) checkLineAlgo3(l *Line) bool {
	rsg := l.filledGroups()
	// for _, r := range rsg {
	// 	fmt.Printf("\nRange(b:%d,e:%d):", r.min, r.max)
	// }

	if len(rsg) == 0 {
		return true
	}

	fmt.Println("\nA3 Checking group size:")
	fmt.Printf("Line clue range: cb:%d, ce:%d", l.cb, l.ce)

	for _, r := range rsg {
		fmt.Printf("\nRange(b:%d,e:%d):", r.min, r.max)
		cs := make([](*Clue), 0)
		for _, c := range l.clues[l.cb:l.ce] {
			if c.begin <= r.min && c.end >= r.max {
				switch {
				case c.length < r.length():
				case c.length == r.length():
					fmt.Printf("\nClue(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
					cs = append(cs, c)
				case c.length > r.length():
					return true
				}
			}
		}
		if len(cs) > 0 {
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
					fmt.Printf("\nNewClue(A3b)(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
				}

				if c.index == l.ce {
					for i := l.length - 1; i > r.max+1; i-- {
						g.setValue(c.l.squares[i], 1)
					}
					c.l.decrementCluesEnd(c.index, c.end-r.max)
					fmt.Printf("\nNewClue(A3e)(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
				}
			}
		}
	}

	return true
}

func (g *Griddler) checkLineAlgo4(l *Line) bool {
	rsg := l.filledGroups()

	if len(rsg) > l.ce-l.cb {
		fmt.Println("\nA4 Checking filled group mapping:")

		if l.checkMapping(rsg) {
			for i, c := range l.clues[l.cb:l.ce] {
				fmt.Printf("\nClue(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
				fmt.Printf("\nRange(b:%d,e:%d):", rsg[i].min, rsg[i].max)
				g.checkClueAlgo4(c, rsg)
			}
		}
	}

	return true
}

func (g *Griddler) checkClueAlgo4(c *Clue, r [](*Range)) bool {
	fmt.Printf("\nAlgo 4:")

	// if first, blank up to potential beginning of clue
	if c.index == c.l.cb {
		for i := c.begin; i < r[c.index-c.l.cb].max-c.length+1; i++ {
			g.setValue(c.l.squares[i], 1)
		}
		c.l.incrementCluesBegin(c.index, r[c.index-c.l.cb].max-c.length+1-c.begin)
		fmt.Printf("\nNewClue(A4b)(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
	}
	// take into account left neighbour
	if c.index > c.l.cb {
		for i := r[c.index-c.l.cb].max - c.length; i > r[c.index-c.l.cb-1].max+c.l.clues[c.index-1].length; i-- {
			g.setValue(c.l.squares[i], 1)
		}
	}
	// take into account right neighbour
	if c.index < c.l.ce {
		for i := r[c.index-c.l.cb].min + c.length; i < r[c.index-c.l.cb+1].min-c.l.clues[c.index+1].length; i++ {
			g.setValue(c.l.squares[i], 1)
		}
	}

	// if last, blank down to potential ending of clue
	if c.index == c.l.ce {
		for i := c.end; i > r[c.index-c.l.cb].min+c.length-1; i-- {
			g.setValue(c.l.squares[i], 1)
		}
		c.l.decrementCluesEnd(c.index, c.end-r[c.index-c.l.cb].min-c.length+1)
		fmt.Printf("\nNewClue(A4e)(n:%d,b:%d,e:%d,l:%d):", c.index+1, c.begin+1, c.end+1, c.length)
	}

	return true
}

func (g *Griddler) checkLineAlgo5(l *Line) bool {
	rsg := l.solvedGroups()
	// for _, r := range rsg {
	// 	fmt.Printf("\nRange(b:%d,e:%d):", r.min, r.max)
	// }
	if len(rsg) == 0 {
		return true
	}

	fmt.Println("\nA5 Checking solved group mapping:")

	// we want to map the solved group to clues and see if the first or last clue are among those for all possible mapping
	// if this is the case, we can blank and resolve those!
	// to verify that, we only have to check that the first/last clue is in the first/last range available
	// and that no other mapping is possible, i.e. if we find another mapping, it fails
	fmt.Println("\nA5 From beginning:")
	cbeg := l.clues[l.cb]
	// if the first clue contains the first range, we can proceed further
	if cbeg.begin <= rsg[0].min && cbeg.end >= rsg[0].max && cbeg.length == rsg[0].length() {
		isFound := false
		lastIndex := cbeg.index
		// we check if we found a possible mapping without the first clue
		for _, r := range rsg {
			isFound = false
			for _, c := range l.clues[lastIndex+1 : l.ce] {
				if c.begin <= r.min && c.end >= r.max && c.length == r.length() {
					isFound = true
					lastIndex = c.index
					break
				}
			}
		}
		// if not found, the first range is the first clue!
		if !isFound {
			for i := 0; i < rsg[0].min-1; i++ {
				g.setValue(l.squares[i], 1)
			}
			cbeg.isDone = true
			l.updateCluesRanges(cbeg, rsg[0].min-cbeg.begin, false)
			l.updateCluesIndexes(cbeg, false)
		}
	}

	fmt.Println("\nA5 From end:")
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
			for i := l.length - 1; i > rsg[len(rsg)-1].max+1; i++ {
				g.setValue(l.squares[i], 1)
			}
			cend.isDone = true
			l.updateCluesRanges(cend, cend.end-rsg[len(rsg)-1].max, true)
			l.updateCluesIndexes(cend, true)
		}
	}

	return true
}

func (g *Griddler) checkClueAlgo5(c *Clue, r [](*Range)) bool {
	fmt.Printf("\nAlgo 5:")
	return true
}
