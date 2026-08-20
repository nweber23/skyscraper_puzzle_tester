package main

import (
	"strings"
	"testing"
)

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

func TestFormatViewStringOrders(t *testing.T) {
	colTop := []int{1, 2}
	colBottom := []int{3, 4}
	rowLeft := []int{5, 6}
	rowRight := []int{7, 8}

	clockwise := FormatViewString(colTop, colBottom, rowLeft, rowRight, OrderClockwise)
	wantClockwise := "1 2 7 8 4 3 6 5"
	if clockwise != wantClockwise {
		t.Errorf("clockwise = %q, want %q", clockwise, wantClockwise)
	}

	simple := FormatViewString(colTop, colBottom, rowLeft, rowRight, OrderSimple)
	wantSimple := "1 2 3 4 5 6 7 8"
	if simple != wantSimple {
		t.Errorf("simple = %q, want %q", simple, wantSimple)
	}
}

func TestBuildErrorCases(t *testing.T) {
	for _, order := range []ViewOrder{OrderClockwise, OrderSimple} {
		cases := BuildErrorCases(4, order)
		wantNames := []string{
			"no arguments", "too many arguments", "too few values", "too many values",
			"value out of range", "non-numeric characters", "empty string",
			"whitespace-only string", "leading zero", "negative number",
			"unsolvable view combination",
		}
		if len(cases) != len(wantNames) {
			t.Fatalf("order=%s: got %d cases, want %d", order, len(cases), len(wantNames))
		}
		for i, name := range wantNames {
			if cases[i].Name != name {
				t.Errorf("order=%s: case %d name = %q, want %q", order, i, cases[i].Name, name)
			}
		}

		byName := make(map[string]ErrorCase)
		for _, c := range cases {
			byName[c.Name] = c
		}

		if len(byName["no arguments"].Args) != 0 {
			t.Errorf("expected zero args for 'no arguments'")
		}
		if len(byName["too many arguments"].Args) != 2 {
			t.Errorf("expected 2 args for 'too many arguments'")
		}
		if byName["empty string"].Args[0] != "" {
			t.Errorf("expected empty string arg")
		}
		if strings.TrimSpace(byName["whitespace-only string"].Args[0]) != "" {
			t.Errorf("expected whitespace-only arg to trim to empty")
		}

		tooFewFields := len(strings.Fields(byName["too few values"].Args[0]))
		tooManyFields := len(strings.Fields(byName["too many values"].Args[0]))
		if tooFewFields != 4*4-1 {
			t.Errorf("'too few values' has %d fields, want %d", tooFewFields, 4*4-1)
		}
		if tooManyFields != 4*4+1 {
			t.Errorf("'too many values' has %d fields, want %d", tooManyFields, 4*4+1)
		}
	}
}
