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
	solveInitAlgo  Algorithm
	solveAlgos     []Algorithm
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
		solveInitAlgo:  solveInitAlgo,
		solveAlgos: []Algorithm{
			solveFilledRange,
			//solveAlgo1,
			//solveAlgo9,
			//solveAlgo2,
			//solveAlgo3,
			//solveAlgo4,
			//solveAlgo5,
			//solveAlgo6,
			//solveAlgo7,
			//solveAlgo8,
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

func (g *Griddler) incrementSolvedLines() {
	g.sumSolvedLines++
	if g.sumSolvedLines == g.width+g.height {
		g.isDone = true
	}
}

func (g *Griddler) Solve() {
	g.solveInit()
	l := g.lStack.pop()
	c := g.cStack.pop()
	for l != nil || c != nil {
		if l != nil && !l.isDone {
			fmt.Printf("\n=================== checking line %d ===================\n", l.index+1)
			Pause()
			g.solveLine(l)
		}
		if c != nil && !c.isDone {
			fmt.Printf("\n=================== checking column %d ===================\n", c.index+1)
			Pause()
			g.solveLine(c)
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
			g.setValue(s, 1)
		}
		return
	}
	// if we found all blanks, we can set the remaining clues
	if l.sumBlanks == l.length-l.totalClues {
		for _, s := range l.squares {
			g.setValue(s, 2)
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

// TODO: similar to algo 5, check empty surrounded by blank that does not fit any clue in size
// verify that it is not already covered by algo 2..., or somehow badly implemented
// TODO: find a general use-case to be able to solve this pattern:
// ......XX.X00.000 with (4,1) -> we can blank the beginning
// TODO: try & error case on the borders, one with the line empty, the other with some clue found
// TODO: refactor the whole function to suppress from the boolean returned to print the summary of execution for example,
// exception in setValue() will be used to check a bad line solving during trial&error
// TODO: evaluate the refactor of the code to be able to use range operator, mostly clues need to be double in reverse
// TODO: for all algorith, the way of modifying a value will be different in normal solving or trial & error,
// therefore, we should be able to have this in parameter
