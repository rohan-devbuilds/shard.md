package tools

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"shard/internal/permissions"
)

type Shell struct {
	Permissions *permissions.Manager
	In          io.Reader
	Out         io.Writer
	Timeout     time.Duration
}

type CommandResult struct {
	Command  string
	Stdout   string
	Stderr   string
	ExitCode int
	TimedOut bool
}

func (s *Shell) Run(command string) (CommandResult, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return CommandResult{}, fmt.Errorf("command is required")
	}
	if IsDangerousCommand(command) {
		if err := confirmDangerous(command, s.In, s.Out); err != nil {
			return CommandResult{}, err
		}
	}

	ok, err := s.Permissions.Approve(permissions.RunCommand, command, s.In, s.Out)
	if err != nil {
		return CommandResult{}, err
	}
	if !ok {
		return CommandResult{}, fmt.Errorf("permission denied for command: %s", command)
	}

	timeout := s.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := shellCommand(ctx, command)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = strings.NewReader("")

	err = cmd.Run()
	result := CommandResult{
		Command:  command,
		Stdout:   Truncate(stdout.String(), MaxOutputBytes),
		Stderr:   Truncate(stderr.String(), MaxOutputBytes),
		ExitCode: 0,
		TimedOut: ctx.Err() == context.DeadlineExceeded,
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
		return result, nil
	}
	if err != nil {
		return result, err
	}
	return result, nil
}

func FormatCommandResult(result CommandResult) string {
	timeout := ""
	if result.TimedOut {
		timeout = "\ntimed_out: true"
	}
	return fmt.Sprintf("Tool result:\ncommand: %s\nexit_code: %d%s\nstdout:\n%s\nstderr:\n%s", result.Command, result.ExitCode, timeout, result.Stdout, result.Stderr)
}

func IsDangerousCommand(command string) bool {
	lower := strings.ToLower(command)
	dangerous := []string{
		"rm -rf",
		"del /s",
		"format",
		"shutdown",
		"reboot",
		"diskpart",
		"sudo",
		"remove-item",
		" rmdir ",
	}
	for _, pattern := range dangerous {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

func confirmDangerous(command string, in io.Reader, out io.Writer) error {
	fmt.Fprintf(out, "\nThis command may be dangerous:\n%s\n\nType RUN to confirm: ", command)
	var answer string
	if _, err := fmt.Fscanln(in, &answer); err != nil {
		return err
	}
	if answer != "RUN" {
		return fmt.Errorf("dangerous command not confirmed")
	}
	return nil
}

func shellCommand(ctx context.Context, command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", command)
	}
	return exec.CommandContext(ctx, "sh", "-c", command)
}
