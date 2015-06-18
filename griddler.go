package griddler

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func Pause() {
	var b []byte = make([]byte, 2)
	os.Stdin.Read(b)
}

type Square struct {
	x, y int
	val  int
	g    *Griddler
}

func NewSquare(x, y, v int, g *Griddler) *Square {
	return &Square{
		x:   x,
		y:   y,
		val: v,
		g:   g,
	}
}

func (s *Square) setValue(val int) {
	if s.val == 0 {
		s.val = val
		s.g.s.push(s)
		if val != 1 {
			s.g.lines[s.x].incrementClueTotal()
			s.g.columns[s.y].incrementClueTotal()
		}
		s.g.lines[s.x].increment()
		s.g.columns[s.y].increment()
		fmt.Printf("FOUND (%d,%d)\n", s.x+1, s.y+1)
		s.g.Show()
		Pause()
	}
}

func (s Square) show() {
	//fmt.Printf("(%d,%d,", s.x, s.y)
	switch s.val {
	case 0:
		fmt.Printf(" ")
	case 1:
		fmt.Printf(".")
	case 2:
		fmt.Printf("X")
	}
	//fmt.Printf(")")
}

type Stack [](*Square)

func (st *Stack) push(sq *Square) {
	*st = append(*st, sq)
}

func (s *Stack) pop() *Square {
	if len(*s) == 0 {
		return nil
	}
	ret := (*s)[len(*s)-1]
	*s = (*s)[0 : len(*s)-1]
	return ret
}

type Line struct {
	g        *Griddler
	length   int
	squares  [](*Square)
	clues    [](*Clue)
	sum      int
	sumClues int
	isDone   bool
	total    int
}

func NewLine(g *Griddler, i, length int) *Line {
	return &Line{
		g:        g,
		length:   length,
		squares:  make([](*Square), length),
		total:    0,
		sum:      0,
		sumClues: 0,
		isDone:   false,
	}
}

func (l *Line) addClues(cs [](*Clue)) {
	l.clues = cs
	for i, val := range cs {
		l.total += val.length
		val.l = l
		val.index = i
	}
}

func (l *Line) increment() {
	l.sum++
	if (l.sum == l.length) && !l.isDone {
		l.isDone = true
		l.g.increment()
	}
}

func (l *Line) incrementClueTotal() {
	l.sumClues++
}

func (l *Line) incrementCluesBegin(index, n int) {
	for i, c := range l.clues {
		if i > index {
			c.begin += n
		}
	}
}

func (l *Line) decrementCluesEnd(index, n int) {
	for i, c := range l.clues {
		if i < index {
			c.end -= n
		}
	}
}

type Clue struct {
	index      int
	length     int
	begin, end int
	isDone     bool
	l          *Line
}

func NewClue(l int) *Clue {
	return &Clue{
		length: l,
		isDone: false,
	}
}

func (c *Clue) solve() {
	diff := c.begin + c.length - (c.end + 1 - c.length)
	if diff > 0 {
		for j := 0; j < diff; j++ {
			c.l.squares[c.end-c.length+1+j].setValue(2)
		}
	}
}

type ParseError struct {
	line int
	err  error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("Error line %d: %s", e.line, e.err)
}

var (
	ErrInvalidGridSizeFormat = errors.New("invalid format for first line")
	ErrInvalidGridSizeValue  = errors.New("invalid value for griddler size")
	ErrMissingSemiColon      = errors.New("missing semicolon to delimit line info and values")
	ErrInvalidIntValue       = errors.New("invalid integer for value(s)")
	ErrInvalidIntLine        = errors.New("invalid integer for line info")
	ErrInvalidTokenLine      = errors.New("invalid starting token for line info")
	ErrTooManyLine           = errors.New("too many line compared to the size specified")
)

type Griddler struct {
	width   int
	height  int
	lines   [](*Line)
	columns [](*Line)
	sum     int
	isDone  bool
	s       *Stack
}

