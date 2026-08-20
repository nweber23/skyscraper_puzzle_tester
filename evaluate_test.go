package main

import (
	"testing"
)

func TestParseGridValidAndInvalid(t *testing.T) {
	grid, ok := ParseGrid("1 2 3 4\n2 1 4 3\n3 4 1 2\n4 3 2 1\n", 4)
	if !ok {
		t.Fatalf("expected valid parse")
	}
	if len(grid) != 4 || len(grid[0]) != 4 {
		t.Fatalf("unexpected grid shape: %v", grid)
	}

	if _, ok := ParseGrid("1 2 3\n2 1 4 3\n3 4 1 2\n4 3 2 1\n", 4); ok {
		t.Errorf("expected parse failure for wrong column count")
	}
	if _, ok := ParseGrid("1 2 3 4\n2 1 4 3\n3 4 1 2\n", 4); ok {
		t.Errorf("expected parse failure for wrong row count")
	}
	if _, ok := ParseGrid("a b c d\n2 1 4 3\n3 4 1 2\n4 3 2 1\n", 4); ok {
		t.Errorf("expected parse failure for non-numeric content")
	}
	if _, ok := ParseGrid("", 4); ok {
		t.Errorf("expected parse failure for empty output")
	}
}

func fuzzCaseFromKnownGrid() TestCase {
	return TestCase{
		Name: "fuzz #1", IsFuzz: true, N: 4, Grid: knownGrid,
		ColTop: knownColTop, ColBottom: knownColBottom,
		RowLeft: knownRowLeft, RowRight: knownRowRight,
		Args: []string{"unused"},
	}
}

func TestEvaluateResultFuzzPass(t *testing.T) {
	tc := fuzzCaseFromKnownGrid()
	res := Result{Stdout: "1 2 3 4\n2 1 4 3\n3 4 1 2\n4 3 2 1\n"}
	tr := EvaluateResult(tc, res)
	if !tr.Passed {
		t.Fatalf("expected pass, got reason=%q mismatches=%v", tr.Reason, tr.Mismatches)
	}
}

func TestEvaluateResultFuzzFailWrongGrid(t *testing.T) {
	tc := fuzzCaseFromKnownGrid()
	// A different, still-valid Latin square that does NOT satisfy this case's views.
	res := Result{Stdout: "2 1 4 3\n1 2 3 4\n4 3 2 1\n3 4 1 2\n"}
	tr := EvaluateResult(tc, res)
	if tr.Passed {
		t.Fatalf("expected failure for grid that doesn't match target views")
	}
}

func TestEvaluateResultErrorCasePass(t *testing.T) {
	tc := TestCase{Name: "no arguments", IsFuzz: false, N: 4}
	res := Result{Stdout: "Error\n"}
	tr := EvaluateResult(tc, res)
	if !tr.Passed {
		t.Fatalf("expected pass for 'Error' output, got reason=%q", tr.Reason)
	}
}

func TestEvaluateResultErrorCaseFailWhenGridPrinted(t *testing.T) {
	tc := TestCase{Name: "no arguments", IsFuzz: false, N: 4}
	res := Result{Stdout: "1 2 3 4\n2 1 4 3\n3 4 1 2\n4 3 2 1\n"}
	tr := EvaluateResult(tc, res)
	if tr.Passed {
		t.Fatalf("expected failure when a grid is printed for invalid input")
	}
}

func TestEvaluateResultTimeoutAndExecError(t *testing.T) {
	tc := fuzzCaseFromKnownGrid()

	if tr := EvaluateResult(tc, Result{TimedOut: true}); tr.Passed {
		t.Errorf("expected failure on timeout")
	}
	if tr := EvaluateResult(tc, Result{Err: errTest}); tr.Passed {
		t.Errorf("expected failure on exec error")
	}
}

var errTest = &testExecError{}

type testExecError struct{}

func (*testExecError) Error() string { return "boom" }

func TestBuildTestQueueShape(t *testing.T) {
	queue := BuildTestQueue(4, 5, OrderClockwise)
	if len(queue) != 5+11 {
		t.Fatalf("queue length = %d, want %d", len(queue), 5+11)
	}
	for i := 0; i < 5; i++ {
		if !queue[i].IsFuzz {
			t.Errorf("case %d should be a fuzz case", i)
		}
	}
	for i := 5; i < len(queue); i++ {
		if queue[i].IsFuzz {
			t.Errorf("case %d should be an error case", i)
		}
	}
}
