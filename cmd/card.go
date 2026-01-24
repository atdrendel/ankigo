package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	cardFront string
	cardBack  string
	cardDeck  string
)

var cardCmd = &cobra.Command{
	Use:   "card",
	Short: "Manage Anki cards",
	Long:  `Commands for adding, searching, and managing Anki cards.`,
}

var cardAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new card",
	Long:  `Add a new flashcard to a deck with front and back content.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintf(cmd.OutOrStdout(), "card add: not yet implemented\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  deck:  %s\n", cardDeck)
		fmt.Fprintf(cmd.OutOrStdout(), "  front: %s\n", cardFront)
		fmt.Fprintf(cmd.OutOrStdout(), "  back:  %s\n", cardBack)
		return nil
	},
}

var cardSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for cards",
	Long:  `Search for cards in your Anki collection using a query string.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		fmt.Fprintf(cmd.OutOrStdout(), "card search %q: not yet implemented\n", query)
		return nil
	},
}

func init() {
	cardAddCmd.Flags().StringVarP(&cardFront, "front", "f", "", "front of the card (required)")
	cardAddCmd.Flags().StringVarP(&cardBack, "back", "b", "", "back of the card (required)")
	cardAddCmd.Flags().StringVarP(&cardDeck, "deck", "d", "Default", "deck to add the card to")

	cardAddCmd.MarkFlagRequired("front")
	cardAddCmd.MarkFlagRequired("back")

	cardCmd.AddCommand(cardAddCmd)
	cardCmd.AddCommand(cardSearchCmd)
	rootCmd.AddCommand(cardCmd)
}
