package vcluster

import (
	"context"
	"fmt"
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

// Create creates a new vCluster in the given namespace.
func (m *Manager) Create(ctx context.Context, name, namespace string) error {
	fmt.Printf("  ☸️  Creating vCluster '%s' in namespace '%s'...\n", name, namespace)

	cmd := exec.CommandContext(ctx, "vcluster", "create", name,
		"--namespace", namespace,
		"--connect=false",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("vcluster create failed: %s: %w", string(output), err)
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

// Connect connects to a running vCluster.
func (m *Manager) Connect(ctx context.Context, name, namespace string) error {
	fmt.Printf("  🔗 Connecting to vCluster '%s'...\n", name)

	cmd := exec.CommandContext(ctx, "vcluster", "connect", name,
		"--namespace", namespace,
		"--update-current=false",
		"--kube-config", fmt.Sprintf("/tmp/vcluster-%s-kubeconfig.yaml", name),
	)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("vcluster connect failed: %w", err)
	}

	// Give it time to establish the connection
	time.Sleep(3 * time.Second)
	fmt.Println("  ✅ Connected to vCluster")
	return nil
}

// Delete deletes a vCluster.
func (m *Manager) Delete(ctx context.Context, name, namespace string) error {
	fmt.Printf("  🗑️  Deleting vCluster '%s'...\n", name)

	cmd := exec.CommandContext(ctx, "vcluster", "delete", name,
		"--namespace", namespace,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("vcluster delete failed: %s: %w", string(output), err)
	}
	fmt.Println("  ✅ vCluster deleted")
	return nil
}

// KubeconfigPath returns the path to the vCluster kubeconfig.
func KubeconfigPath(name string) string {
	return fmt.Sprintf("/tmp/vcluster-%s-kubeconfig.yaml", name)
}
