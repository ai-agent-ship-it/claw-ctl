package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ai-agent-ship-it/claw-ctl/pkg/config"
	"github.com/ai-agent-ship-it/claw-ctl/pkg/k8s"
	"github.com/ai-agent-ship-it/claw-ctl/pkg/manifests"
	"github.com/ai-agent-ship-it/claw-ctl/pkg/secrets"
	"github.com/ai-agent-ship-it/claw-ctl/pkg/vcluster"
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
			if deployAgents != "" {
				for _, name := range strings.Split(deployAgents, ",") {
					name = strings.TrimSpace(name)
					if name != "" {
						model := deployModel
						if model == "" {
							model = "ollama/llama3.1:8b"
						}
						cfg.Agents = append(cfg.Agents, config.AgentConfig{
							Name:        name,
							Model:       model,
							MaxTokens:   8192,
							Temperature: 0.2,
							Channels: config.ChannelsConfig{
								HTTP: &config.ChannelHTTP{Enabled: true},
							},
						})
					}
				}
			}
		} else {
			return fmt.Errorf("provide a cluster name, --preset, or --config")
		}

		if cfg.Cluster == "" {
			return fmt.Errorf("cluster name is required")
		}
		if len(cfg.Agents) == 0 {
			return fmt.Errorf("at least one agent is required (use --agents or --preset)")
		}

		// Set secret mode
		if useVault {
			addr, _, err := resolveVaultConfig(deployEnvFile)
			if err != nil {
				return err
			}
			cfg.Secrets.Mode = "vault"
			cfg.Secrets.VaultAddr = addr
		} else if deployEnvFile != "" {
			cfg.Secrets.Mode = "env"
			cfg.Secrets.EnvFile = deployEnvFile
		}

		// Load secrets from .env if applicable
		var secretValues map[string]string
		if cfg.Secrets.Mode == "env" && cfg.Secrets.EnvFile != "" {
			var err error
			secretValues, err = secrets.LoadEnvFile(cfg.Secrets.EnvFile)
			if err != nil {
				return fmt.Errorf("failed to load .env file: %w", err)
			}
		}

		// Secret gate: validate required tokens are present
		required := cfg.AllRequiredSecrets()
		if len(required) > 0 && secretValues != nil {
			missing := secrets.ValidateSecrets(secretValues, required)
			if len(missing) > 0 {
				fmt.Println("\n  ❌ Missing required secrets:")
				for _, s := range missing {
					fmt.Printf("     • %s\n", s)
				}
				return fmt.Errorf("add missing secrets to %s and retry", cfg.Secrets.EnvFile)
			}
		}

		return runDeploy(cfg, secretValues)
	},
}

