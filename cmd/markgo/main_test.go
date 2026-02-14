package main

import (
	"os/exec"
	"strings"
	"testing"
)

// runMarkgo runs the CLI via `go run` and returns stdout, stderr, and exit code.
func runMarkgo(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmdArgs := append([]string{"run", "."}, args...)
	cmd := exec.Command("go", cmdArgs...)
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run markgo: %v", err)
		}
	}

	return outBuf.String(), errBuf.String(), exitCode
}

func TestHelpOutput(t *testing.T) {
	for _, arg := range []string{"help", "--help", "-h"} {
		t.Run(arg, func(t *testing.T) {
			stdout, _, exitCode := runMarkgo(t, arg)
			if exitCode != 0 {
				t.Errorf("markgo %s exited with code %d, want 0", arg, exitCode)
			}
			for _, want := range []string{"USAGE:", "COMMANDS:", "ALIASES:", "EXAMPLES:", "serve", "init", "new"} {
				if !strings.Contains(stdout, want) {
					t.Errorf("markgo %s output missing %q", arg, want)
				}
			}
		})
	}
}

func TestVersionOutput(t *testing.T) {
	for _, arg := range []string{"version", "--version", "-v"} {
		t.Run(arg, func(t *testing.T) {
			stdout, _, exitCode := runMarkgo(t, arg)
			if exitCode != 0 {
				t.Errorf("markgo %s exited with code %d, want 0", arg, exitCode)
			}
			if !strings.Contains(stdout, "MarkGo ") {
				t.Errorf("markgo %s output missing version string, got: %q", arg, stdout)
			}
		})
	}
}

func TestUnknownCommand(t *testing.T) {
	_, stderr, exitCode := runMarkgo(t, "nonsense")
	if exitCode != 1 {
		t.Errorf("markgo nonsense exited with code %d, want 1", exitCode)
	}
	if !strings.Contains(stderr, "Unknown command: nonsense") {
		t.Errorf("markgo nonsense stderr missing error message, got: %q", stderr)
	}
}

func TestServeHelp(t *testing.T) {
	stdout, _, exitCode := runMarkgo(t, "serve", "--help")
	if exitCode != 0 {
		t.Errorf("markgo serve --help exited with code %d, want 0", exitCode)
	}
	for _, want := range []string{"markgo serve", "--port", "CONFIGURATION:", ".env"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("markgo serve --help output missing %q", want)
		}
	}
}

func TestInitHelp(t *testing.T) {
	stdout, _, exitCode := runMarkgo(t, "init", "--help")
	if exitCode != 0 {
		t.Errorf("markgo init --help exited with code %d, want 0", exitCode)
	}
	if !strings.Contains(stdout, "--quick") {
		t.Errorf("markgo init --help output missing --quick flag")
	}
}

func TestNewHelp(t *testing.T) {
	stdout, _, exitCode := runMarkgo(t, "new", "--help")
	if exitCode != 0 {
		t.Errorf("markgo new --help exited with code %d, want 0", exitCode)
	}
	for _, want := range []string{"--title", "--template", "AVAILABLE TEMPLATES:"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("markgo new --help output missing %q", want)
		}
	}
}
