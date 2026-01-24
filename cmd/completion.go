package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for ankigo.

To load completions:

Bash:
  $ source <(ankigo completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ ankigo completion bash > /etc/bash_completion.d/ankigo
  # macOS:
  $ ankigo completion bash > $(brew --prefix)/etc/bash_completion.d/ankigo

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ ankigo completion zsh > "${fpath[1]}/_ankigo"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ ankigo completion fish | source

  # To load completions for each session, execute once:
  $ ankigo completion fish > ~/.config/fish/completions/ankigo.fish

PowerShell:
  PS> ankigo completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> ankigo completion powershell > ankigo.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
