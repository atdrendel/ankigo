package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var verbose bool

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "ankigo",
	Short: "A CLI for interacting with Anki via anki-connect",
	Long: `ankigo is a command-line interface for managing Anki flashcards.

It communicates with Anki through the anki-connect plugin, allowing you to
create decks, add cards, search your collection, and more—all from the terminal.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

// NewRootCmd returns the root command for testing purposes.
func NewRootCmd() *cobra.Command {
	return rootCmd
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
}

// Verbose returns whether verbose mode is enabled.
func Verbose() bool {
	return verbose
}

func init() {
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
}
