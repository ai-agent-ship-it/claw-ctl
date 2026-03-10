package cmd

import (
	"fmt"
	"os"

	"github.com/ai-agent-ship-it/claw-ctl/pkg/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	deployPreset  string
	deployEnvFile string
	deployAgents  string
	deployModel   string
)

var deployCmd = &cobra.Command{
	Use:   "deploy [cluster-name]",
	Short: "Deploy a vCluster with PicoClaw agents",
	Long: `Deploy a new vCluster and provision PicoClaw agents inside it.

Can be used with:
  - A preset: claw-ctl deploy finance --preset financial-controller
  - A config file: claw-ctl deploy --config picoclaw.yaml
  - Flags: claw-ctl deploy finance --agents agent-a,agent-b --model ollama/llama3.1:8b`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var cfg config.ClusterConfig

		// Load from config file
		if cfgFile != "" {
			data, err := os.ReadFile(cfgFile)
			if err != nil {
				return fmt.Errorf("failed to read config file: %w", err)
			}
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return fmt.Errorf("failed to parse config file: %w", err)
			}
			fmt.Printf("  📄 Loaded config from %s\n", cfgFile)
		} else if deployPreset != "" {
			// Load from preset
			preset, ok := config.Presets[deployPreset]
			if !ok {
				return fmt.Errorf("unknown preset: %s (run 'claw-ctl presets' to list)", deployPreset)
			}
			cfg = preset
			if len(args) > 0 {
				cfg.Cluster = args[0]
			}
			fmt.Printf("  🎭 Using preset: %s\n", deployPreset)
		} else if len(args) > 0 {
			// Build from flags
			cfg.Cluster = args[0]
			// TODO: parse --agents and --model flags into AgentConfig
		} else {
			return fmt.Errorf("provide a cluster name, --preset, or --config")
		}

		if cfg.Cluster == "" {
			return fmt.Errorf("cluster name is required")
		}

		// Set secret mode
		if vaultAddr != "" {
			cfg.Secrets.Mode = "vault"
			cfg.Secrets.VaultAddr = vaultAddr
		} else if deployEnvFile != "" {
			cfg.Secrets.Mode = "env"
			cfg.Secrets.EnvFile = deployEnvFile
		}

		// Secret gate: check required tokens
		required := cfg.AllRequiredSecrets()
		if len(required) > 0 {
			fmt.Println()
			fmt.Println("  🚧 Required credentials for this configuration:")
			for _, s := range required {
				fmt.Printf("     • %s\n", s)
			}
			fmt.Println()
			fmt.Println("  [TODO] Secret gate will interactively collect missing tokens")
		}

		// TODO: Implement deployment phases
		fmt.Println()
		fmt.Printf("  🚀 Deploying vCluster '%s' with %d agent(s)...\n", cfg.Cluster, len(cfg.Agents))
		for _, agent := range cfg.Agents {
			fmt.Printf("     • %s (%s)\n", agent.Name, agent.Model)
		}
		fmt.Println()
		fmt.Println("  [TODO] Phase 1: Create namespace + vCluster")
		fmt.Println("  [TODO] Phase 2: Provision secrets")
		fmt.Println("  [TODO] Phase 3: Generate and apply manifests")
		fmt.Println("  [TODO] Phase 4: Health check")

		return nil
	},
}

func init() {
	deployCmd.Flags().StringVar(&deployPreset, "preset", "", "Use a built-in preset configuration")
	deployCmd.Flags().StringVar(&deployEnvFile, "env-file", "", "Path to .env file for secrets")
	deployCmd.Flags().StringVar(&deployAgents, "agents", "", "Comma-separated list of agent names")
	deployCmd.Flags().StringVar(&deployModel, "model", "", "Default LLM model for agents")
	rootCmd.AddCommand(deployCmd)
}
