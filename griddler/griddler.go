package griddler

import (
	"bufio"
	"container/heap"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	UseTrial bool
)

type Griddler struct {
	width         int
	height        int
	lines         [](*Line)
	columns       [](*Line)
	lStack        lStack
	cStack        lStack
	solveInitAlgo Algorithm
	solveAlgos    []Algorithm
	//solveQueue    chan (*Square)
}

func New() *Griddler {
	g := &Griddler{
		lStack:        lStack{},
		cStack:        lStack{},
		solveInitAlgo: solveInitAlgo,
		solveAlgos: []Algorithm{
			solveFilledRanges,
			solveEmptyRanges,
			solveAlgo6,
			solveAlgo7,
			solveAlgo8,
		},
	}
	return g
}

func (g *Griddler) error(e error, l int) error {
	return &ParseError{
		line: l,
		err:  e,
	}
}

func (g *Griddler) Load(filename string) error {
	// Loading the specified file
	gFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer gFile.Close()

	// Reading the griddler size on the first line
	gScanner := bufio.NewScanner(gFile)
	gScanner.Scan()
	firstLine := gScanner.Text()

	// the line should have the format AAAxBBB where AAA is the width and BBB the height
	line := 1
	firstLineSizes := strings.Split(firstLine, "x")
	if len(firstLineSizes) != 2 {
		return g.error(ErrInvalidGridSizeFormat, line)
	}

	width, err := strconv.Atoi(firstLineSizes[0])
	if err != nil {
		return g.error(ErrInvalidGridSizeValue, line)
	}
	g.width = width

	height, err := strconv.Atoi(firstLineSizes[1])
	if err != nil {
		return g.error(ErrInvalidGridSizeValue, line)
	}
	g.height = height

	// and init the board with it
	g.initBoard()

	// Reading the clue lines until the end of the file
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
				// TODO error
			}
			g.lines[index-1].addClues(gLineNumbers)
		case "V":
			if index > width {
				// TODO error
			}
			g.columns[index-1].addClues(gLineNumbers)
		default:
			return g.error(ErrInvalidTokenLine, line)
		}
	}
	return nil
}

func (g *Griddler) Show() {
	g.showColumnHeader()
	g.showBody()
	g.showColumnFooter()
}

func (g *Griddler) showColumnHeader() {
	fmt.Printf("    ")
	for i := 0; i < g.width; i++ {
		fmt.Printf("%d", (i+1)/10)
	}
	fmt.Println()
	fmt.Printf("    ")
	for i := 0; i < g.width; i++ {
		fmt.Printf("%d", (i+1)%10)
	}
	fmt.Println()
	fmt.Printf("   +")
	for i := 0; i < g.width; i++ {
		fmt.Printf("-")
	}
	fmt.Println("+")
}

func (g *Griddler) showBody() {
	for i := 0; i < g.height; i++ {
		fmt.Printf("%2d |", i+1)
		for j := 0; j < g.width; j++ {
			g.lines[i].squares[j].show()
		}
		fmt.Printf("| %-2d", i+1)
		if g.lines[i].isDone {
			fmt.Printf(" D")
		}
		fmt.Println()
	}
}

func (g *Griddler) showColumnFooter() {
	fmt.Printf("   +")
	for i := 0; i < g.width; i++ {
		fmt.Printf("-")
	}
	fmt.Println("+")
	fmt.Printf("    ")
	for i := 0; i < g.width; i++ {
		fmt.Printf("%d", (i+1)/10)
	}
	fmt.Println()
	fmt.Printf("    ")
	for i := 0; i < g.width; i++ {
		fmt.Printf("%d", (i+1)%10)
	}
	fmt.Println()
	fmt.Println()
	fmt.Printf("    ")
	for i := 0; i < g.width; i++ {
		if g.columns[i].isDone {
			fmt.Printf("D")
		} else {
			fmt.Printf(" ")
		}
	}
	fmt.Println()
}

func (g *Griddler) initBoard() {
	g.lines = make([](*Line), g.height)
	for i := 0; i < g.height; i++ {
		g.lines[i] = NewLine(g, i, g.width)
		for j := 0; j < g.width; j++ {
			g.lines[i].squares[j] = NewSquare(i, j, EMPTY)
		}
	}
	g.columns = make([](*Line), g.width)
	for i := 0; i < g.width; i++ {
		g.columns[i] = NewLine(g, i, g.height)
		for j := 0; j < g.height; j++ {
			g.columns[i].squares[j] = g.lines[j].squares[i]
		}
	}
	//g.solveQueue = make(chan (*Square), g.width*g.height)
}

