package main

import (
	"fmt"
	"strconv"
	"strings"
)

// Mismatch describes one view-count discrepancy between the value derived
// from a candidate grid and the target value under test.
type Mismatch struct {
	Direction string // "colTop", "colBottom", "rowLeft", "rowRight"
	Index     int    // 0-based row/column index
	Expected  int
	Actual    int
}

// CountVisible returns how many towers are visible looking down heights
// from the near end (index 0 is closest to the viewer). A tower is visible
// if it is taller than every tower before it in the slice.
func CountVisible(heights []int) int {
	count := 0
	max := 0
	for _, h := range heights {
		if h > max {
			count++
			max = h
		}
	}
	return count
}

// isPermutation reports whether values is a permutation of 1..n.
func isPermutation(values []int, n int) bool {
	if len(values) != n {
		return false
	}
	seen := make([]bool, n+1)
	for _, v := range values {
		if v < 1 || v > n || seen[v] {
			return false
		}
		seen[v] = true
	}
	return true
}

// reverse returns a new slice with the elements of s in reverse order.
func reverse(s []int) []int {
	out := make([]int, len(s))
	for i, v := range s {
		out[len(s)-1-i] = v
	}
	return out
}

// ValidateGrid checks that grid is an NxN Latin square (every row and
// column a permutation of 1..N) and that the 4N visibility counts derived
// from grid match colTop, colBottom, rowLeft, rowRight.
//
// Reading directions:
//   - rowLeft[i]:  row i viewed from the left  (columns 0..N-1 order)
//   - rowRight[i]: row i viewed from the right (columns N-1..0 order)
//   - colTop[j]:   column j viewed from the top    (rows 0..N-1 order)
//   - colBottom[j]:column j viewed from the bottom (rows N-1..0 order)
//
// ok is true only when every check passes. When permutation checks fail,
// mismatches is nil (view checks are meaningless on a malformed grid) and
// permErrors lists every offending row/column. Otherwise permErrors is nil
// and mismatches lists every view-count discrepancy found (empty when ok).
func ValidateGrid(grid [][]int, colTop, colBottom, rowLeft, rowRight []int) (ok bool, mismatches []Mismatch, permErrors []string) {
	n := len(grid)
	if n == 0 || len(colTop) != n || len(colBottom) != n || len(rowLeft) != n || len(rowRight) != n {
		return false, nil, []string{"grid size does not match view slice lengths"}
	}

	for i, row := range grid {
		if len(row) != n {
			permErrors = append(permErrors, fmt.Sprintf("row %d has %d columns, want %d", i, len(row), n))
			continue
		}
		if !isPermutation(row, n) {
			permErrors = append(permErrors, fmt.Sprintf("row %d is not a permutation of 1..%d", i, n))
		}
	}

	for j := 0; j < n; j++ {
		col := make([]int, n)
		malformed := false
		for i := 0; i < n; i++ {
			if len(grid[i]) != n {
				malformed = true
				break
			}
			col[i] = grid[i][j]
		}
		if malformed || !isPermutation(col, n) {
			permErrors = append(permErrors, fmt.Sprintf("column %d is not a permutation of 1..%d", j, n))
		}
	}

	if len(permErrors) > 0 {
		return false, nil, permErrors
	}

	for i := 0; i < n; i++ {
		row := grid[i]
		if got := CountVisible(row); got != rowLeft[i] {
			mismatches = append(mismatches, Mismatch{"rowLeft", i, rowLeft[i], got})
		}
		if got := CountVisible(reverse(row)); got != rowRight[i] {
			mismatches = append(mismatches, Mismatch{"rowRight", i, rowRight[i], got})
		}
	}

	for j := 0; j < n; j++ {
		col := make([]int, n)
		for i := 0; i < n; i++ {
			col[i] = grid[i][j]
		}
		if got := CountVisible(col); got != colTop[j] {
			mismatches = append(mismatches, Mismatch{"colTop", j, colTop[j], got})
		}
		if got := CountVisible(reverse(col)); got != colBottom[j] {
			mismatches = append(mismatches, Mismatch{"colBottom", j, colBottom[j], got})
		}
	}

	return len(mismatches) == 0, mismatches, nil
}

// ParseGrid extracts an NxN grid from the tested binary's stdout. It
// accepts one row per non-blank line, columns separated by whitespace, and
// tolerates blank lines around the grid. Returns ok=false if the output
// doesn't contain exactly n non-blank lines of exactly n integers each.
func ParseGrid(output string, n int) ([][]int, bool) {
	if n <= 0 {
		return nil, false
	}
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return nil, false
	}

	lines := strings.Split(trimmed, "\n")
	rows := make([][]int, 0, n)
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if len(fields) != n {
			return nil, false
		}
		row := make([]int, n)
		for i, f := range fields {
			v, err := strconv.Atoi(f)
			if err != nil {
				return nil, false
			}
			row[i] = v
		}
		rows = append(rows, row)
	}

	if len(rows) != n {
		return nil, false
	}
	return rows, true
}

// TestResult is the outcome of running and evaluating one TestCase.
type TestResult struct {
	Case       TestCase
	Passed     bool
	Reason     string
	Stdout     string
	Mismatches []Mismatch
	PermErrors []string
	ParsedGrid [][]int // set when the binary's output parsed as an NxN grid
	TimedOut   bool
}

// EvaluateResult judges a raw subprocess Result against the expectations
// encoded in TestCase:
//   - For fuzz cases (well-formed, solvable input), the binary must print a
//     grid that satisfies ValidateGrid against the case's target views.
//   - For error cases (deliberately invalid input), the binary must either
//     print exactly "Error" or otherwise not print a valid-looking grid.
//   - Any exec-level failure or timeout is always a failure.
func EvaluateResult(tc TestCase, res Result) TestResult {
	tr := TestResult{Case: tc, Stdout: res.Stdout, TimedOut: res.TimedOut}

	if res.Err != nil {
		tr.Passed = false
		tr.Reason = "exec error: " + res.Err.Error()
		return tr
	}

	if res.TimedOut {
		tr.Passed = false
		tr.Reason = "timed out"
		return tr
	}

	if tc.IsFuzz {
		grid, ok := ParseGrid(res.Stdout, tc.N)
		if !ok {
			tr.Passed = false
			tr.Reason = "output did not contain a parseable NxN grid"
			return tr
		}
		tr.ParsedGrid = grid
		valid, mismatches, permErrors := ValidateGrid(grid, tc.ColTop, tc.ColBottom, tc.RowLeft, tc.RowRight)
		tr.Mismatches = mismatches
		tr.PermErrors = permErrors
		tr.Passed = valid
		if !valid {
			tr.Reason = "grid failed constraint validation"
		}
		return tr
	}

	if strings.TrimSpace(res.Stdout) == "Error" {
		tr.Passed = true
		tr.Reason = "printed Error as expected"
		return tr
	}
	if grid, ok := ParseGrid(res.Stdout, tc.N); ok {
		tr.ParsedGrid = grid
		tr.Passed = false
		tr.Reason = "printed a grid for invalid input instead of failing"
		return tr
	}
	tr.Passed = true
	tr.Reason = "did not print a valid grid (no crash)"
	return tr
}
