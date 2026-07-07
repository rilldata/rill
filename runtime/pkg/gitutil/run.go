// Package gitutil provides utilities for working with git repositories.
// All git operations shell out to the git CLI via Run; the go-git library must not be used here.
// Credential-embedded remote URLs may be passed as command-line arguments, but must never be
// persisted to .git/config; Run redacts URL credentials in error messages.
package gitutil

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("github.com/rilldata/rill/runtime/pkg/gitutil")

// Run executes a git command with the specified arguments in the given path and returns its output or an error.
// If path is empty, the command runs without -C (use for commands like `clone` that take an explicit destination).
// Use it to run one-off git commands that don't fit into the other helper functions in this package.
func Run(ctx context.Context, path string, args ...string) (string, error) {
	// Only record the subcommand (e.g. "clone", "push"): later args may contain credential-embedded remote URLs.
	var subcommand string
	if len(args) > 0 {
		subcommand = args[0]
	}
	ctx, span := tracer.Start(ctx, "gitutil.Run", oteltrace.WithAttributes(attribute.String("git.command", subcommand)))
	defer span.End()

	fullArgs := args
	if path != "" {
		fullArgs = append([]string{"-C", path}, args...)
	}
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", fullArgs...)
	// Force English error messages so stderr substring checks are stable, and disable interactive credential prompts.
	cmd.Env = append(os.Environ(), "LC_ALL=C", "GIT_TERMINAL_PROMPT=0")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			err = fmt.Errorf("git executable not found: install git from https://git-scm.com (%w)", err)
			span.SetStatus(codes.Error, "git executable not found")
			return "", err
		}
		// Some commands report failures (e.g. merge conflicts) on stdout rather than stderr, so include both.
		output := strings.TrimSpace(stderr.String())
		if so := strings.TrimSpace(stdout.String()); so != "" {
			if output != "" {
				output += "\n" + so
			} else {
				output = so
			}
		}
		// Redact credentials: args, stderr, and stdout may contain credential-embedded remote URLs.
		msg := redactURLCredentials(fmt.Sprintf("git %s: %s", strings.Join(args, " "), output))
		err = fmt.Errorf("%s(%w)", msg, err)
		span.SetStatus(codes.Error, msg)
		return "", err
	}
	return strings.TrimSpace(stdout.String()), nil
}

// urlCredentialsRegexp matches the userinfo component of a URL (e.g. "https://user:token@host").
var urlCredentialsRegexp = regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9+.-]*://)[^@/\s]+@`)

// redactURLCredentials masks credentials embedded in URLs in s.
func redactURLCredentials(s string) string {
	return urlCredentialsRegexp.ReplaceAllString(s, "$1<redacted>@")
}