func (g *Griddler) save() *Griddler {
	result := *g

	result.lines = make([](*Line), g.height)
	for i := 0; i < g.height; i++ {
		result.lines[i] = NewLine(&result, i, g.width)
		result.lines[i].sumBlanks = g.lines[i].sumBlanks
		result.lines[i].sumClues = g.lines[i].sumClues
		result.lines[i].totalClues = g.lines[i].totalClues
		result.lines[i].cb = g.lines[i].cb
		result.lines[i].ce = g.lines[i].ce
		result.lines[i].isDone = g.lines[i].isDone
		for j := 0; j < g.width; j++ {
			result.lines[i].squares[j] = NewSquare(i, j, g.lines[i].squares[j].value)
		}
		result.lines[i].clues = make([](*Clue), len(g.lines[i].clues))
		for j := 0; j < len(g.lines[i].clues); j++ {
			result.lines[i].clues[j] = NewClue(g.lines[i].clues[j].length)
			result.lines[i].clues[j].l = result.lines[i]
			result.lines[i].clues[j].index = j
			result.lines[i].clues[j].begin = g.lines[i].clues[j].begin
			result.lines[i].clues[j].end = g.lines[i].clues[j].end
		}
	}

	result.columns = make([](*Line), g.width)
	for i := 0; i < g.width; i++ {
		result.columns[i] = NewLine(&result, i, g.height)
		result.columns[i].sumBlanks = g.columns[i].sumBlanks
		result.columns[i].sumClues = g.columns[i].sumClues
		result.columns[i].totalClues = g.columns[i].totalClues
		result.columns[i].cb = g.columns[i].cb
		result.columns[i].ce = g.columns[i].ce
		result.columns[i].isDone = g.columns[i].isDone
		for j := 0; j < g.height; j++ {
			result.columns[i].squares[j] = result.lines[j].squares[i]
		}
		result.columns[i].clues = make([](*Clue), len(g.columns[i].clues))
		for j := 0; j < len(g.columns[i].clues); j++ {
			result.columns[i].clues[j] = NewClue(g.columns[i].clues[j].length)
			result.columns[i].clues[j].l = result.columns[i]
			result.columns[i].clues[j].index = j
			result.columns[i].clues[j].begin = g.columns[i].clues[j].begin
			result.columns[i].clues[j].end = g.columns[i].clues[j].end
		}
	}

	return &result
}

func (g *Griddler) restore(clone *Griddler) {
	for i := 0; i < clone.height; i++ {
		g.lines[i].sumBlanks = clone.lines[i].sumBlanks
		g.lines[i].sumClues = clone.lines[i].sumClues
		g.lines[i].totalClues = clone.lines[i].totalClues
		g.lines[i].cb = clone.lines[i].cb
		g.lines[i].ce = clone.lines[i].ce
		g.lines[i].isDone = clone.lines[i].isDone
		for j := 0; j < clone.width; j++ {
			g.lines[i].squares[j].value = clone.lines[i].squares[j].value
		}
		for j := 0; j < len(clone.lines[i].clues); j++ {
			g.lines[i].clues[j].begin = clone.lines[i].clues[j].begin
			g.lines[i].clues[j].end = clone.lines[i].clues[j].end
		}
	}

	for i := 0; i < clone.width; i++ {
		g.columns[i].sumBlanks = clone.columns[i].sumBlanks
		g.columns[i].sumClues = clone.columns[i].sumClues
		g.columns[i].totalClues = clone.columns[i].totalClues
		g.columns[i].cb = clone.columns[i].cb
		g.columns[i].ce = clone.columns[i].ce
		g.columns[i].isDone = clone.columns[i].isDone
		for j := 0; j < clone.height; j++ {
			g.columns[i].squares[j] = g.lines[j].squares[i]
		}
		for j := 0; j < len(clone.columns[i].clues); j++ {
			g.columns[i].clues[j].begin = clone.columns[i].clues[j].begin
			g.columns[i].clues[j].end = clone.columns[i].clues[j].end
		}
	}
}

func (g *Griddler) Solve() bool {
	g.solveInit()
	nbTrial, nbTrialSuccess := 0, 0

	for {
		fmt.Println("\nSolving")
		g.solveByLogic()

		if !g.isDone() && UseTrial {
			saved := g.save()

			pq := make(prioQueue, 0)
			selected, potential, total := g.populateForTrial(&pq)
			fmt.Printf("Entering Trial&Error phase: %d / %d / %d\n", selected, potential, total)

			hasError := false
			attempt := 1
			for !hasError {
				if pq.Len() == 0 {
					break
				}
				s := heap.Pop(&pq).(*PrioSquare)
				fmt.Printf("\rAttempt %3d / %3d", attempt, selected)
				g.SetValue(g.lines[s.x].squares[s.y], s.pvalue)
				hasError = g.solveByTrial()
				nbTrial++
				if g.isDone() {
					break
				}
				g.restore(saved)
				if hasError {
					fmt.Printf("\nFOUND (%d,%d)\n", s.x+1, s.y+1)
					nbTrialSuccess++
					if s.pvalue == FILLED {
						g.SetValue(g.lines[s.x].squares[s.y], BLANK)
					} else {
						g.SetValue(g.lines[s.x].squares[s.y], FILLED)
					}
					g.Show()
					break
				}
				attempt++
			}
		} else {
			break
		}
	}

	g.Show()
	fmt.Printf("\nTotal trial attempts: %d/%d\n", nbTrialSuccess, nbTrial)
	return g.isDone()
}

