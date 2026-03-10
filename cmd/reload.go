package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ai-agent-ship-it/claw-ctl/pkg/vcluster"
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
		ctx := context.Background()
		kubeconfig := vcluster.KubeconfigPath(clusterName)

		if reloadAll {
			// Discover agents from workspace directory
			entries, err := os.ReadDir("workspace")
			if err != nil {
				return fmt.Errorf("cannot read workspace/ directory: %w", err)
			}
			for _, e := range entries {
				if e.IsDir() {
					if err := reloadAgent(ctx, kubeconfig, e.Name()); err != nil {
						fmt.Printf("  ⚠️  Failed to reload %s: %v\n", e.Name(), err)
					}
				}
			}
		} else if len(args) == 2 {
			return reloadAgent(ctx, kubeconfig, args[1])
		} else {
			return fmt.Errorf("specify an agent name or use --all")
		}

		return nil
	},
}

func reloadAgent(ctx context.Context, kubeconfig, agentName string) error {
	fmt.Printf("\n  🔄 Reloading agent '%s'...\n", agentName)

	wsDir := filepath.Join("workspace", agentName)
	if _, err := os.Stat(wsDir); os.IsNotExist(err) {
		return fmt.Errorf("workspace directory not found: %s", wsDir)
	}

	// Read workspace files and create/update ConfigMap
	files := []string{"SOUL.md", "IDENTITY.md", "USER.md", "AGENT.md", "ENVIRONMENT.md"}
	configMapArgs := []string{"create", "configmap", agentName + "-workspace",
		"--namespace", "agents",
		"--dry-run=client", "-o", "yaml",
	}

	for _, f := range files {
		path := filepath.Join(wsDir, f)
		if _, err := os.Stat(path); err == nil {
			configMapArgs = append(configMapArgs, fmt.Sprintf("--from-file=%s=%s", f, path))
		}
	}

	// Generate ConfigMap YAML
	genCmd := exec.CommandContext(ctx, "kubectl", configMapArgs...)
	cmYAML, err := genCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to generate configmap: %w", err)
	}

	// Apply it
	applyArgs := []string{"apply", "-f", "-"}
	if kubeconfig != "" {
		applyArgs = append([]string{"--kubeconfig", kubeconfig}, applyArgs...)
	}
	applyCmd := exec.CommandContext(ctx, "kubectl", applyArgs...)
	applyCmd.Stdin = strings.NewReader(string(cmYAML))
	if output, err := applyCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to apply configmap: %s: %w", string(output), err)
	}

	fmt.Printf("  ✅ Agent '%s' workspace files reloaded\n", agentName)
	return nil
}

func init() {
	reloadCmd.Flags().BoolVar(&reloadAll, "all", false, "Reload all agents in the cluster")
	rootCmd.AddCommand(reloadCmd)
}
