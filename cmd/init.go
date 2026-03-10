package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactive wizard to configure and deploy agents",
	Long: `Launch the interactive setup wizard that guides you through:
  1. Selecting a preset or custom configuration
  2. Naming your cluster
  3. Configuring agents (model, channels, tokens, temperature)
  4. Collecting required secrets
  5. Reviewing and deploying`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println()
		fmt.Println("  🐾 PicoClaw Setup Wizard")
		fmt.Println("  ========================")
		fmt.Println()
		fmt.Println("  [TODO] Interactive wizard will be implemented with Bubbletea TUI")
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