func (g *Griddler) solveInit() {
	for _, line := range g.lines {
		g.solveInitAlgo(g, line)
	}
	for _, col := range g.columns {
		g.solveInitAlgo(g, col)
	}
}

func (g *Griddler) solveByLogic() {
	defer func() {
		if r := recover(); r != nil {
			if serr, ok := r.(*SolveError); ok {
				fmt.Printf("%v\n", serr)
				fmt.Printf("Please verify your input file...\n")
				os.Exit(1)
			} else {
				panic(r)
			}
		}
	}()
	g.solveGeneric()
}

func (g *Griddler) populateForTrial(pq *prioQueue) (selected int, potential int, total int) {
	for i := 0; i < g.height; i++ {
		for j := 0; j < g.width; j++ {
			total++
			if g.lines[i].squares[j].value == EMPTY {
				potential++
				// to assign a higher priority, we check for borders and neighbours
				priority := 0
				pvalue := FILLED
				if i == 0 || i == g.height-1 {
					priority++
				}
				if j == 0 || j == g.width-1 {
					priority++
				}
				if i > 0 && g.lines[i-1].squares[j].value != EMPTY {
					if g.lines[i-1].squares[j].value == FILLED {
						pvalue = BLANK
					}
					priority++
				}
				if i < g.height-1 && g.lines[i+1].squares[j].value != EMPTY {
					if g.lines[i+1].squares[j].value == FILLED {
						pvalue = BLANK
					}
					priority++
				}
				if j > 0 && g.lines[i].squares[j-1].value != EMPTY {
					if g.lines[i].squares[j-1].value == FILLED {
						pvalue = BLANK
					}
					priority++
				}
				if j < g.width-1 && g.lines[i].squares[j+1].value != EMPTY {
					if g.lines[i].squares[j+1].value == FILLED {
						pvalue = BLANK
					}
					priority++
				}
				// we only add those
				if priority > 0 {
					selected++
					heap.Push(pq, &PrioSquare{g.lines[i].squares[j], pvalue, priority})
				}
			}
		}
	}
	return
}

// TODO add a parameter to indicate the depth of the trial
func (g *Griddler) solveByTrial() (hasError bool) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(*SolveError); ok {
				hasError = true
			} else {
				panic(r)
			}
		}
	}()
	g.solveGeneric()

	return
}

func (g *Griddler) solveGeneric() {
	l := g.lStack.pop()
	c := g.cStack.pop()
	for l != nil || c != nil {
		if l != nil && !l.isDone {
			//fmt.Printf("\n=================== checking line %d ===================\n", l.index+1)
			g.solveLine(l)
		}
		if c != nil && !c.isDone {
			//fmt.Printf("\n=================== checking column %d ===================\n", c.index+1)
			g.solveLine(c)
		}
		l = g.lStack.pop()
		c = g.cStack.pop()
	}
}

func (g *Griddler) solveLine(l *Line) {
	// if we found all clues, we can blank all remaining square
	if l.sumClues == l.totalClues {
		for _, s := range l.squares {
			if s.value == EMPTY {
				g.SetValue(s, BLANK)
			}
		}
		return
	}
	// if we found all blanks, we can set the remaining clues
	if l.sumBlanks == l.length-l.totalClues {
		for _, s := range l.squares {
			if s.value == EMPTY {
				g.SetValue(s, FILLED)
			}
		}
		return
	}

	for _, algo := range g.solveAlgos {
		if !l.isDone {
			algo(g, l)
		} else {
			return
		}
	}
}

func (g *Griddler) SetValue(s *Square, value int) {
	switch {
	case s.value == EMPTY:
		s.value = value
		g.lStack.push(g.lines[s.x])
		g.cStack.push(g.columns[s.y])
		if value == FILLED {
			g.lines[s.x].incrementClues()
			g.columns[s.y].incrementClues()
		} else {
			g.lines[s.x].incrementBlanks()
			g.columns[s.y].incrementBlanks()
		}
		//fmt.Printf("FOUND (%d,%d)\n", s.x+1, s.y+1)
		//g.solveQueue <- s
		//g.Show()
	case s.value != value:
		panic(&SolveError{s, ErrOverridingValue})
	}
}

func (g *Griddler) isDone() bool {
	for _, l := range g.lines {
		if !l.isDone {
			return false
		}
	}
	for _, c := range g.columns {
		if !c.isDone {
			return false
		}
	}
	return true
}

// TODO: evaluate the refactor of the code to be able to use range operator, mostly clues need to be double in reverse
// => try to see if using defer might help in this direction...
// TODO: try & error case on the borders, one with the line empty, the other with some clue found
// TODO: for all algorith, the way of modifying a value will be different in normal solving or trial & error,
// therefore, we should be able to have this in parameter
