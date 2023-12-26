package cmd

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"log/slog"
	"os/exec"
	"sync"
)

type Runner struct {
	buf    bytes.Buffer
	mutex  sync.Mutex
	logger *slog.Logger
}

func New(logger *slog.Logger) *Runner {
	if logger == nil {
		logger = &slog.Logger{}
	}
	logger = logger.With("process", "kwok")
	return &Runner{logger: logger}
}

// Run executes the given command and logs its output in real-time.
func (r *Runner) Run(ctx context.Context, command string, args ...string) (io.Reader, error) {
	// Create the command with the provided context
	cmd := exec.CommandContext(ctx, command, args...)

	// Get the stdout and stderr pipes
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	// Start the command
	if err = cmd.Start(); err != nil {
		return nil, err
	}

	// Read stdout and stderr concurrently
	go r.printOutput(stdoutPipe, "STDOUT")
	go r.printOutput(stderrPipe, "STDERR")

	// Wait for the command to finish
	if err = cmd.Wait(); err != nil {
		r.logger.Error("failed to wait for command", "error", err)
		return nil, err
	}

	return &r.buf, nil
}

// printOutput reads from a reader and logs the output using slog.
func (r *Runner) printOutput(rc io.ReadCloser, label string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	logger := slog.With("process", "kwok")
	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		r.buf.WriteString(scanner.Text())
		if label == "STDERR" {
			logger.Error(scanner.Text())
		} else {
			logger.Info(scanner.Text())
		}
	}
}
