package vcluster

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Manager handles vCluster lifecycle operations.
type Manager struct{}

// NewManager creates a new vCluster manager.
func NewManager() *Manager {
	return &Manager{}
}

// Create creates a new vCluster in the given namespace (idempotent).
func (m *Manager) Create(ctx context.Context, name, namespace string) error {
	fmt.Printf("  ☸️  Creating vCluster '%s' in namespace '%s'...\n", name, namespace)

	cmd := exec.CommandContext(ctx, "vcluster", "create", name,
		"--namespace", namespace,
		"--connect=false",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outStr := string(output)
		if strings.Contains(outStr, "already exists") || strings.Contains(outStr, "already running") {
			fmt.Println("  ℹ️  vCluster already exists, reusing")
			return nil
		}
		return fmt.Errorf("vcluster create failed: %s: %w", outStr, err)
	}
	fmt.Println("  ✅ vCluster created")
	return nil
}

// WaitReady waits until the vCluster pod is ready.
func (m *Manager) WaitReady(ctx context.Context, name, namespace string) error {
	fmt.Printf("  ⏳ Waiting for vCluster '%s' readiness...\n", name)

	deadline := time.Now().Add(3 * time.Minute)
	for time.Now().Before(deadline) {
		cmd := exec.CommandContext(ctx, "kubectl", "get", "pods",
			"-n", namespace,
			"-l", fmt.Sprintf("app=vcluster,release=%s", name),
			"-o", "jsonpath={.items[0].status.phase}",
		)
		output, err := cmd.Output()
		if err == nil && strings.TrimSpace(string(output)) == "Running" {
			fmt.Println("  ✅ vCluster is ready")
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("vcluster '%s' did not become ready within 3 minutes", name)
}

// ApplyManifest applies a YAML manifest inside the vCluster using `vcluster connect -- kubectl apply`.
// This handles the proxy/port-forward automatically without needing a kubeconfig file.
func (m *Manager) ApplyManifest(ctx context.Context, name, namespace, manifestPath string) error {
	cmd := exec.CommandContext(ctx, "vcluster", "connect", name,
		"--namespace", namespace,
		"--", "kubectl", "apply", "-f", manifestPath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}

// Connect exports the vCluster kubeconfig to a file (for status/debug use).
// Extracts the kubeconfig from the vc-<name> Secret and base64-decodes in Go.
func (m *Manager) Connect(ctx context.Context, name, namespace string) error {
	fmt.Printf("  🔗 Exporting vCluster '%s' kubeconfig...\n", name)

	kubeconfigPath := KubeconfigPath(name)

	// Extract base64-encoded kubeconfig from the vc-<name> secret
	secretName := fmt.Sprintf("vc-%s", name)
	cmd := exec.CommandContext(ctx, "kubectl", "get", "secret", secretName,
		"-n", namespace,
		"-o", "jsonpath={.data.config}",
	)
	encoded, err := cmd.Output()
	if err == nil && len(encoded) > 0 {
		decoded, err := base64.StdEncoding.DecodeString(string(encoded))
		if err != nil {
			return fmt.Errorf("failed to decode kubeconfig base64: %w", err)
		}
		if err := os.WriteFile(kubeconfigPath, decoded, 0600); err != nil {
			return fmt.Errorf("failed to write kubeconfig: %w", err)
		}
		fmt.Printf("  ✅ Kubeconfig written to %s\n", kubeconfigPath)
		return nil
	}

	return fmt.Errorf("could not find kubeconfig in secret '%s' in namespace '%s'", secretName, namespace)
}

// Delete deletes a vCluster (idempotent).
func (m *Manager) Delete(ctx context.Context, name, namespace string) error {
	fmt.Printf("  🗑️  Deleting vCluster '%s'...\n", name)

	cmd := exec.CommandContext(ctx, "vcluster", "delete", name,
		"--namespace", namespace,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outStr := string(output)
		if strings.Contains(outStr, "not found") || strings.Contains(outStr, "couldn't find") {
			fmt.Println("  ℹ️  vCluster not found (already deleted)")
			return nil
		}
		return fmt.Errorf("vcluster delete failed: %s: %w", outStr, err)
	}
	fmt.Println("  ✅ vCluster deleted")
	return nil
}

// KubeconfigPath returns the path to the vCluster kubeconfig.
func KubeconfigPath(name string) string {
	return fmt.Sprintf("/tmp/vcluster-%s-kubeconfig.yaml", name)
}
