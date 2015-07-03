package griddler

import (
	"fmt"
	"os"
)

// utility function to pause and wait for the user to press enter
func PauseEnter() {
	fmt.Println("Press Enter to continue...")
	var b []byte = make([]byte, 2)
	os.Stdin.Read(b)
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func IncOrDec(i int, reverse bool) int {
	if reverse {
		i--
	} else {
		i++
	}
	return i
}

// Special stack implementation where elements are unique
type Element struct {
	Value interface{}
}

type Stack []Element

func (st *Stack) push(elt interface{}) {
	for _, ste := range *st {
		if ste.Value == elt {
			return
		}
	}
	*st = append(*st, Element{Value: elt})
}

func (st *Stack) pop() interface{} {
	if len(*st) == 0 {
		return nil
	}
	ret := (*st)[len(*st)-1]
	*st = (*st)[0 : len(*st)-1]
	return ret.Value
}
