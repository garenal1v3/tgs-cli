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

## Смотрите также

- [tgs logout]({{< relref "/reference/commands/logout" >}})
- [Руководство по аутентификации]({{< relref "/getting-started/authentication" >}})
