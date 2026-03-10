# 🐾 claw-ctl

**Deploy PicoClaw AI agents to Kubernetes in seconds.**

`claw-ctl` is a CLI tool that guides you through deploying autonomous AI agents on Kubernetes using [vClusters](https://www.vcluster.com/). It handles everything — from interactive configuration to secret management, manifest generation, and deployment.

**The only requirement is an LLM token or a local Ollama address.**

---

## Quick Start

### Install

```bash
# macOS (Apple Silicon)
curl -sSfL https://github.com/ai-agent-ship-it/claw-ctl/releases/latest/download/claw-ctl_darwin_arm64.tar.gz | tar xz
sudo mv claw-ctl /usr/local/bin/

# macOS (Intel)
curl -sSfL https://github.com/ai-agent-ship-it/claw-ctl/releases/latest/download/claw-ctl_darwin_amd64.tar.gz | tar xz
sudo mv claw-ctl /usr/local/bin/

# Linux (x86_64)
curl -sSfL https://github.com/ai-agent-ship-it/claw-ctl/releases/latest/download/claw-ctl_linux_amd64.tar.gz | tar xz
sudo mv claw-ctl /usr/local/bin/

# Linux (ARM64)
curl -sSfL https://github.com/ai-agent-ship-it/claw-ctl/releases/latest/download/claw-ctl_linux_arm64.tar.gz | tar xz
sudo mv claw-ctl /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/ai-agent-ship-it/claw-ctl.git
cd claw-ctl
make build
```

---

## Usage

### Interactive Wizard

```bash
claw-ctl init
```

The wizard guides you step by step:
1. **Choose a preset** or configure manually
2. **Name your cluster**
3. **Configure agents** — model, channels (Telegram, Discord, WhatsApp, HTTP)
4. **Provide required tokens** — only the ones your config needs
5. **Review and deploy**

### Deploy from Preset

```bash
claw-ctl deploy finance --preset financial-controller
```

### Deploy from Config File

```bash
claw-ctl deploy --config picoclaw.yaml
```

### Other Commands

```bash
claw-ctl presets                          # List available presets
claw-ctl status finance                   # Check cluster health
claw-ctl add-agent finance agent-reportes # Add agent to running cluster
claw-ctl reload finance agent-financiero  # Hot-reload workspace files
claw-ctl destroy finance                  # Tear down everything
```

---

## Presets

| Preset | Agents | Model | Channels |
|---|---|---|---|
| `financial-controller` | 1: agent-financiero | qwen2.5-coder:14b | Telegram + HTTP |
| `devops-engineer` | 1: agent-devops | qwen2.5-coder:14b | Discord + HTTP |
| `personal-assistant` | 1: agent-assistant | llama3.1:8b | Telegram + WhatsApp |
| `multi-team` | 3 agents | mixed | All channels |
| `minimal` | 1: agent | llama3.1:8b | HTTP only |

---

## Configuration

`claw-ctl init` generates a `picoclaw.yaml` that you can edit and redeploy:

```yaml
cluster: finance
preset: financial-controller
secrets:
  mode: env
  envFile: .env
agents:
  - name: agent-financiero
    model: ollama/qwen2.5-coder:14b
    maxTokens: 32000
    temperature: 0.1
    channels:
      telegram: { enabled: true }
      http: { enabled: true }
    workspace:
      soul: workspace/agent-financiero/SOUL.md
      identity: workspace/agent-financiero/IDENTITY.md
      skills:
        - workspace/agent-financiero/skills/financial-analyst/
```

### Agent Workspace Files

Each agent has editable files that define its personality and capabilities:

| File | Purpose |
|---|---|
| `SOUL.md` | Personality, values, behavior |
| `IDENTITY.md` | Name, version, capabilities |
| `USER.md` | User preferences |
| `AGENT.md` | Operational rules |
| `skills/` | Custom skill definitions |
| `memory/` | Persistent memory (auto-managed) |

Edit these files and push changes without restarting:

```bash
vim workspace/agent-financiero/SOUL.md
claw-ctl reload finance agent-financiero
```

---

## Secret Management

`claw-ctl` dynamically detects which secrets are needed based on your configuration:

| If you enable... | You'll be asked for... |
|---|---|
| Telegram channel | `TELEGRAM_BOT_TOKEN` |
| Discord channel | `DISCORD_BOT_TOKEN` |
| WhatsApp channel | `WHATSAPP_API_TOKEN` |
| Gemini model | `GEMINI_API_KEY` |
| Ollama model | Nothing — runs locally |

Secrets can be provided via:
- **`.env` file** — `claw-ctl deploy finance --env-file .env`
- **HashiCorp Vault** — `claw-ctl deploy finance --vault-addr https://vault.example.com`
- **Interactive prompt** — the wizard asks for each token securely

---

## Requirements

| Requirement | Required? | Notes |
|---|---|---|
| **LLM Access** | ✅ Yes | API key or Ollama address |
| Kubernetes cluster | Recommended | Uses `~/.kube/config` |
| Vault | Optional | Only with `--vault-addr` |

---

## Architecture

Each agent runs in an isolated [vCluster](https://www.vcluster.com/) with:
- **Crystal Wall RBAC** — agents cannot read Kubernetes secrets
- **Dedicated workspace** — persistent storage + editable config files
- **Multi-agent support** — multiple agents per cluster, each with isolated secrets

---

## License

MIT
