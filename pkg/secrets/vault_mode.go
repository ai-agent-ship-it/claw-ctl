package secrets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ai-agent-ship-it/claw-ctl/pkg/config"
)

// VaultProvisioner handles Vault operations for secret management.
type VaultProvisioner struct {
	Address string
	Token   string
	client  *http.Client
}

// NewVaultProvisioner creates a new Vault provisioner with the given address and token.
func NewVaultProvisioner(address, token string) (*VaultProvisioner, error) {
	if address == "" {
		return nil, fmt.Errorf("vault address is required")
	}
	if token == "" {
		return nil, fmt.Errorf("vault token is required")
	}

	return &VaultProvisioner{
		Address: strings.TrimRight(address, "/"),
		Token:   token,
		client:  &http.Client{},
	}, nil
}

// ProvisionAgent creates KV path, policy, and K8s auth role for an agent.
func (v *VaultProvisioner) ProvisionAgent(clusterName string, agent config.AgentConfig) error {
	agentKey := fmt.Sprintf("%s-%s", clusterName, agent.Name)
	kvPath := fmt.Sprintf("agents/%s/%s", clusterName, agent.Name)
	policyName := fmt.Sprintf("picoclaw-%s", agentKey)

	// Step 1: Enable KV v2 secrets engine (idempotent)
	fmt.Printf("  🏦 Vault: Ensuring KV engine at secret/...\n")
	if err := v.enableKV(); err != nil {
		fmt.Printf("  ℹ️  KV engine: %v (may already be enabled)\n", err)
	}

	// Step 2: Write initial secret data at the KV path
	fmt.Printf("  🏦 Vault: Creating KV path: secret/%s\n", kvPath)
	secretData := map[string]interface{}{
		"managed_by": "claw-ctl",
		"agent":      agent.Name,
		"cluster":    clusterName,
	}
	// Add required secrets as placeholders
	for _, key := range agent.RequiredSecrets() {
		secretData[key] = "REPLACE_ME"
	}
	if err := v.writeKV(kvPath, secretData); err != nil {
		return fmt.Errorf("failed to write KV: %w", err)
	}

	// Step 3: Create policy granting read access to this path
	fmt.Printf("  🏦 Vault: Creating policy: %s\n", policyName)
	policy := fmt.Sprintf(`
path "secret/data/%s" {
  capabilities = ["read", "list"]
}
path "secret/metadata/%s" {
  capabilities = ["read", "list"]
}
`, kvPath, kvPath)
	if err := v.writePolicy(policyName, policy); err != nil {
		return fmt.Errorf("failed to create policy: %w", err)
	}

	// Step 4: Create K8s auth role
	roleName := fmt.Sprintf("picoclaw-%s", agentKey)
	fmt.Printf("  🏦 Vault: Creating K8s auth role: %s\n", roleName)
	if err := v.createK8sAuthRole(roleName, policyName, clusterName, agent.Name); err != nil {
		fmt.Printf("  ⚠️  K8s auth role: %v (may need K8s auth method enabled)\n", err)
	}

	fmt.Printf("  ✅ Vault provisioned for agent '%s'\n", agent.Name)
	return nil
}

// CleanupAgent removes Vault resources for an agent.
func (v *VaultProvisioner) CleanupAgent(clusterName, agentName string) error {
	agentKey := fmt.Sprintf("%s-%s", clusterName, agentName)
	kvPath := fmt.Sprintf("agents/%s/%s", clusterName, agentName)
	policyName := fmt.Sprintf("picoclaw-%s", agentKey)
	roleName := fmt.Sprintf("picoclaw-%s", agentKey)

	fmt.Printf("  🏦 Vault: Removing policy: %s\n", policyName)
	_ = v.deletePolicy(policyName)

	fmt.Printf("  🏦 Vault: Removing K8s auth role: %s\n", roleName)
	_ = v.deleteK8sAuthRole(roleName)

	fmt.Printf("  🏦 Vault: Removing KV path: secret/%s\n", kvPath)
	_ = v.deleteKV(kvPath)

	return nil
}

// --- Vault HTTP API methods ---

func (v *VaultProvisioner) vaultRequest(method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}

	url := fmt.Sprintf("%s/v1/%s", v.Address, path)
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Vault-Token", v.Token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vault request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("vault %s %s: %d %s", method, path, resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

func (v *VaultProvisioner) enableKV() error {
	payload := map[string]interface{}{
		"type":    "kv",
		"options": map[string]string{"version": "2"},
	}
	_, err := v.vaultRequest("POST", "sys/mounts/secret", payload)
	return err
}

func (v *VaultProvisioner) writeKV(path string, data map[string]interface{}) error {
	payload := map[string]interface{}{
		"data": data,
	}
	_, err := v.vaultRequest("POST", "secret/data/"+path, payload)
	return err
}

func (v *VaultProvisioner) deleteKV(path string) error {
	_, err := v.vaultRequest("DELETE", "secret/metadata/"+path, nil)
	return err
}

func (v *VaultProvisioner) writePolicy(name, rules string) error {
	payload := map[string]interface{}{
		"policy": rules,
	}
	_, err := v.vaultRequest("PUT", "sys/policies/acl/"+name, payload)
	return err
}

func (v *VaultProvisioner) deletePolicy(name string) error {
	_, err := v.vaultRequest("DELETE", "sys/policies/acl/"+name, nil)
	return err
}

func (v *VaultProvisioner) createK8sAuthRole(roleName, policyName, namespace, serviceAccount string) error {
	payload := map[string]interface{}{
		"bound_service_account_names":      []string{serviceAccount},
		"bound_service_account_namespaces": []string{"vcluster-" + namespace, "agents"},
		"policies":                         []string{policyName},
		"ttl":                              "24h",
	}
	_, err := v.vaultRequest("POST", "auth/kubernetes/role/"+roleName, payload)
	return err
}

func (v *VaultProvisioner) deleteK8sAuthRole(roleName string) error {
	_, err := v.vaultRequest("DELETE", "auth/kubernetes/role/"+roleName, nil)
	return err
}
