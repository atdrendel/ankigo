package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestRequireConfirmation_NonInteractiveReturnsError(t *testing.T) {
	var stderr bytes.Buffer
	stdin := &bytes.Buffer{}

	err := requireConfirmation(stdin, &stderr, func() bool { return false })

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "--force") {
		t.Errorf("expected error mentioning --force, got: %v", err)
	}
}

func TestRequireConfirmation_InteractiveYes(t *testing.T) {
	var stderr bytes.Buffer
	stdin := bytes.NewBufferString("y\n")

	err := requireConfirmation(stdin, &stderr, func() bool { return true })

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr.String(), "Continue?") {
		t.Errorf("expected prompt on stderr, got %q", stderr.String())
	}
}

func TestRequireConfirmation_InteractiveYesFull(t *testing.T) {
	var stderr bytes.Buffer
	stdin := bytes.NewBufferString("yes\n")

	err := requireConfirmation(stdin, &stderr, func() bool { return true })

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRequireConfirmation_InteractiveNo(t *testing.T) {
	var stderr bytes.Buffer
	stdin := bytes.NewBufferString("n\n")

	err := requireConfirmation(stdin, &stderr, func() bool { return true })

	if !errors.Is(err, ErrCancelled) {
		t.Errorf("expected ErrCancelled, got: %v", err)
	}
}

func TestRequireConfirmation_InteractiveEmpty(t *testing.T) {
	var stderr bytes.Buffer
	stdin := bytes.NewBufferString("\n")

	err := requireConfirmation(stdin, &stderr, func() bool { return true })

	if !errors.Is(err, ErrCancelled) {
		t.Errorf("expected ErrCancelled, got: %v", err)
	}
}

func TestRequireConfirmation_InteractiveEOF(t *testing.T) {
	var stderr bytes.Buffer
	stdin := &bytes.Buffer{} // empty = EOF

	err := requireConfirmation(stdin, &stderr, func() bool { return true })

	if !errors.Is(err, ErrCancelled) {
		t.Errorf("expected ErrCancelled, got: %v", err)
	}
}