func NewGriddler() *Griddler {
	g := &Griddler{
		sum:    0,
		isDone: false,
		s:      &Stack{},
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
}

func (g *Griddler) increment() {
	g.sum++
	if g.sum == g.width+g.height {
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
	s := g.s.pop()
	for s != nil {
		if g.isDone {
			fmt.Println("Griddler completed!!!")
			return
		} else {
			fmt.Printf("\nchecking line %d\n", s.x+1)
			g.checkLine(s.g.lines[s.x])
			//Pause()
			fmt.Printf("\nchecking column %d\n", s.y+1)
			g.checkLine(s.g.columns[s.y])
			//Pause()
		}
		s = g.s.pop()
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
	for _, line := range g.columns {
		g.solveInitLine(line)
	}
}

func (g *Griddler) solveInitLine(line *Line) {
	switch {
	case line.total == 0:
		for _, s := range line.squares {
			s.setValue(1)
		}
	case line.total == g.width:
		for _, s := range line.squares {
			s.setValue(2)
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
					line.squares[clue.end-clue.length+1+j].setValue(2)
				}
			}
		}
	}
}

func (g *Griddler) checkLine(l *Line) bool {
	// if the line is done, return
	if l.isDone {
		return true
	}
	// if we found all clues, we can blank all remaining square
	if l.sumClues == l.total {
		for _, s := range l.squares {
			s.setValue(1)
		}
		return true
	}
	// if the remaining squares are clues, we can set them
	if (l.length - l.sum) == (l.total - l.sumClues) {
		for _, s := range l.squares {
			s.setValue(2)
		}
		return true
	}
	// algo 1, fill the beginning and end of line if possible
	res := g.checkLineAlgo1(l)
	if !res {
		return false
	}

	// algo 2, update the valid range for each clue and update possible values
	res = g.checkLineAlgo2(l)
	if !res {
		return false
	}

	return true
}

func (g *Griddler) checkLineAlgo1(l *Line) bool {
	// From the beginning
	fmt.Println("From beginning:")
	for _, cb := range l.clues {
		if cb.isDone {
			continue
		}
		fmt.Printf("\nClue(n:%d,b:%d,e:%d,l:%d):", cb.index+1, cb.begin+1, cb.end+1, cb.length)
		if !g.checkClueAlgo1(cb, false) {
			return false
		}
		if cb.isDone {
			continue
		}
		break
	}

	// From the end (note: sadly, range cannot be used for reverse iteration)
	fmt.Println("\nFrom end:")
	for i := len(l.clues); i > 0; i-- {
		cb := l.clues[i-1]
		if cb.isDone {
			continue
		}
		fmt.Printf("\nClue(n:%d,b:%d,e:%d,l:%d):", cb.index+1, cb.begin+1, cb.end+1, cb.length)
		if !g.checkClueAlgo1(cb, true) {
			return false
		}
		if cb.isDone {
			continue
		}
		break
	}
	return true
}

func (g *Griddler) checkClueAlgo1(c *Clue, reverse bool) bool {
	empty := 0
	filled := 0
	l := c.l
	i := c.begin
	if reverse {
		i = c.end
	}
	fmt.Printf("Algo 1:")
	for {
		switch {
		case l.squares[i].val == 0:
			fmt.Printf("(%d,%d).", l.squares[i].x+1, l.squares[i].y+1)
			if filled > 0 {
				// we can potentially fill more square
				for j := 0; j < c.length-(filled+empty); j++ {
					fmt.Printf("Filling with empty:%d, filled:%d length:%d", empty, filled, c.length)
					if reverse {
						l.squares[i-j].setValue(2)
					} else {
						l.squares[i+j].setValue(2)
					}
				}
				// we can potentially end with a blank square if it was the first empty
				if empty == 0 {
					fmt.Printf("Filling ended with a blank")
					if reverse {
						l.squares[i-c.length+filled].setValue(1)
					} else {
						l.squares[i+c.length-filled].setValue(1)
					}
					c.isDone = true
					return true
				}
			}
			empty++
			if empty+filled > c.length {
				return true
			}
		case l.squares[i].val == 1:
			fmt.Printf("(%d,%d)0", l.squares[i].x+1, l.squares[i].y+1)
			if filled != 0 {
				switch {
				// we can pursue the filling going backward (ex: .X0)
				case filled < c.length:
					for j := 0; j < c.length; j++ {
						if reverse {
							l.squares[i+1+j].setValue(2)
						} else {
							l.squares[i-1-j].setValue(2)
						}
					}
					// ending with a blank if we don't reach the border
					if reverse {
						if i+1+c.length < l.length {
							l.squares[i+c.length+1].setValue(1)
						}
					} else {
						if i-1-c.length > 0 {
							l.squares[i-c.length-1].setValue(1)
						}
					}
					c.isDone = true
					return true
				// this clue is done and correct (ex: XX0)
				case filled == c.length:
					c.isDone = true
					return true
				// we filled too much, we return false (ex: XXX)
				case filled > c.length:
					fmt.Println("WWAARRNNIINNGG: filled > c.length 1")
					return false
				}
			}
			if empty != 0 {
				// if no place for this clue, we can blank
				if empty < c.length {
					for j := 0; j < empty; j++ {
						if reverse {
							l.squares[i+1+j].setValue(1)
						} else {
							l.squares[i-1-j].setValue(1)
						}
					}
				}
			}
		case l.squares[i].val == 2:
			fmt.Printf("(%d,%d)X", l.squares[i].x+1, l.squares[i].y+1)
			filled++
			switch {
			// nothing to do we continue
			case filled < c.length:
			// here we found it all, we can blank the beginning and the end
			case filled == c.length:
				if reverse {
					l.squares[i-1].setValue(1)
					if i < c.end-c.length {
						l.squares[i+c.length].setValue(1)
					}
				} else {
					l.squares[i+1].setValue(1)
					if i > c.begin+c.length {
						l.squares[i-c.length].setValue(1)
					}
				}
				c.isDone = true
			case filled > c.length:
				fmt.Println("WWAARRNNIINNGG: filled > c.length 2")
				return false
			}
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

func (g *Griddler) checkLineAlgo2(l *Line) bool {
	for _, cb := range l.clues {
		if cb.isDone {
			continue
		}
		fmt.Printf("\nClue(n:%d,b:%d,e:%d,l:%d):", cb.index+1, cb.begin+1, cb.end+1, cb.length)
		if !g.checkClueAlgo2(cb, false) {
			return false
		}
		if !g.checkClueAlgo2(cb, true) {
			return false
		}
		cb.solve()
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
	fmt.Printf("Algo 2:")
	for {
		switch {
		case l.squares[i].val == 0:
			fmt.Printf("(%d,%d).", l.squares[i].x+1, l.squares[i].y+1)
			empty++
		case l.squares[i].val == 1:
			fmt.Printf("(%d,%d)0", l.squares[i].x+1, l.squares[i].y+1)
			if (empty + filled) < c.length {
				if reverse {
					l.decrementCluesEnd(c.index, empty+filled)
				} else {
					l.incrementCluesBegin(c.index, empty+filled)
				}
			}
		case l.squares[i].val == 2:
			fmt.Printf("(%d,%d)X", l.squares[i].x+1, l.squares[i].y+1)
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
