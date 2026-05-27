---
title: 介绍
---

# tgs

从终端搜索 Telegram

### 功能特点

- **专为 AI 代理设计** — 默认 JSON 输出，适用于 Claude Code、Codex 等工具
- **快速** — 唯一瓶颈是 Telegram API 本身
- **多账户** — 像 AWS CLI 一样切换配置文件
- **安全** — 从 Telegram Desktop 导入会话，不存储密码
- **跨平台** — macOS、Linux、Windows

### 安装

```bash
brew install garenal1v3/tap/tgs
```

或查看[安装指南]({{< relref "/getting-started" >}})了解其他方式。

### 快速开始

```bash
# 认证（导入 Telegram Desktop 会话）
tgs login

# 查看当前账户
tgs whoami
```
