package main

import "testing"

func TestCountVisible(t *testing.T) {
	cases := []struct {
		name   string
		height []int
		want   int
	}{
		{"increasing", []int{1, 2, 3, 4}, 4},
		{"decreasing", []int{4, 3, 2, 1}, 1},
		{"single peak first", []int{4, 1, 2, 3}, 1},
		{"mixed", []int{2, 1, 4, 3}, 2},
		{"empty", []int{}, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := CountVisible(c.height); got != c.want {
				t.Errorf("CountVisible(%v) = %d, want %d", c.height, got, c.want)
			}
		})
	}
}

// knownGrid is a hand-verified 4x4 Latin square used across tests.
var knownGrid = [][]int{
	{1, 2, 3, 4},
	{2, 1, 4, 3},
	{3, 4, 1, 2},
	{4, 3, 2, 1},
}

// Views for knownGrid, computed by hand:
//   rowLeft  = [4,2,2,1]  rowRight  = [1,2,2,4]
//   colTop   = [4,2,2,1]  colBottom = [1,2,2,4]
var (
	knownColTop    = []int{4, 2, 2, 1}
	knownColBottom = []int{1, 2, 2, 4}
	knownRowLeft   = []int{4, 2, 2, 1}
	knownRowRight  = []int{1, 2, 2, 4}
)

func TestValidateGridValid(t *testing.T) {
	ok, mismatches, permErrors := ValidateGrid(knownGrid, knownColTop, knownColBottom, knownRowLeft, knownRowRight)
	if !ok {
		t.Fatalf("expected valid grid, got mismatches=%v permErrors=%v", mismatches, permErrors)
	}
	if len(mismatches) != 0 || len(permErrors) != 0 {
		t.Fatalf("expected no errors, got mismatches=%v permErrors=%v", mismatches, permErrors)
	}
}

func TestValidateGridTamperedView(t *testing.T) {
	badRowLeft := []int{3, 2, 2, 1} // rowLeft[0] should be 4, not 3
	ok, mismatches, permErrors := ValidateGrid(knownGrid, knownColTop, knownColBottom, badRowLeft, knownRowRight)
	if ok {
		t.Fatalf("expected invalid result for tampered view")
	}
	if len(permErrors) != 0 {
		t.Fatalf("expected no permutation errors, got %v", permErrors)
	}
	if len(mismatches) != 1 {
		t.Fatalf("expected exactly 1 mismatch, got %v", mismatches)
	}
	want := Mismatch{"rowLeft", 0, 3, 4}
	if mismatches[0] != want {
		t.Errorf("mismatch = %+v, want %+v", mismatches[0], want)
	}
}

func TestValidateGridNonPermutationRow(t *testing.T) {
	badGrid := [][]int{
		{1, 1, 3, 4}, // duplicate 1
		{2, 1, 4, 3},
		{3, 4, 1, 2},
		{4, 3, 2, 1},
	}
	ok, mismatches, permErrors := ValidateGrid(badGrid, knownColTop, knownColBottom, knownRowLeft, knownRowRight)
	if ok {
		t.Fatalf("expected invalid result for non-permutation row")
	}
	if mismatches != nil {
		t.Fatalf("expected nil mismatches when permutation checks fail, got %v", mismatches)
	}
	if len(permErrors) == 0 {
		t.Fatalf("expected at least one permutation error")
	}
}

func TestValidateGridSizeMismatch(t *testing.T) {
	ok, _, permErrors := ValidateGrid(knownGrid, []int{1, 2, 3}, knownColBottom, knownRowLeft, knownRowRight)
	if ok {
		t.Fatalf("expected invalid result for mismatched view slice length")
	}
	if len(permErrors) == 0 {
		t.Fatalf("expected an error describing the size mismatch")
	}
}
