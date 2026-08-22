package main

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestRunBinaryCapturesStdoutAndExitCode(t *testing.T) {
	res := RunBinary(context.Background(), "/bin/sh", []string{"-c", "echo hello"}, time.Second)
	if res.Err != nil {
		t.Fatalf("unexpected exec error: %v", res.Err)
	}
	if res.TimedOut {
		t.Fatalf("unexpected timeout")
	}
	if strings.TrimSpace(res.Stdout) != "hello" {
		t.Errorf("stdout = %q, want %q", res.Stdout, "hello\n")
	}
	if res.ExitCode != 0 {
		t.Errorf("exit code = %d, want 0", res.ExitCode)
	}
}

func TestRunBinaryReportsNonZeroExit(t *testing.T) {
	res := RunBinary(context.Background(), "/bin/sh", []string{"-c", "exit 3"}, time.Second)
	if res.Err != nil {
		t.Fatalf("unexpected exec error: %v", res.Err)
	}
	if res.TimedOut {
		t.Fatalf("unexpected timeout")
	}
	if res.ExitCode != 3 {
		t.Errorf("exit code = %d, want 3", res.ExitCode)
	}
}

func TestRunBinaryTimesOut(t *testing.T) {
	res := RunBinary(context.Background(), "/bin/sh", []string{"-c", "sleep 5"}, 200*time.Millisecond)
	if !res.TimedOut {
		t.Fatalf("expected TimedOut = true")
	}
}

func TestRunBinaryMissingExecutable(t *testing.T) {
	res := RunBinary(context.Background(), "/no/such/binary-rush01-tester", nil, time.Second)
	if res.Err == nil {
		t.Fatalf("expected exec error for missing binary")
	}
}

func TestRunBinaryStopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	res := RunBinary(ctx, "/bin/sh", []string{"-c", "sleep 5"}, time.Second)
	if res.Err == nil && !res.TimedOut {
		t.Fatalf("expected an error or timeout for an already-canceled context")
	}
}
