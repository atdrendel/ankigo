package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// isStdinTerminal reports whether stdin is a terminal (not piped/redirected).
func isStdinTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// requireConfirmation prompts the user for y/N confirmation on stderr.
// If isTerminal returns false, it returns an error telling the user to use --force.
// Returns nil if confirmed, ErrCancelled if declined, or an error if non-interactive.
func requireConfirmation(stdin io.Reader, stderr io.Writer, isTerminal func() bool) error {
	if !isTerminal() {
		return fmt.Errorf("confirmation required: use --force to skip prompts in non-interactive mode")
	}
	fmt.Fprint(stderr, "Continue? [y/N] ")
	reader := bufio.NewReader(stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return ErrCancelled
	}
	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		return ErrCancelled
	}
	return nil
}
