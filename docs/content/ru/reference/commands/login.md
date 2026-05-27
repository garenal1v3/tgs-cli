---
title: tgs login
weight: 10
---

# tgs login

Аутентификация в Telegram.

## Использование

```
tgs login [flags]
```

## Флаги

| Флаг | Краткий | По умолчанию | Описание |
|------|---------|--------------|----------|
| `--type` | `-T` | `desktop` | Метод входа: `desktop`, `code`, `qr` |
| `--desktop-dir` | `-d` | (автоопределение) | Путь к директории tdata Telegram Desktop |
| `--passcode` | | | Пароль Telegram Desktop |
| `--phone` | | | Номер телефона (для метода `code`) |
| `--profile` | `-p` | (определяется автоматически) | Профиль для входа |

## Примеры

```bash
# Импорт из Telegram Desktop
tgs login

# Телефон + код
tgs login --type code --phone +71234567890

# QR-код
tgs login --type qr

# Вход в конкретный профиль
tgs login --profile work --type code

# Telegram Desktop с кастомным путём и паролем
tgs login --desktop-dir ~/snap/telegram-desktop/current/.local/share/TelegramDesktop/tdata
tgs login --passcode SECRET
```
