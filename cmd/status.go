package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [cluster-name]",
	Short: "Show cluster and agent status",
	Long:  "Display health information for a vCluster, its agents, secret sync state, and model info.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		clusterName := args[0]
		namespace := "vcluster-" + clusterName

		fmt.Printf("\n  📊 Status: %s (ns: %s)\n", clusterName, namespace)
		fmt.Println("  ──────────────────────────────")
		fmt.Println("  [TODO] vCluster status")
		fmt.Println("  [TODO] Pod status per agent")
		fmt.Println("  [TODO] Secret sync state")
		fmt.Println("  [TODO] Model and channel info")

		return nil
	},
}

var addAgentCmd = &cobra.Command{
	Use:   "add-agent [cluster-name] [agent-name]",
	Short: "Add a new agent to an existing vCluster",
	Long:  "Hot-add a PicoClaw agent to a running vCluster with its own config, secrets, and workspace.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		clusterName := args[0]
		agentName := args[1]

		fmt.Printf("\n  ➕ Adding agent '%s' to cluster '%s'...\n", agentName, clusterName)
		fmt.Println("  [TODO] Agent configuration wizard")
		fmt.Println("  [TODO] Secret gate")
		fmt.Println("  [TODO] Generate manifests for new agent")
		fmt.Println("  [TODO] Patch vCluster secret sync")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(addAgentCmd)
}
