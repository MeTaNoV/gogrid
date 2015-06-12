package main

import (
	"fmt"
	"os"
)

const VMIN = 10
const HMIN = 10

func main() {
	if len(os.Args) > 2 {
		fmt.Println("Too many arguments")
	}

	gFileName := os.Args[1]
	fmt.Println("Reading "+ gFileName)
	
	gFile, err := os.Open(gFileName)
	if err != nil {
		fmt.Println("Error opening file: %v\n", err)
		return 
	}

	gReader := csv.NewReader(gFile)
	gReader.Comma = ','

	vlines := make([]int, VMIN)
	hlines := make([]int, HMIN)
	lineNumber := 0

	for gLine, err := gReader.Read() {
		if err != nil {
			fmt.Println("Error reading file: %v\n", err)
			return
		}
		if gLine == nil {
			break
		}
		switch gLine[0] {
		case 'V':
			vline := make([]int, len(gLine)-1)
			for i := 1; i < len(gLine); i++ {
				vline[i-1] = gLine[i]
			}
			vlines.append(vline)
		case 'H':
			hline := make([]int, len(gLine)-1)
			for i := 1; i < len(gLine); i++ {
				hline[i-1] = gLine[i]
			}
			hlines.append(hline)
		default:
			fmt.Println("Error in line %d, \n", lineNumber)
		}
		lineNumber++
	}


}