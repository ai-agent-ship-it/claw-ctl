package workspace

import (
	"fmt"
	"os"
	"path/filepath"
)

// DefaultSOUL returns the default SOUL.md content for an agent.
func DefaultSOUL(agentName string) string {
	return fmt.Sprintf(`# Soul

I am %s, an AI assistant powered by PicoClaw.

## Personality

- Helpful and proactive
- Concise and to the point
- Curious and eager to learn
- Honest and transparent

## Values

- Accuracy over speed
- User privacy and safety
- Transparency in actions
- Continuous improvement
`, agentName)
}

// DefaultIDENTITY returns the default IDENTITY.md content.
func DefaultIDENTITY(agentName, model string) string {
	return fmt.Sprintf(`# Identity

## Name
%s 🦞

## Model
%s

## Purpose
- Provide intelligent AI assistance
- Support autonomous DevOps operations
- Enable easy customization through skills

## Capabilities
- Web search and content fetching
- File system operations
- Shell command execution
- Multi-channel messaging
- Skill-based extensibility
- Memory and context management
`, agentName, model)
}

// DefaultUSER returns the default USER.md content.
func DefaultUSER() string {
	return `# User

## Preferences
- Language: Spanish (es-MX) / English
- Timezone: America/Mexico_City
- Communication style: Direct and concise
`
}

// DefaultAGENT returns the default AGENT.md content.
func DefaultAGENT(agentName string) string {
	return fmt.Sprintf(`# Agent Instructions

## Name
%s

## Rules
- Never expose secrets or API keys
- Always verify before destructive operations
- Keep responses concise
- Use tools efficiently
`, agentName)
}

// DefaultENVIRONMENT returns the auto-generated ENVIRONMENT.md.
func DefaultENVIRONMENT(capabilities []string) string {
	content := `# Cluster Environment Context

You are running inside a virtualized Kubernetes cluster (vCluster).

## Core Capabilities & Installed Controllers
When generating manifests, use the following pre-installed infrastructure:

1. **Ingress Controller (Traefik)**
   - Class Name: traefik
   - Usage: For exposing services, always use ingressClassName: traefik

2. **Databases (CloudNativePG)**
   - API Group: postgresql.cnpg.io/v1
   - Kind: Cluster, Pooler, ScheduledBackup

3. **Secrets Management**
   - Use native Kubernetes Secret resources
   - CANNOT read secrets due to RBAC (Crystal Wall)

## Security Constraints (Crystal Wall)
- **No Secret Access**: Blocked from list/read Kubernetes Secrets
- **No Pod Exec**: Cannot exec into pods
`
	return content
}

// GenerateWorkspace creates the workspace directory with default files for an agent.
func GenerateWorkspace(baseDir, agentName, model string, capabilities []string) error {
	dir := filepath.Join(baseDir, agentName)

	dirs := []string{
		dir,
		filepath.Join(dir, "memory"),
		filepath.Join(dir, "skills"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", d, err)
		}
	}

	files := map[string]string{
		filepath.Join(dir, "SOUL.md"):             DefaultSOUL(agentName),
		filepath.Join(dir, "IDENTITY.md"):         DefaultIDENTITY(agentName, model),
		filepath.Join(dir, "USER.md"):             DefaultUSER(),
		filepath.Join(dir, "AGENT.md"):            DefaultAGENT(agentName),
		filepath.Join(dir, "ENVIRONMENT.md"):      DefaultENVIRONMENT(capabilities),
		filepath.Join(dir, "memory", "MEMORY.md"): "# Memory\n\nPersistent facts this agent has learned.\n",
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", path, err)
		}
	}

	return nil
}
