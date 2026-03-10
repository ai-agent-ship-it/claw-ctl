package wizard

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ai-agent-ship-it/claw-ctl/pkg/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99")).MarginBottom(1)
	promptStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Bold(true)
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	successStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("120"))
)

type step int

const (
	stepPreset step = iota
	stepClusterName
	stepNumAgents
	stepAgentName
	stepModel
	stepChannels
	stepMaxTokens
	stepTemperature
	stepSecretMode
	stepEnvFile
	stepVaultAddr
	stepSecretGate
	stepReview
)

type model struct {
	step         step
	cursor       int
	input        string
	cfg          config.ClusterConfig
	currentAgent int
	agentCount   int
	err          string
	done         bool

	// Secret gate
	secretKeys   []string
	secretValues map[string]string
	secretCursor int
	secretInput  string

	// Channel selection
	channelSelected [4]bool // telegram, discord, whatsapp, http
}

var presetOptions = config.PresetNames()

var modelOptions = []string{
	"ollama/qwen2.5-coder:14b",
	"ollama/llama3.1:8b",
	"gemini/gemini-3-flash",
	"openai/gpt-4o-mini",
}

func initialModel() model {
	return model{
		step:         stepPreset,
		secretValues: make(map[string]string),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.step == stepReview || m.step == stepPreset {
				m.done = true
				return m, tea.Quit
			}
		case "enter":
			return m.handleEnter()
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			m.cursor = m.advanceCursor()
		case " ":
			if m.step == stepChannels {
				m.channelSelected[m.cursor] = !m.channelSelected[m.cursor]
			}
		case "backspace":
			if m.step == stepSecretGate {
				if len(m.secretInput) > 0 {
					m.secretInput = m.secretInput[:len(m.secretInput)-1]
				}
			} else if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		default:
			if len(msg.String()) == 1 {
				if m.step == stepSecretGate {
					m.secretInput += msg.String()
				} else {
					m.input += msg.String()
				}
			}
		}
	}
	return m, nil
}

func (m model) advanceCursor() int {
	switch m.step {
	case stepPreset:
		if m.cursor < len(presetOptions)-1 {
			return m.cursor + 1
		}
	case stepModel:
		if m.cursor < len(modelOptions)-1 {
			return m.cursor + 1
		}
	case stepChannels:
		if m.cursor < 3 {
			return m.cursor + 1
		}
	case stepSecretMode:
		if m.cursor < 2 {
			return m.cursor + 1
		}
	case stepReview:
		if m.cursor < 2 {
			return m.cursor + 1
		}
	}
	return m.cursor
}

func (m model) handleEnter() (tea.Model, tea.Cmd) {
	m.err = ""

	switch m.step {
	case stepPreset:
		selected := presetOptions[m.cursor]
		if selected == "custom" {
			m.step = stepClusterName
			m.cursor = 0
		} else {
			preset := config.Presets[selected]
			m.cfg = preset
			m.step = stepClusterName
			m.cursor = 0
		}

	case stepClusterName:
		name := strings.TrimSpace(m.input)
		if name == "" {
			m.err = "Cluster name cannot be empty"
			return m, nil
		}
		m.cfg.Cluster = name
		m.input = ""
		if len(m.cfg.Agents) > 0 {
			// Preset already has agents, skip to secret mode
			m.step = stepSecretMode
		} else {
			m.step = stepNumAgents
		}
		m.cursor = 0

	case stepNumAgents:
		n, err := strconv.Atoi(strings.TrimSpace(m.input))
		if err != nil || n < 1 || n > 10 {
			m.err = "Enter a number between 1 and 10"
			return m, nil
		}
		m.agentCount = n
		m.currentAgent = 0
		m.input = ""
		m.step = stepAgentName

	case stepAgentName:
		name := strings.TrimSpace(m.input)
		if name == "" {
			m.err = "Agent name cannot be empty"
			return m, nil
		}
		m.cfg.Agents = append(m.cfg.Agents, config.AgentConfig{Name: name})
		m.input = ""
		m.step = stepModel
		m.cursor = 0

	case stepModel:
		m.cfg.Agents[m.currentAgent].Model = modelOptions[m.cursor]
		m.step = stepChannels
		m.cursor = 0
		m.channelSelected = [4]bool{false, false, false, true} // HTTP default on

	case stepChannels:
		agent := &m.cfg.Agents[m.currentAgent]
		if m.channelSelected[0] {
			agent.Channels.Telegram = &config.ChannelTelegram{Enabled: true}
		}
		if m.channelSelected[1] {
			agent.Channels.Discord = &config.ChannelDiscord{Enabled: true}
		}
		if m.channelSelected[2] {
			agent.Channels.WhatsApp = &config.ChannelWhatsApp{Enabled: true}
		}
		if m.channelSelected[3] {
			agent.Channels.HTTP = &config.ChannelHTTP{Enabled: true}
		}
		m.step = stepMaxTokens
		m.input = "8192"

	case stepMaxTokens:
		n, err := strconv.Atoi(strings.TrimSpace(m.input))
		if err != nil || n < 256 {
			m.err = "Enter a valid number (min 256)"
			return m, nil
		}
		m.cfg.Agents[m.currentAgent].MaxTokens = n
		m.input = "0.2"
		m.step = stepTemperature

	case stepTemperature:
		t, err := strconv.ParseFloat(strings.TrimSpace(m.input), 64)
		if err != nil || t < 0 || t > 2 {
			m.err = "Enter a number between 0.0 and 2.0"
			return m, nil
		}
		m.cfg.Agents[m.currentAgent].Temperature = t
		m.input = ""
		m.currentAgent++
		if m.currentAgent < m.agentCount {
			m.step = stepAgentName
		} else {
			m.step = stepSecretMode
			m.cursor = 0
		}

	case stepSecretMode:
		switch m.cursor {
		case 0:
			m.cfg.Secrets.Mode = "env"
			m.step = stepEnvFile
			m.input = ".env"
		case 1:
			m.cfg.Secrets.Mode = "vault"
			m.step = stepVaultAddr
			m.input = "https://vault.reynoso.pro"
		case 2:
			m.cfg.Secrets.Mode = "manual"
			m.step = stepSecretGate
		}
		m.cursor = 0

	case stepEnvFile:
		m.cfg.Secrets.EnvFile = strings.TrimSpace(m.input)
		m.input = ""
		m.step = stepSecretGate

	case stepVaultAddr:
		m.cfg.Secrets.VaultAddr = strings.TrimSpace(m.input)
		m.input = ""
		m.step = stepSecretGate

	case stepSecretGate:
		if len(m.secretKeys) == 0 {
			m.secretKeys = m.cfg.AllRequiredSecrets()
			m.secretCursor = 0
		}
		if m.secretCursor < len(m.secretKeys) {
			val := strings.TrimSpace(m.secretInput)
			if val == "" {
				m.err = fmt.Sprintf("%s is required", m.secretKeys[m.secretCursor])
				return m, nil
			}
			m.secretValues[m.secretKeys[m.secretCursor]] = val
			m.secretInput = ""
			m.secretCursor++
		}
		if m.secretCursor >= len(m.secretKeys) {
			m.step = stepReview
			m.cursor = 0
		}

	case stepReview:
		switch m.cursor {
		case 0: // Deploy
			m.done = true
			return m, tea.Quit
		case 1: // Save
			m.done = true
			return m, tea.Quit
		case 2: // Edit → restart
			m.step = stepPreset
			m.cursor = 0
			m.cfg = config.ClusterConfig{}
			m.secretKeys = nil
			m.secretValues = make(map[string]string)
		}
	}

	return m, nil
}

