package griddler

import (
	"fmt"
	"os"
)

// utility function to pause and wait for the user to press enter
func Pause() {
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

type Range struct {
	min, max int
}

func (r *Range) length() int {
	return r.max - r.min + 1
}

func (r *Range) print(prefix string) {
	fmt.Printf("%s-->Range(b:%d,e:%d)\n", prefix, r.min+1, r.max+1)
}

func IncOrDec(i int, reverse bool) int {
	if reverse {
		i--
	} else {
		i++
	}
	return i
}
