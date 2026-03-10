package cmd

import (
	"fmt"
	"os"

	"github.com/ai-agent-ship-it/claw-ctl/pkg/config"
	"github.com/ai-agent-ship-it/claw-ctl/pkg/wizard"
	"github.com/ai-agent-ship-it/claw-ctl/pkg/workspace"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := wizard.RunWizard()
		if err != nil {
			return fmt.Errorf("wizard failed: %w", err)
		}

		if result.Action == "cancel" {
			fmt.Println("\n  ❌ Wizard cancelled.")
			return nil
		}

		// Generate workspace files for each agent
		for _, agent := range result.Config.Agents {
			if err := workspace.GenerateWorkspace("workspace", agent.Name, agent.Model, agent.Capabilities); err != nil {
				return fmt.Errorf("failed to generate workspace for %s: %w", agent.Name, err)
			}
			fmt.Printf("  📂 Workspace generated: workspace/%s/\n", agent.Name)

			// Set workspace paths in config
			agent.Workspace = config.WorkspaceConfig{
				Soul:        fmt.Sprintf("workspace/%s/SOUL.md", agent.Name),
				Identity:    fmt.Sprintf("workspace/%s/IDENTITY.md", agent.Name),
				User:        fmt.Sprintf("workspace/%s/USER.md", agent.Name),
				Agent:       fmt.Sprintf("workspace/%s/AGENT.md", agent.Name),
				Environment: fmt.Sprintf("workspace/%s/ENVIRONMENT.md", agent.Name),
				Memory:      fmt.Sprintf("workspace/%s/memory/", agent.Name),
			}
		}

		if result.Action == "save" {
			data, err := yaml.Marshal(result.Config)
			if err != nil {
				return fmt.Errorf("failed to marshal config: %w", err)
			}
			if err := os.WriteFile("picoclaw.yaml", data, 0644); err != nil {
				return fmt.Errorf("failed to write config: %w", err)
			}
			fmt.Println("\n  💾 Config saved to picoclaw.yaml")
			fmt.Println("  📝 Edit workspace files to customize your agent's personality.")
			fmt.Println("  🚀 Run: claw-ctl deploy --config picoclaw.yaml")
			return nil
		}

		// Deploy
		fmt.Println("\n  🚀 Starting deployment...")
		return runDeploy(result.Config, result.SecretValues)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
