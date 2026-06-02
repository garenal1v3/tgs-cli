---
title: Introduction
---

# tgs

Telegram search from your terminal

### Features

- **Built for AI agents** — JSON output by default, designed as a tool for Claude Code, Codex, and other AI assistants
- **Fast** — the only bottleneck is Telegram API itself
- **Multi-account** — switch between profiles like AWS CLI
- **Secure** — imports sessions from Telegram Desktop, no password storage
- **Cross-platform** — macOS, Linux, Windows

### Install

```bash
brew install garenal1v3/tap/tgs-cli
```

Or see [Installation]({{< relref "/getting-started/installation" >}}) for other methods.

### Quick Start

```bash
# Authenticate (imports Telegram Desktop session)
tgs login

# Check who you are
tgs whoami
```
