package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
)

// GenerateSolvedGrid produces a random NxN grid where every row and column
// is a permutation of 1..N, using randomized backtracking: at each cell it
// tries candidate values in random order and backtracks on dead ends.
func GenerateSolvedGrid(n int) [][]int {
	grid := make([][]int, n)
	for i := range grid {
		grid[i] = make([]int, n)
	}

	var fill func(pos int) bool
	fill = func(pos int) bool {
		if pos == n*n {
			return true
		}
		row, col := pos/n, pos%n

		for _, c := range rand.Perm(n) {
			v := c + 1
			if rowHas(grid, row, v) || colHas(grid, col, v) {
				continue
			}
			grid[row][col] = v
			if fill(pos + 1) {
				return true
			}
			grid[row][col] = 0
		}
		return false
	}

	fill(0)
	return grid
}

func rowHas(grid [][]int, row, v int) bool {
	for _, x := range grid[row] {
		if x == v {
			return true
		}
	}
	return false
}

func colHas(grid [][]int, col, v int) bool {
	for _, r := range grid {
		if r[col] == v {
			return true
		}
	}
	return false
}

// ComputeViews derives the 4N target view values from a solved grid. See
// ValidateGrid's doc comment for the meaning of each returned slice.
func ComputeViews(grid [][]int) (colTop, colBottom, rowLeft, rowRight []int) {
	n := len(grid)
	colTop = make([]int, n)
	colBottom = make([]int, n)
	rowLeft = make([]int, n)
	rowRight = make([]int, n)

	for i := 0; i < n; i++ {
		row := grid[i]
		rowLeft[i] = CountVisible(row)
		rowRight[i] = CountVisible(reverse(row))
	}

	for j := 0; j < n; j++ {
		col := make([]int, n)
		for i := 0; i < n; i++ {
			col[i] = grid[i][j]
		}
		colTop[j] = CountVisible(col)
		colBottom[j] = CountVisible(reverse(col))
	}

	return colTop, colBottom, rowLeft, rowRight
}

// ViewOrder controls how the 4N view values are serialized into the single
// argument string passed to the tested binary.
type ViewOrder string

const (
	// OrderClockwise lists values clockwise starting at the top-left: top
	// (left-to-right), right (top-to-bottom), bottom (right-to-left), left
	// (bottom-to-top). This is the most common Rush01 subject convention.
	OrderClockwise ViewOrder = "clockwise"
	// OrderSimple lists values as four straight blocks: top, bottom, left,
	// right, each in left-to-right / top-to-bottom order.
	OrderSimple ViewOrder = "simple"
)

// FormatViewString serializes the 4N view values into the single
// space-separated argument string the tested binary expects.
func FormatViewString(colTop, colBottom, rowLeft, rowRight []int, order ViewOrder) string {
	values := make([]int, 0, 4*len(colTop))

	switch order {
	case OrderSimple:
		values = append(values, colTop...)
		values = append(values, colBottom...)
		values = append(values, rowLeft...)
		values = append(values, rowRight...)
	default: // OrderClockwise
		values = append(values, colTop...)
		values = append(values, rowRight...)
		values = append(values, reverse(colBottom)...)
		values = append(values, reverse(rowLeft)...)
	}

	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = strconv.Itoa(v)
	}
	return strings.Join(parts, " ")
}

// ErrorCase describes one invalid-input scenario to feed the tested binary.
// A correct implementation is expected to reject it (print "Error" and/or
// exit non-zero) without printing a valid-looking grid.
type ErrorCase struct {
	Name string
	Args []string
}

