package vcluster

import (
	"context"
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

// Connect exports the vCluster kubeconfig to a file.
func (m *Manager) Connect(ctx context.Context, name, namespace string) error {
	fmt.Printf("  🔗 Exporting vCluster '%s' kubeconfig...\n", name)

	kubeconfigPath := KubeconfigPath(name)

	// Use --print to get kubeconfig YAML to stdout
	cmd := exec.CommandContext(ctx, "vcluster", "connect", name,
		"--namespace", namespace,
		"--print",
	)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("vcluster connect --print failed: %w", err)
	}

	// Write kubeconfig to disk
	if err := os.WriteFile(kubeconfigPath, output, 0600); err != nil {
		return fmt.Errorf("failed to write kubeconfig to %s: %w", kubeconfigPath, err)
	}

	fmt.Printf("  ✅ Kubeconfig written to %s\n", kubeconfigPath)
	return nil
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
