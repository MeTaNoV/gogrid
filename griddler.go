package griddler

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Griddler struct {
	width         int
	height        int
	lines         [](*Line)
	columns       [](*Line)
	lStack        Stack
	cStack        Stack
	solveInitAlgo Algorithm
	solveAlgos    []Algorithm
	solveQueue    chan (*Square)
}

func New() *Griddler {
	g := &Griddler{
		lStack:        Stack{},
		cStack:        Stack{},
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

func (g *Griddler) SetValue(s *Square, val int) {
	if s.val == EMPTY {
		s.val = val
		g.lStack.push(g.lines[s.x])
		g.cStack.push(g.columns[s.y])
		if val == FILLED {
			g.lines[s.x].incrementClues()
			g.columns[s.y].incrementClues()
		} else {
			g.lines[s.x].incrementBlanks()
			g.columns[s.y].incrementBlanks()
		}
		fmt.Printf("FOUND (%d,%d)\n", s.x+1, s.y+1)
		g.solveQueue <- s
		//g.Show()
		//Pause()
	}
}

func (g *Griddler) Solve() {
	g.solveInit()
	l := g.lStack.pop()
	c := g.cStack.pop()
	for l != nil || c != nil {
		if l != nil && !l.isDone {
			fmt.Printf("\n=================== checking line %d ===================\n", l.index+1)
			//Pause()
			g.solveLine(l)
		}
		if c != nil && !c.isDone {
			fmt.Printf("\n=================== checking column %d ===================\n", c.index+1)
			//Pause()
			g.solveLine(c)
		}
		l = g.lStack.pop()
		c = g.cStack.pop()
	}
	if g.isDone() {
		fmt.Println("Griddler completed!!!")
	} else {
		fmt.Println("Griddler uncompleted, find new search algorithm!")
	}
}

func (g *Griddler) solveInit() {
	for _, line := range g.lines {
		g.solveInitAlgo(g, line)
	}
	for _, col := range g.columns {
		g.solveInitAlgo(g, col)
	}
}

func (g *Griddler) solveLine(l *Line) {
	// if we found all clues, we can blank all remaining square
	if l.sumClues == l.totalClues {
		for _, s := range l.squares {
			g.SetValue(s, BLANK)
		}
		return
	}
	// if we found all blanks, we can set the remaining clues
	if l.sumBlanks == l.length-l.totalClues {
		for _, s := range l.squares {
			g.SetValue(s, FILLED)
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
// TODO: try & error case on the borders, one with the line empty, the other with some clue found
// TODO: for all algorith, the way of modifying a value will be different in normal solving or trial & error,
// therefore, we should be able to have this in parameter