// BuildErrorCases returns one ErrorCase per invalid-input category for grid
// size n: argument count, value count, value range, non-numeric input,
// empty/whitespace input, leading zero, negative number, and a
// well-formed-but-unsolvable clue combination. Requires n >= 2.
func BuildErrorCases(n int, order ViewOrder) []ErrorCase {
	grid := GenerateSolvedGrid(n)
	colTop, colBottom, rowLeft, rowRight := ComputeViews(grid)
	validString := FormatViewString(colTop, colBottom, rowLeft, rowRight, order)
	values := strings.Fields(validString)

	tooFew := strings.Join(values[:len(values)-1], " ")
	tooMany := validString + " 1"

	outOfRange := append([]string(nil), values...)
	outOfRange[0] = strconv.Itoa(n + 1)

	nonNumeric := append([]string(nil), values...)
	nonNumeric[0] = "a"

	leadingZero := append([]string(nil), values...)
	leadingZero[0] = "0" + leadingZero[0]

	negative := append([]string(nil), values...)
	negative[0] = "-" + negative[0]

	// A row that is fully increasing (rowLeft == n) forces a strictly
	// increasing sequence, whose reverse is strictly decreasing and thus has
	// visibility 1 from the other end. So rowLeft[0] == n and rowRight[0] ==
	// n can never both hold for n > 1 — this combination is provably
	// unsolvable without needing a solver. See TestVisibilityBothEndsMaxImpossible.
	unsolvableRowLeft := append([]int(nil), rowLeft...)
	unsolvableRowRight := append([]int(nil), rowRight...)
	unsolvableRowLeft[0] = n
	unsolvableRowRight[0] = n
	unsolvableString := FormatViewString(colTop, colBottom, unsolvableRowLeft, unsolvableRowRight, order)

	return []ErrorCase{
		{Name: "no arguments", Args: []string{}},
		{Name: "too many arguments", Args: []string{validString, "extra"}},
		{Name: "too few values", Args: []string{tooFew}},
		{Name: "too many values", Args: []string{tooMany}},
		{Name: "value out of range", Args: []string{strings.Join(outOfRange, " ")}},
		{Name: "non-numeric characters", Args: []string{strings.Join(nonNumeric, " ")}},
		{Name: "empty string", Args: []string{""}},
		{Name: "whitespace-only string", Args: []string{"   "}},
		{Name: "leading zero", Args: []string{strings.Join(leadingZero, " ")}},
		{Name: "negative number", Args: []string{strings.Join(negative, " ")}},
		{Name: "unsolvable view combination", Args: []string{unsolvableString}},
	}
}

// TestCase is one fully-built test to run against the tested binary: a
// fuzz case (well-formed, solvable input with known target views) or an
// error case (deliberately invalid input).
type TestCase struct {
	Name      string
	IsFuzz    bool
	N         int
	Grid      [][]int // nil for error cases
	ColTop    []int
	ColBottom []int
	RowLeft   []int
	RowRight  []int
	Args      []string
}

// BuildTestQueue builds the full ordered list of test cases for one run:
// `runs` random fuzz cases followed by one case per BuildErrorCases
// category, all for an NxN grid. Fuzz case generation is CPU-bound (random
// backtracking per grid) and independent per case, so it's fanned out
// across workers goroutines - for large `runs` this is the difference
// between a near-instant queue and a multi-second freeze before the first
// test even starts.
func BuildTestQueue(n, runs int, order ViewOrder, workers int) []TestCase {
	errorCases := BuildErrorCases(n, order)
	cases := make([]TestCase, runs+len(errorCases))

	if workers < 1 {
		workers = 1
	}
	if workers > runs {
		workers = runs
	}

	if runs > 0 {
		jobs := make(chan int)
		var wg sync.WaitGroup
		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := range jobs {
					grid := GenerateSolvedGrid(n)
					colTop, colBottom, rowLeft, rowRight := ComputeViews(grid)
					args := []string{FormatViewString(colTop, colBottom, rowLeft, rowRight, order)}
					cases[i] = TestCase{
						Name:      fmt.Sprintf("fuzz #%d", i+1),
						IsFuzz:    true,
						N:         n,
						Grid:      grid,
						ColTop:    colTop,
						ColBottom: colBottom,
						RowLeft:   rowLeft,
						RowRight:  rowRight,
						Args:      args,
					}
				}
			}()
		}
		for i := 0; i < runs; i++ {
			jobs <- i
		}
		close(jobs)
		wg.Wait()
	}

	for i, ec := range errorCases {
		cases[runs+i] = TestCase{
			Name:   ec.Name,
			IsFuzz: false,
			N:      n,
			Args:   ec.Args,
		}
	}

	return cases
}
