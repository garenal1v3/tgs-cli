---
title: Введение
---

# tgs

Поиск по Telegram из терминала

### Возможности

- **Создан для AI-агентов** — JSON по умолчанию, инструмент для Claude Code, Codex и других
- **Быстрый** — единственное узкое место — сам Telegram API
- **Мульти-аккаунт** — переключение между профилями как в AWS CLI
- **Безопасный** — импорт сессии из Telegram Desktop, без хранения паролей
- **Кроссплатформенный** — macOS, Linux, Windows

### Установка

```bash
brew install garenal1v3/tap/tgs-cli
```

Или смотрите [Установку]({{< relref "/getting-started/installation" >}}) для других способов.

### Быстрый старт

```bash
# Аутентификация (импорт сессии Telegram Desktop)
tgs login

# Проверить текущий аккаунт
tgs whoami
```
