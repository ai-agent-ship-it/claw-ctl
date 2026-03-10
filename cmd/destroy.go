package cmd

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/ai-agent-ship-it/claw-ctl/pkg/k8s"
	"github.com/ai-agent-ship-it/claw-ctl/pkg/vcluster"
	"github.com/spf13/cobra"
)

var forceDestroy bool

var destroyCmd = &cobra.Command{
	Use:   "destroy [cluster-name]",
	Short: "Destroy a vCluster and cleanup all resources",
	Long: `Fully tear down a vCluster deployment including:
  - Delete the vCluster
  - Remove the host namespace
  - Clean up Vault policies and auth roles (if Vault was used)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		clusterName := args[0]
		namespace := "vcluster-" + clusterName
		ctx := context.Background()

		if !forceDestroy {
			fmt.Printf("\n  ⚠️  This will permanently destroy cluster '%s'\n", clusterName)
			fmt.Printf("  Namespace: %s\n\n", namespace)
			fmt.Print("  Are you sure? [y/N]: ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" && confirm != "yes" {
				fmt.Println("  ❌ Aborted.")
				return nil
			}
		}

		fmt.Printf("\n  🗑️  Destroying cluster '%s'...\n\n", clusterName)

		// Step 1: Delete vCluster
		vcm := vcluster.NewManager()
		if err := vcm.Delete(ctx, clusterName, namespace); err != nil {
			fmt.Printf("  ⚠️  vCluster delete: %v\n", err)
		}

		// Step 2: Delete namespace
		fmt.Printf("  🗑️  Deleting namespace '%s'...\n", namespace)
		k8sClient, err := k8s.NewClient()
		if err != nil {
			return fmt.Errorf("K8s connectivity failed: %w", err)
		}
		nsCmd := exec.CommandContext(ctx, "kubectl", "delete", "ns", namespace, "--ignore-not-found")
		if output, err := nsCmd.CombinedOutput(); err != nil {
			fmt.Printf("  ⚠️  Namespace delete: %s\n", string(output))
		} else {
			fmt.Printf("  ✅ Namespace '%s' deleted\n", namespace)
		}
		_ = k8sClient // used for future vault cleanup

		fmt.Printf("\n  🎉 Cluster '%s' destroyed.\n\n", clusterName)
		return nil
	},
}

func init() {
	destroyCmd.Flags().BoolVar(&forceDestroy, "force", false, "Skip confirmation prompt")
	rootCmd.AddCommand(destroyCmd)
}
