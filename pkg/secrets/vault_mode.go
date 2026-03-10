package secrets

import (
	"fmt"

	"github.com/ai-agent-ship-it/claw-ctl/pkg/config"
)

// VaultProvisioner handles Vault operations for secret management.
type VaultProvisioner struct {
	Address string
	Token   string
}

// NewVaultProvisioner creates a new Vault provisioner.
func NewVaultProvisioner(address string) (*VaultProvisioner, error) {
	if address == "" {
		return nil, fmt.Errorf("vault address is required")
	}

	// TODO: Read token from ~/.vault-token or VAULT_TOKEN env
	return &VaultProvisioner{
		Address: address,
	}, nil
}

// ProvisionAgent creates KV path, policy, and auth role for an agent.
func (v *VaultProvisioner) ProvisionAgent(clusterName string, agent config.AgentConfig) error {
	agentKey := fmt.Sprintf("%s-%s", clusterName, agent.Name)
	kvPath := fmt.Sprintf("secret/agents/%s/%s", clusterName, agent.Name)

	fmt.Printf("  🏦 Vault: Creating KV path: %s\n", kvPath)
	// TODO: v.createKVPath(kvPath)

	policyName := fmt.Sprintf("picoclaw-%s", agentKey)
	fmt.Printf("  🏦 Vault: Creating policy: %s\n", policyName)
	// TODO: v.createPolicy(policyName, kvPath)

	roleName := fmt.Sprintf("picoclaw-%s", agentKey)
	fmt.Printf("  🏦 Vault: Creating auth role: %s\n", roleName)
	// TODO: v.createAuthRole(roleName, clusterName, agent.Name)

	return nil
}

// CleanupAgent removes Vault resources for an agent.
func (v *VaultProvisioner) CleanupAgent(clusterName string, agentName string) error {
	agentKey := fmt.Sprintf("%s-%s", clusterName, agentName)

	fmt.Printf("  🏦 Vault: Removing policy: picoclaw-%s\n", agentKey)
	// TODO: v.deletePolicy

	fmt.Printf("  🏦 Vault: Removing auth role: picoclaw-%s\n", agentKey)
	// TODO: v.deleteAuthRole

	fmt.Printf("  🏦 Vault: Removing KV path: secret/agents/%s/%s\n", clusterName, agentName)
	// TODO: v.deleteKVPath

	return nil
}
