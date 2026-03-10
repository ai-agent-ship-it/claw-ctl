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
# Basic deploy (no Vault)
claw-ctl deploy finance --preset financial-controller

# With .env file for secrets
claw-ctl deploy finance --preset financial-controller --env-file .env

# With HashiCorp Vault
claw-ctl deploy finance --preset financial-controller --vault
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
  mode: env        # "env", "vault", or "manual"
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
| `ENVIRONMENT.md` | K8s context, security constraints |
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

### Three Secret Modes

#### 1. `.env` File Mode

```bash
claw-ctl deploy finance --preset financial-controller --env-file .env
```

Creates native Kubernetes Secrets from your `.env` file.

#### 2. HashiCorp Vault Mode

```bash
# From env vars
export VAULT_ADDR="https://vault.example.com"
export VAULT_TOKEN="hvs.xxxx"
claw-ctl deploy finance --preset financial-controller --vault

# From .env file
claw-ctl deploy finance --preset financial-controller --vault --env-file .env

# From flags
claw-ctl deploy finance --preset financial-controller --vault \
  --vault-addr https://vault.example.com --vault-token hvs.xxxx
```

When using `--vault`, claw-ctl automatically:
1. **Creates KV v2 secrets** at `secret/agents/<cluster>/<agent>` (idempotent — won't overwrite existing)
2. **Creates ACL policies** with read-only access per agent
3. **Creates K8s auth roles** bound to the agent's ServiceAccount
4. **Applies VSO CRDs** — `VaultConnection`, `VaultAuth`, `VaultStaticSecret`
5. **VSO syncs secrets** from Vault into K8s Secrets automatically

Resolution priority: **flags → env vars → .env file**

| Source | Variables |
|---|---|
| Flags | `--vault-addr`, `--vault-token` |
| Environment | `VAULT_ADDR`, `VAULT_TOKEN` |
| .env file | `VAULT_ADDR=...`, `VAULT_TOKEN=...` |

#### 3. Manual Mode

Skips automatic secret provisioning. You manage secrets yourself.

---

## Idempotency

All commands are idempotent — safe to run multiple times:

| Operation | Behavior on re-run |
|---|---|
| Namespace creation | Skips if exists |
| vCluster creation | Reuses existing |
| vCluster deletion | OK if not found |
| K8s Secrets | Create-or-update (upsert) |
| Vault KV write | Skips if path exists (won't overwrite real secrets) |
| Vault policies/roles | Overwrites (idempotent) |
| VSO CRDs | `kubectl apply` (idempotent) |
| Agent manifests | `kubectl apply` via `vcluster connect` |

---

## Requirements

| Requirement | Required? | Notes |
|---|---|---|
| **LLM Access** | ✅ Yes | API key or Ollama address |
| Kubernetes cluster | Recommended | Uses `~/.kube/config` |
| `vcluster` CLI | ✅ Yes | [Install](https://www.vcluster.com/docs/getting-started/setup) |
| `kubectl` | ✅ Yes | For manifest operations |
| Vault | Optional | Only with `--vault` flag |
| Vault Secrets Operator | Optional | Required for auto-sync with `--vault` |

---

## Architecture

Each agent runs in an isolated [vCluster](https://www.vcluster.com/) with:
- **Crystal Wall RBAC** — agents cannot read Kubernetes secrets
- **Dedicated workspace** — persistent storage + editable config files via ConfigMaps
- **Multi-agent support** — multiple agents per cluster, each with isolated secrets
- **Vault integration** — VSO syncs secrets from host namespace into vCluster
- **fromHost secret sync** — enabled automatically when using `--vault`

### Deployment Flow

```
Phase 1: Infrastructure → K8s connectivity + namespace creation
Phase 2: vCluster      → Create + wait ready (+ fromHost secret sync if Vault)
Phase 3: Secrets        → Vault API provisioning + VSO CRDs (or .env → K8s Secrets)
Phase 4: Manifests      → Apply via `vcluster connect -- kubectl apply`
```

---

## License

MIT
