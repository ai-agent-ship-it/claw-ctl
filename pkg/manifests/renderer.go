package manifests

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/ai-agent-ship-it/claw-ctl/pkg/config"
)

//go:embed embed/*
var templateFS embed.FS

// templateFuncs provides helper functions for templates.
var templateFuncs = template.FuncMap{
	"hasPrefix": strings.HasPrefix,
}

// RenderManifest renders a single template file with the given agent config.
func RenderManifest(templateName string, agent config.AgentConfig) (string, error) {
	data, err := templateFS.ReadFile("embed/" + templateName)
	if err != nil {
		return "", fmt.Errorf("template %s not found: %w", templateName, err)
	}

	tmpl, err := template.New(templateName).Funcs(templateFuncs).Parse(string(data))
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", templateName, err)
	}

	// Build template data from agent config
	tmplData := map[string]interface{}{
		"AgentName":   agent.Name,
		"Model":       agent.Model,
		"MaxTokens":   agent.MaxTokens,
		"Temperature": agent.Temperature,
		"Channels":    agent.Channels,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, tmplData); err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", templateName, err)
	}

	return buf.String(), nil
}

// RenderStaticManifest renders a static (non-templated) manifest file.
func RenderStaticManifest(filename string) (string, error) {
	data, err := templateFS.ReadFile("embed/" + filename)
	if err != nil {
		return "", fmt.Errorf("manifest %s not found: %w", filename, err)
	}
	return string(data), nil
}

// RenderAllForAgent renders all manifest templates for a single agent.
func RenderAllForAgent(agent config.AgentConfig) (map[string]string, error) {
	results := make(map[string]string)

	// Static manifests
	ns, err := RenderStaticManifest("namespace.yaml")
	if err != nil {
		return nil, err
	}
	results["namespace"] = ns

	// Templated manifests
	templates := []struct {
		key      string
		filename string
	}{
		{"rbac", "rbac.yaml.tmpl"},
		{"pvc", "pvc.yaml.tmpl"},
		{"configmap", "configmap.yaml.tmpl"},
		{"deployment", "deployment.yaml.tmpl"},
		{"service", "service.yaml.tmpl"},
		{"ingress", "ingress.yaml.tmpl"},
	}

	for _, t := range templates {
		rendered, err := RenderManifest(t.filename, agent)
		if err != nil {
			return nil, fmt.Errorf("rendering %s: %w", t.key, err)
		}
		results[t.key] = rendered
	}

	return results, nil
}
