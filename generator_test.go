package main

import "testing"

func TestGenerateSolvedGridIsLatinSquareAndViewsValidate(t *testing.T) {
	for n := 4; n <= 9; n++ {
		for iter := 0; iter < 20; iter++ {
			grid := GenerateSolvedGrid(n)
			colTop, colBottom, rowLeft, rowRight := ComputeViews(grid)
			ok, mismatches, permErrors := ValidateGrid(grid, colTop, colBottom, rowLeft, rowRight)
			if !ok {
				t.Fatalf("n=%d iter=%d: generated grid failed self-validation, grid=%v mismatches=%v permErrors=%v",
					n, iter, grid, mismatches, permErrors)
			}
		}
	}
}

func TestComputeViewsKnownGrid(t *testing.T) {
	colTop, colBottom, rowLeft, rowRight := ComputeViews(knownGrid)
	assertIntSlice(t, "colTop", colTop, knownColTop)
	assertIntSlice(t, "colBottom", colBottom, knownColBottom)
	assertIntSlice(t, "rowLeft", rowLeft, knownRowLeft)
	assertIntSlice(t, "rowRight", rowRight, knownRowRight)
}

func assertIntSlice(t *testing.T, name string, got, want []int) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: length = %d, want %d", name, len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("%s[%d] = %d, want %d", name, i, got[i], want[i])
		}
	}
}
