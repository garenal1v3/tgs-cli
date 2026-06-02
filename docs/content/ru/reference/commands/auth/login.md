---
title: tgs login
weight: 10
---

# tgs login

Аутентификация в Telegram. Сессия сохраняется локально и используется всеми последующими командами.

## Синопсис

{{< snippet "cmd-login/synopsis.md" >}}

## Флаги

{{< snippet "cmd-login/flags.md" >}}

Порядок определения профиля, если `--profile` не указан: переменная `TGS_PROFILE` -> файл `.tgs.yaml` -> `"default"`.

## Примеры

{{< snippet "cmd-login/examples.md" >}}

## Методы аутентификации

{{< snippet "cmd-login/methods.md" >}}

## Вывод

При успешном входе:

**JSON (по умолчанию):**

```json
{"profile":"default","status":"logged_in","user":{"id":261054642,"phone":"+79001234567","username":"alice","first_name":"Alice","last_name":"Doe"}}
```

**Текст (`--output text`):**

```
Logged in as Alice Doe (id: 261054642, profile: default)
```

В двухшаговом флоу `code` первый вызов (без `--code`) отправляет код подтверждения:

```json
{"profile":"default","status":"code_sent"}
```

```
Verification code sent (profile: default). Re-run with --code to complete login.
```

При ошибке команда завершается ненулевым кодом и пишет ошибку в stderr. Типичные причины: неверный код, истёкший QR-токен, неверный 2FA-пароль, или сессия уже существует (нужно сначала `tgs logout`).

## Смотрите также

- [tgs logout]({{< relref "/reference/commands/auth/logout" >}})
- [Руководство по аутентификации]({{< relref "/getting-started/authentication" >}})
