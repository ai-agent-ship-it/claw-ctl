package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	useVault   bool
	vaultAddr  string
	vaultToken string
	presetName string
	ollamaAddr string
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
	rootCmd.PersistentFlags().BoolVar(&useVault, "vault", false, "Use Vault for secrets (reads VAULT_ADDR, VAULT_TOKEN from env or .env)")
	rootCmd.PersistentFlags().StringVar(&vaultAddr, "vault-addr", "", "Vault server address (overrides VAULT_ADDR)")
	rootCmd.PersistentFlags().StringVar(&vaultToken, "vault-token", "", "Vault token (overrides VAULT_TOKEN)")
	rootCmd.PersistentFlags().StringVar(&ollamaAddr, "ollama-addr", "", "Ollama server address (overrides OLLAMA_ADDR, e.g. http://192.168.1.100:11434)")
}

// resolveVaultConfig resolves Vault address and token from flags, env vars, or .env file.
func resolveVaultConfig(envFile string) (string, string, error) {
	addr := vaultAddr
	token := vaultToken

	// Fallback to env vars
	if addr == "" {
		addr = os.Getenv("VAULT_ADDR")
	}
	if token == "" {
		token = os.Getenv("VAULT_TOKEN")
	}

	// Fallback to .env file
	if (addr == "" || token == "") && envFile != "" {
		// Lazy import to avoid circular deps — read .env inline
		envVars, _ := readEnvSimple(envFile)
		if addr == "" {
			addr = envVars["VAULT_ADDR"]
		}
		if token == "" {
			token = envVars["VAULT_TOKEN"]
		}
	}

	if addr == "" {
		return "", "", fmt.Errorf("VAULT_ADDR not set (use --vault-addr, env var, or .env file)")
	}
	if token == "" {
		return "", "", fmt.Errorf("VAULT_TOKEN not set (use --vault-token, env var, or .env file)")
	}

	return addr, token, nil
}

// resolveOllamaAddr resolves the Ollama server address from flags, env vars, or .env file.
func resolveOllamaAddr(envFile string) string {
	addr := ollamaAddr

	// Fallback to env var
	if addr == "" {
		addr = os.Getenv("OLLAMA_ADDR")
	}

	// Fallback to .env file
	if addr == "" && envFile != "" {
		envVars, _ := readEnvSimple(envFile)
		addr = envVars["OLLAMA_ADDR"]
	}

	return addr
}

// readEnvSimple reads key=value pairs from a file (minimal .env parser).
func readEnvSimple(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, line := range splitLines(string(data)) {
		line = trimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}
		idx := indexOf(line, '=')
		if idx < 0 {
			continue
		}
		key := trimSpace(line[:idx])
		val := trimSpace(line[idx+1:])
		val = trimQuotes(val)
		result[key] = val
	}
	return result, nil
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimSpace(s string) string {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\r') {
		i++
	}
	j := len(s)
	for j > i && (s[j-1] == ' ' || s[j-1] == '\t' || s[j-1] == '\r') {
		j--
	}
	return s[i:j]
}

func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func trimQuotes(s string) string {
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		return s[1 : len(s)-1]
	}
	return s
}
