package cmd

import (
	"fmt"

	"github.com/ai-agent-ship-it/claw-ctl/pkg/config"
	"github.com/spf13/cobra"
)

var presetsCmd = &cobra.Command{
	Use:   "presets",
	Short: "List available configuration presets",
	Long:  "Display all built-in presets with their agent configurations, models, and channels.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println()
		fmt.Println("  🐾 Available Presets")
		fmt.Println("  ====================")
		fmt.Println()
		for _, name := range config.PresetNames() {
			desc := config.PresetDescriptions[name]
			fmt.Printf("  %-24s %s\n", name, desc)
		}
		fmt.Println()
		fmt.Println("  Usage: picoclaw-ctl deploy <cluster-name> --preset <preset-name>")
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(presetsCmd)
}
