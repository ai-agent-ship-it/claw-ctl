package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var reloadAll bool

var reloadCmd = &cobra.Command{
	Use:   "reload [cluster-name] [agent-name]",
	Short: "Hot-reload agent workspace files (SOUL, skills, etc.) without restart",
	Long: `Push updated workspace files to a running agent by updating the
ConfigMaps in-place. PicoClaw watches for file changes and reloads
its context automatically.

Examples:
  claw-ctl reload finance agent-financiero   # Reload one agent
  claw-ctl reload finance --all               # Reload all agents`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		clusterName := args[0]

		if reloadAll {
			fmt.Printf("\n  🔄 Reloading ALL agents in cluster '%s'...\n", clusterName)
		} else if len(args) == 2 {
			agentName := args[1]
			fmt.Printf("\n  🔄 Reloading agent '%s' in cluster '%s'...\n", agentName, clusterName)
		} else {
			return fmt.Errorf("specify an agent name or use --all")
		}

		fmt.Println("  [TODO] Read workspace files from disk")
		fmt.Println("  [TODO] Update ConfigMaps in vCluster")
		fmt.Println("  [TODO] Verify agent picked up changes")

		return nil
	},
}

func init() {
	reloadCmd.Flags().BoolVar(&reloadAll, "all", false, "Reload all agents in the cluster")
	rootCmd.AddCommand(reloadCmd)
}
