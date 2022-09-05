package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(completionCmd)
}

// See: https://github.com/spf13/cobra/blob/main/shell_completions.md
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script for your shell",
	Long: `To load completions:
Bash:
  $ source <(rill completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ rill completion bash > /etc/bash_completion.d/rill
  # macOS:
  $ rill completion bash > /usr/local/etc/bash_completion.d/rill
Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  # To load completions for each session, execute once:
  $ rill completion zsh > "${fpath[1]}/_rill"
  # You will need to start a new shell for this setup to take effect.
fish:
  $ rill completion fish | source
  # To load completions for each session, execute once:
  $ rill completion fish > ~/.config/fish/completions/rill.fish
PowerShell:
  PS> rill completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> rill completion powershell > rill.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}