// runDeploy orchestrates the full deployment.
func runDeploy(cfg config.ClusterConfig, secretValues map[string]string) error {
	ctx := context.Background()
	namespace := "vcluster-" + cfg.Cluster

	fmt.Printf("\n  🚀 Deploying vCluster '%s' with %d agent(s)...\n\n", cfg.Cluster, len(cfg.Agents))

	// Phase 1: K8s connectivity + namespace
	fmt.Println("  ── Phase 1: Infrastructure ──")
	k8sClient, err := k8s.NewClient()
	if err != nil {
		return fmt.Errorf("K8s connectivity failed: %w", err)
	}
	if err := k8sClient.HealthCheck(ctx); err != nil {
		return fmt.Errorf("K8s health check failed: %w", err)
	}
	fmt.Println("  ✅ Kubernetes connected")

	if err := k8sClient.CreateNamespace(ctx, namespace); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
		fmt.Printf("  ℹ️  Namespace '%s' already exists\n", namespace)
	} else {
		fmt.Printf("  ✅ Namespace '%s' created\n", namespace)
	}

	// Phase 2: vCluster
	fmt.Println("\n  ── Phase 2: vCluster ──")
	vcm := vcluster.NewManager()
	// Build agent names for secret sync mapping (only needed for Vault mode)
	var syncAgentNames []string
	if cfg.Secrets.Mode == "vault" {
		for _, a := range cfg.Agents {
			syncAgentNames = append(syncAgentNames, a.Name)
		}
	}
	if err := vcm.Create(ctx, cfg.Cluster, namespace, syncAgentNames); err != nil {
		return err
	}
	if err := vcm.WaitReady(ctx, cfg.Cluster, namespace); err != nil {
		return err
	}

	// Phase 3: Secrets
	fmt.Println("\n  ── Phase 3: Secrets ──")
	if cfg.Secrets.Mode == "vault" && cfg.Secrets.VaultAddr != "" {
		addr, token, err := resolveVaultConfig(cfg.Secrets.EnvFile)
		if err != nil {
			return err
		}
		vp, err := secrets.NewVaultProvisioner(addr, token)
		if err != nil {
			return err
		}
		for _, agent := range cfg.Agents {
			if err := vp.ProvisionAgent(cfg.Cluster, agent); err != nil {
				return err
			}

			// Render and apply VSO CRDs (VaultConnection, VaultAuth, VaultStaticSecret)
			// These are applied to the HOST namespace so VSO creates K8s Secrets
			fmt.Printf("  📜 Applying VSO CRDs for '%s' in host namespace...\n", agent.Name)
			vsoManifest, err := manifests.RenderVaultManifests(cfg.Cluster, namespace, agent)
			if err != nil {
				return fmt.Errorf("failed to render VSO manifests: %w", err)
			}

			tmpFile, err := os.CreateTemp("", fmt.Sprintf("claw-ctl-vso-%s-*.yaml", agent.Name))
			if err != nil {
				return err
			}
			tmpFile.WriteString(vsoManifest)
			tmpFile.Close()

			vsoCmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", tmpFile.Name())
			if output, err := vsoCmd.CombinedOutput(); err != nil {
				fmt.Printf("  ⚠️  VSO CRD apply: %s\n", strings.TrimSpace(string(output)))
			} else {
				fmt.Printf("  ✅ VSO CRDs applied — secrets will sync automatically\n")
			}
			os.Remove(tmpFile.Name())
		}
	} else if secretValues != nil {
		for _, agent := range cfg.Agents {
			required := agent.RequiredSecrets()
			agentSecrets := secrets.FilterSecrets(secretValues, required)
			secretName := agent.Name + "-secret"
			created, err := k8sClient.EnsureSecret(ctx, namespace, secretName, agentSecrets)
			if err != nil {
				return err
			}
			if created {
				fmt.Printf("  ✅ Secret '%s' created (%d keys)\n", secretName, len(agentSecrets))
			} else {
				fmt.Printf("  ℹ️  Secret '%s' updated (%d keys)\n", secretName, len(agentSecrets))
			}
		}
	} else {
		fmt.Println("  ⚠️  No secrets configured (mode: manual)")
	}

	// Phase 4: Apply manifests inside vCluster
	fmt.Println("\n  ── Phase 4: Agent Manifests ──")

	for _, agent := range cfg.Agents {
		fmt.Printf("  📦 Deploying agent '%s' (%s)...\n", agent.Name, agent.Model)
		rendered, err := manifests.RenderAllForAgent(agent)
		if err != nil {
			return fmt.Errorf("failed to render manifests for %s: %w", agent.Name, err)
		}

		// Write all manifests to a single temp file and apply via vcluster connect
		tmpFile, err := os.CreateTemp("", fmt.Sprintf("claw-ctl-%s-*.yaml", agent.Name))
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		tmpPath := tmpFile.Name()

		// Write all manifests separated by ---
		order := []string{"namespace", "rbac", "pvc", "configmap", "workspace-configmap", "deployment", "service", "ingress"}
		for _, key := range order {
			if content, ok := rendered[key]; ok {
				tmpFile.WriteString("---\n")
				tmpFile.WriteString(content)
				tmpFile.WriteString("\n")
			}
		}
		tmpFile.Close()

		if err := vcm.ApplyManifest(ctx, cfg.Cluster, namespace, tmpPath); err != nil {
			fmt.Printf("  ⚠️  Failed to apply manifests for %s: %v\n", agent.Name, err)
		} else {
			fmt.Printf("  ✅ Agent '%s' deployed\n", agent.Name)
		}
		os.Remove(tmpPath)
	}

	fmt.Printf("\n  🎉 Done! %d agent(s) deployed in vCluster '%s'.\n\n", len(cfg.Agents), cfg.Cluster)
	return nil
}

func init() {
	deployCmd.Flags().StringVar(&deployPreset, "preset", "", "Use a built-in preset configuration")
	deployCmd.Flags().StringVar(&deployEnvFile, "env-file", "", "Path to .env file for secrets")
	deployCmd.Flags().StringVar(&deployAgents, "agents", "", "Comma-separated list of agent names")
	deployCmd.Flags().StringVar(&deployModel, "model", "", "Default LLM model for agents")
	rootCmd.AddCommand(deployCmd)
}
