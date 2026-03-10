package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

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

		fmt.Printf("\n  ⚠️  This will permanently destroy cluster '%s'\n", clusterName)
		fmt.Printf("  Namespace: %s\n\n", namespace)
		fmt.Println("  [TODO] Confirmation prompt")
		fmt.Println("  [TODO] vcluster delete")
		fmt.Println("  [TODO] kubectl delete ns")
		fmt.Println("  [TODO] Vault cleanup (if applicable)")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
