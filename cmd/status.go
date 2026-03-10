package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/ai-agent-ship-it/claw-ctl/pkg/vcluster"
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
		ctx := context.Background()

		fmt.Printf("\n  📊 Status: %s\n", clusterName)
		fmt.Println("  ──────────────────────────────")

		// Check vCluster pod
		fmt.Printf("\n  ☸️  vCluster (ns: %s):\n", namespace)
		vcPods := exec.CommandContext(ctx, "kubectl", "get", "pods",
			"-n", namespace,
			"-l", fmt.Sprintf("app=vcluster,release=%s", clusterName),
			"--no-headers",
		)
		output, err := vcPods.Output()
		if err != nil {
			fmt.Println("     ❌ vCluster not found or not accessible")
		} else {
			for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
				if line != "" {
					fmt.Printf("     %s\n", line)
				}
			}
		}

		// Check agent pods inside vCluster
		kubeconfig := vcluster.KubeconfigPath(clusterName)
		fmt.Println("\n  🤖 Agents (inside vCluster):")
		agentPods := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
			"get", "pods", "-n", "agents", "--no-headers",
		)
		agentOutput, err := agentPods.Output()
		if err != nil {
			fmt.Println("     ⚠️  Cannot reach vCluster (run 'claw-ctl deploy' or connect first)")
		} else {
			lines := strings.Split(strings.TrimSpace(string(agentOutput)), "\n")
			if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
				fmt.Println("     No agents deployed")
			} else {
				for _, line := range lines {
					if line != "" {
						fmt.Printf("     %s\n", line)
					}
				}
			}
		}

		// Check secrets
		fmt.Println("\n  🔐 Secrets (host namespace):")
		secretsCmd := exec.CommandContext(ctx, "kubectl", "get", "secrets",
			"-n", namespace,
			"-l", "managed-by=claw-ctl",
			"--no-headers",
		)
		secretOutput, err := secretsCmd.Output()
		if err != nil {
			fmt.Println("     No managed secrets found")
		} else {
			lines := strings.Split(strings.TrimSpace(string(secretOutput)), "\n")
			if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
				fmt.Println("     No managed secrets found")
			} else {
				for _, line := range lines {
					if line != "" {
						fmt.Printf("     %s\n", line)
					}
				}
			}
		}

		fmt.Println()
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

		fmt.Printf("\n  ➕ Adding agent '%s' to cluster '%s'...\n\n", agentName, clusterName)
		fmt.Println("  [TODO] Full wizard for agent configuration")
		fmt.Println("  [TODO] Secret gate for new agent's tokens")
		fmt.Println("  [TODO] Generate + apply manifests")
		fmt.Println("  [TODO] Patch vCluster secret sync mappings")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(addAgentCmd)
}
