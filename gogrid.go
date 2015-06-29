package main

import (
	"flag"
	"fmt"
	"github.com/MeTaNoV/gogrid/griddler"
	"os"
)

var fileName string

func initFlags() {
	const (
		defaultFilename = "default.grid"
		usageFilename   = "name of the griddler file to load."
	)
	flag.StringVar(&fileName, "file", defaultFilename, usageFilename)
	flag.StringVar(&fileName, "f", defaultFilename, usageFilename)

	flag.Parse()

	if fileName == defaultFilename {
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	initFlags()

	gBoard := griddler.New()
	err := gBoard.Load(fileName)
	if err != nil {
		fmt.Printf("Error loading file: %v\n", err)
		return
	}

	isSolved := gBoard.Solve()
	if isSolved {
		fmt.Println("Griddler completed!!!")
	} else {
		fmt.Println("Griddler uncompleted, find new search algorithm!")
	}
	//gBoard.Show()
}
