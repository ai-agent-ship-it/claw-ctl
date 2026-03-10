package config

// ClusterConfig is the top-level configuration for a claw-ctl deployment.
// It maps directly to picoclaw.yaml.
type ClusterConfig struct {
	Cluster string        `yaml:"cluster" json:"cluster"`
	Preset  string        `yaml:"preset,omitempty" json:"preset,omitempty"`
	Secrets SecretsConfig `yaml:"secrets" json:"secrets"`
	Agents  []AgentConfig `yaml:"agents" json:"agents"`
}

// SecretsConfig defines how secrets are managed.
type SecretsConfig struct {
	Mode      string `yaml:"mode" json:"mode"` // env | vault | manual
	EnvFile   string `yaml:"envFile,omitempty" json:"envFile,omitempty"`
	VaultAddr string `yaml:"vaultAddr,omitempty" json:"vaultAddr,omitempty"`
}

// AgentConfig defines a single agent within a cluster.
type AgentConfig struct {
	Name         string          `yaml:"name" json:"name"`
	Model        string          `yaml:"model" json:"model"`
	MaxTokens    int             `yaml:"maxTokens" json:"maxTokens"`
	Temperature  float64         `yaml:"temperature" json:"temperature"`
	OllamaAddr   string          `yaml:"ollamaAddr,omitempty" json:"ollamaAddr,omitempty"`
	Capabilities []string        `yaml:"capabilities,omitempty" json:"capabilities,omitempty"`
	Channels     ChannelsConfig  `yaml:"channels" json:"channels"`
	Workspace    WorkspaceConfig `yaml:"workspace,omitempty" json:"workspace,omitempty"`
}

// ChannelsConfig defines which communication channels are enabled.
type ChannelsConfig struct {
	Telegram *ChannelTelegram `yaml:"telegram,omitempty" json:"telegram,omitempty"`
	Discord  *ChannelDiscord  `yaml:"discord,omitempty" json:"discord,omitempty"`
	WhatsApp *ChannelWhatsApp `yaml:"whatsapp,omitempty" json:"whatsapp,omitempty"`
	HTTP     *ChannelHTTP     `yaml:"http,omitempty" json:"http,omitempty"`
}

type ChannelTelegram struct {
	Enabled   bool     `yaml:"enabled" json:"enabled"`
	AllowFrom []string `yaml:"allowFrom,omitempty" json:"allowFrom,omitempty"`
}

type ChannelDiscord struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	GuildID string `yaml:"guildId,omitempty" json:"guildId,omitempty"`
}

type ChannelWhatsApp struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

type ChannelHTTP struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// WorkspaceConfig defines paths to the agent workspace value files.
type WorkspaceConfig struct {
	Soul        string   `yaml:"soul,omitempty" json:"soul,omitempty"`
	Identity    string   `yaml:"identity,omitempty" json:"identity,omitempty"`
	User        string   `yaml:"user,omitempty" json:"user,omitempty"`
	Agent       string   `yaml:"agent,omitempty" json:"agent,omitempty"`
	Environment string   `yaml:"environment,omitempty" json:"environment,omitempty"`
	Memory      string   `yaml:"memory,omitempty" json:"memory,omitempty"`
	Skills      []string `yaml:"skills,omitempty" json:"skills,omitempty"`
}

// RequiredSecrets returns the list of secret keys that are mandatory
// based on the agent configuration.
func (a *AgentConfig) RequiredSecrets() []string {
	var required []string

	// Channel tokens
	if a.Channels.Telegram != nil && a.Channels.Telegram.Enabled {
		required = append(required, "TELEGRAM_BOT_TOKEN")
	}
	if a.Channels.Discord != nil && a.Channels.Discord.Enabled {
		required = append(required, "DISCORD_BOT_TOKEN")
	}
	if a.Channels.WhatsApp != nil && a.Channels.WhatsApp.Enabled {
		required = append(required, "WHATSAPP_API_TOKEN")
	}

	// Cloud model API keys
	if len(a.Model) > 0 {
		switch {
		case startsWith(a.Model, "gemini/"):
			required = append(required, "GEMINI_API_KEY")
		case startsWith(a.Model, "openai/"):
			required = append(required, "OPENAI_API_KEY")
		case startsWith(a.Model, "anthropic/"):
			required = append(required, "ANTHROPIC_API_KEY")
			// ollama/* and local/* don't need API keys
		}
	}

	return required
}

// AllRequiredSecrets returns the deduplicated set of required secrets
// across all agents in the cluster.
func (c *ClusterConfig) AllRequiredSecrets() []string {
	seen := make(map[string]bool)
	var result []string
	for _, agent := range c.Agents {
		for _, s := range agent.RequiredSecrets() {
			if !seen[s] {
				seen[s] = true
				result = append(result, s)
			}
		}
	}
	return result
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