func (m model) View() string {
	var s strings.Builder

	s.WriteString("\n")
	s.WriteString(titleStyle.Render("  🐾 PicoClaw Setup Wizard"))
	s.WriteString("\n  ========================\n\n")

	switch m.step {
	case stepPreset:
		s.WriteString(promptStyle.Render("  🎭 Choose a configuration preset:\n\n"))
		for i, name := range presetOptions {
			desc := config.PresetDescriptions[name]
			if i == m.cursor {
				s.WriteString(selectedStyle.Render(fmt.Sprintf("  › %s\n", desc)))
			} else {
				s.WriteString(dimStyle.Render(fmt.Sprintf("    %s\n", desc)))
			}
		}

	case stepClusterName:
		s.WriteString(promptStyle.Render("  🏗️  Cluster name: "))
		s.WriteString(m.input)
		s.WriteString("█")

	case stepNumAgents:
		s.WriteString(promptStyle.Render("  👥 How many agents? (1-10): "))
		s.WriteString(m.input)
		s.WriteString("█")

	case stepAgentName:
		s.WriteString(dimStyle.Render(fmt.Sprintf("  ── Agent %d/%d ──\n\n", m.currentAgent+1, m.agentCount)))
		s.WriteString(promptStyle.Render("  📛 Agent name: "))
		s.WriteString(m.input)
		s.WriteString("█")

	case stepModel:
		s.WriteString(dimStyle.Render(fmt.Sprintf("  ── Agent: %s ──\n\n", m.cfg.Agents[m.currentAgent].Name)))
		s.WriteString(promptStyle.Render("  🧠 LLM Model:\n\n"))
		for i, opt := range modelOptions {
			if i == m.cursor {
				s.WriteString(selectedStyle.Render(fmt.Sprintf("  › %s\n", opt)))
			} else {
				s.WriteString(dimStyle.Render(fmt.Sprintf("    %s\n", opt)))
			}
		}

	case stepChannels:
		s.WriteString(promptStyle.Render("  💬 Channels (space to toggle, enter to confirm):\n\n"))
		labels := []string{"Telegram", "Discord", "WhatsApp", "HTTP API"}
		for i, label := range labels {
			check := "☐"
			if m.channelSelected[i] {
				check = "☑"
			}
			if i == m.cursor {
				s.WriteString(selectedStyle.Render(fmt.Sprintf("  › %s %s\n", check, label)))
			} else {
				s.WriteString(dimStyle.Render(fmt.Sprintf("    %s %s\n", check, label)))
			}
		}

	case stepMaxTokens:
		s.WriteString(promptStyle.Render("  🔧 Max tokens: "))
		s.WriteString(m.input)
		s.WriteString("█")

	case stepTemperature:
		s.WriteString(promptStyle.Render("  🌡️  Temperature: "))
		s.WriteString(m.input)
		s.WriteString("█")

	case stepSecretMode:
		s.WriteString(promptStyle.Render("  🔐 Secret storage:\n\n"))
		opts := []string{".env file", "Vault", "Manual (skip for now)"}
		for i, opt := range opts {
			if i == m.cursor {
				s.WriteString(selectedStyle.Render(fmt.Sprintf("  › %s\n", opt)))
			} else {
				s.WriteString(dimStyle.Render(fmt.Sprintf("    %s\n", opt)))
			}
		}

	case stepEnvFile:
		s.WriteString(promptStyle.Render("  📄 Path to .env file: "))
		s.WriteString(m.input)
		s.WriteString("█")

	case stepVaultAddr:
		s.WriteString(promptStyle.Render("  🏦 Vault address: "))
		s.WriteString(m.input)
		s.WriteString("█")

	case stepSecretGate:
		if len(m.secretKeys) == 0 {
			m.secretKeys = m.cfg.AllRequiredSecrets()
		}
		if len(m.secretKeys) == 0 {
			s.WriteString(successStyle.Render("  ✅ No additional tokens required (Ollama model)\n"))
			s.WriteString(dimStyle.Render("\n  Press enter to continue..."))
		} else {
			s.WriteString(promptStyle.Render("  🚧 Required Credentials\n\n"))
			for i, key := range m.secretKeys {
				if val, ok := m.secretValues[key]; ok && val != "" {
					s.WriteString(successStyle.Render(fmt.Sprintf("  ✅ %s: ••••••••\n", key)))
				} else if i == m.secretCursor {
					s.WriteString(promptStyle.Render(fmt.Sprintf("  🔐 %s: ", key)))
					s.WriteString(strings.Repeat("•", len(m.secretInput)))
					s.WriteString("█\n")
				} else {
					s.WriteString(dimStyle.Render(fmt.Sprintf("  ❌ %s: ______\n", key)))
				}
			}
		}

	case stepReview:
		s.WriteString(promptStyle.Render("  ── Review ───────────────────────────\n"))
		if m.cfg.Preset != "" {
			s.WriteString(fmt.Sprintf("  Preset:   %s\n", m.cfg.Preset))
		}
		s.WriteString(fmt.Sprintf("  Cluster:  %s\n", m.cfg.Cluster))
		for _, a := range m.cfg.Agents {
			channels := agentChannelSummary(a)
			s.WriteString(fmt.Sprintf("  Agent:    %s (%s, %s)\n", a.Name, a.Model, channels))
		}
		s.WriteString(fmt.Sprintf("  Secrets:  %s\n", m.cfg.Secrets.Mode))
		if len(m.secretValues) > 0 {
			s.WriteString(fmt.Sprintf("  Tokens:   %d collected ✅\n", len(m.secretValues)))
		}
		s.WriteString(promptStyle.Render("  ─────────────────────────────────────\n\n"))

		opts := []string{"🚀 Deploy now", "💾 Save config", "✏️  Edit"}
		for i, opt := range opts {
			if i == m.cursor {
				s.WriteString(selectedStyle.Render(fmt.Sprintf("  › %s\n", opt)))
			} else {
				s.WriteString(dimStyle.Render(fmt.Sprintf("    %s\n", opt)))
			}
		}
	}

	if m.err != "" {
		s.WriteString("\n")
		s.WriteString(errorStyle.Render(fmt.Sprintf("  ⚠️  %s", m.err)))
	}

	s.WriteString(dimStyle.Render("\n\n  Press ctrl+c to quit"))
	return s.String()
}

