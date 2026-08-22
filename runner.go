package main

import (
	"bytes"
	"context"
	"os/exec"
	"syscall"
	"time"
)

// Result is the outcome of running the tested binary once.
type Result struct {
	Stdout   string
	ExitCode int
	TimedOut bool
	Err      error // non-nil only for exec-level failures (binary not found, permission denied, etc.)
}

// RunBinary executes path with args, capturing stdout and killing the
// process group if it runs longer than timeout or if ctx is canceled first
// (e.g. the caller aborting a run in progress). Killing the whole process
// group (not just the direct child) catches runaway children the tested
// binary might spawn.
func RunBinary(ctx context.Context, path string, args []string, timeout time.Duration) Result {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()
	result := Result{Stdout: stdout.String()}

	if ctx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		return result
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			return result
		}
		result.Err = err
		return result
	}

	return result
}
