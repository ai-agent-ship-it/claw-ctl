package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	vaultAddr  string
	presetName string
)

var rootCmd = &cobra.Command{
	Use:   "claw-ctl",
	Short: "🐾 PicoClaw Agent Deployment CLI",
	Long: `claw-ctl is a CLI tool for deploying and managing PicoClaw AI agents
on Kubernetes using vClusters. It provides an interactive wizard, preset
configurations, and handles vCluster lifecycle, secret management, and
agent workspace files.

The only requirement is an LLM provider token or a local Ollama address.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: picoclaw.yaml)")
	rootCmd.PersistentFlags().StringVar(&vaultAddr, "vault-addr", "", "Vault server address (env: VAULT_ADDR)")

	// Env var fallbacks
	if vaultAddr == "" {
		vaultAddr = os.Getenv("VAULT_ADDR")
	}
}