func agentChannelSummary(a config.AgentConfig) string {
	var parts []string
	if a.Channels.Telegram != nil && a.Channels.Telegram.Enabled {
		parts = append(parts, "TG")
	}
	if a.Channels.Discord != nil && a.Channels.Discord.Enabled {
		parts = append(parts, "DC")
	}
	if a.Channels.WhatsApp != nil && a.Channels.WhatsApp.Enabled {
		parts = append(parts, "WA")
	}
	if a.Channels.HTTP != nil && a.Channels.HTTP.Enabled {
		parts = append(parts, "HTTP")
	}
	if len(parts) == 0 {
		return "no channels"
	}
	return strings.Join(parts, "+")
}

// WizardResult holds the output from the wizard.
type WizardResult struct {
	Config       config.ClusterConfig
	SecretValues map[string]string
	Action       string // "deploy", "save", "cancel"
}

// RunWizard launches the interactive TUI wizard.
func RunWizard() (*WizardResult, error) {
	m := initialModel()
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("wizard error: %w", err)
	}

	result := finalModel.(model)
	if !result.done {
		return &WizardResult{Action: "cancel"}, nil
	}

	action := "deploy"
	if result.step == stepReview && result.cursor == 1 {
		action = "save"
	}

	return &WizardResult{
		Config:       result.cfg,
		SecretValues: result.secretValues,
		Action:       action,
	}, nil
}
