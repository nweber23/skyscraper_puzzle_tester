package main

import "math/rand"

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
