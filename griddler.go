package griddler

type griddler interface {
	// each call to solveStep will return the next square being solved
	solveStep() gridSquare
	// solve will return the griddler fully solved
	solve()
}

type gBoard struct {
	width, height int
}

type gSquare struct {
	x, y int
}
