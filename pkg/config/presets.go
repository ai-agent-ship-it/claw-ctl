package config

// Presets contains the built-in preset configurations.
var Presets = map[string]ClusterConfig{
	"financial-controller": {
		Preset: "financial-controller",
		Agents: []AgentConfig{
			{
				Name:         "agent-financiero",
				Model:        "ollama/qwen2.5-coder:14b",
				MaxTokens:    32000,
				Temperature:  0.1,
				Capabilities: []string{"traefik", "cnpg", "redis"},
				Channels: ChannelsConfig{
					Telegram: &ChannelTelegram{Enabled: true},
					HTTP:     &ChannelHTTP{Enabled: true},
				},
			},
		},
	},
	"devops-engineer": {
		Preset: "devops-engineer",
		Agents: []AgentConfig{
			{
				Name:         "agent-devops",
				Model:        "ollama/qwen2.5-coder:14b",
				MaxTokens:    32000,
				Temperature:  0.2,
				Capabilities: []string{"traefik", "cnpg"},
				Channels: ChannelsConfig{
					Discord: &ChannelDiscord{Enabled: true},
					HTTP:    &ChannelHTTP{Enabled: true},
				},
			},
		},
	},
	"personal-assistant": {
		Preset: "personal-assistant",
		Agents: []AgentConfig{
			{
				Name:         "agent-assistant",
				Model:        "ollama/llama3.1:8b",
				MaxTokens:    8192,
				Temperature:  0.3,
				Capabilities: []string{"traefik"},
				Channels: ChannelsConfig{
					Telegram: &ChannelTelegram{Enabled: true},
					WhatsApp: &ChannelWhatsApp{Enabled: true},
					HTTP:     &ChannelHTTP{Enabled: true},
				},
			},
		},
	},
	"multi-team": {
		Preset: "multi-team",
		Agents: []AgentConfig{
			{
				Name:         "agent-financiero",
				Model:        "ollama/qwen2.5-coder:14b",
				MaxTokens:    32000,
				Temperature:  0.1,
				Capabilities: []string{"traefik", "cnpg", "redis"},
				Channels: ChannelsConfig{
					Telegram: &ChannelTelegram{Enabled: true},
					HTTP:     &ChannelHTTP{Enabled: true},
				},
			},
			{
				Name:         "agent-devops",
				Model:        "ollama/qwen2.5-coder:14b",
				MaxTokens:    32000,
				Temperature:  0.2,
				Capabilities: []string{"traefik", "cnpg"},
				Channels: ChannelsConfig{
					Discord: &ChannelDiscord{Enabled: true},
					HTTP:    &ChannelHTTP{Enabled: true},
				},
			},
			{
				Name:         "agent-assistant",
				Model:        "ollama/llama3.1:8b",
				MaxTokens:    8192,
				Temperature:  0.3,
				Capabilities: []string{"traefik"},
				Channels: ChannelsConfig{
					Telegram: &ChannelTelegram{Enabled: true},
					HTTP:     &ChannelHTTP{Enabled: true},
				},
			},
		},
	},
	"minimal": {
		Preset: "minimal",
		Agents: []AgentConfig{
			{
				Name:        "agent",
				Model:       "ollama/llama3.1:8b",
				MaxTokens:   8192,
				Temperature: 0.2,
				Channels: ChannelsConfig{
					HTTP: &ChannelHTTP{Enabled: true},
				},
			},
		},
	},
}

// PresetNames returns an ordered list of preset names for display.
func PresetNames() []string {
	return []string{
		"financial-controller",
		"devops-engineer",
		"personal-assistant",
		"multi-team",
		"minimal",
		"custom",
	}
}

// PresetDescriptions returns human-readable descriptions for each preset.
var PresetDescriptions = map[string]string{
	"financial-controller": "💰 Financial Controller — 1 agent, qwen2.5-coder:14b, Telegram+HTTP",
	"devops-engineer":      "🔧 DevOps Engineer — 1 agent, qwen2.5-coder:14b, Discord+HTTP",
	"personal-assistant":   "📋 Personal Assistant — 1 agent, llama3.1:8b, Telegram+WhatsApp",
	"multi-team":           "👥 Multi-Team — 3 agents, mixed models, all channels",
	"minimal":              "⚡ Minimal — 1 agent, llama3.1:8b, HTTP only",
	"custom":               "🛠️  Custom — configure everything manually",
}
