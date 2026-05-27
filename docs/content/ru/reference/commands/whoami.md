---
title: tgs whoami
weight: 25
---

# tgs whoami

Показывает текущий профиль и информацию об аккаунте. Читает данные из локального хранилища — подключение к Telegram не требуется.

## Синопсис

{{< snippet "cmd-whoami/synopsis.md" >}}

## Флаги

{{< snippet "cmd-whoami/flags.md" >}}

Порядок определения профиля, если `--profile` не указан: переменная `TGS_PROFILE` -> файл `.tgs.yaml` -> `"default"`.

## Примеры

{{< snippet "cmd-whoami/examples.md" >}}

## Вывод

{{< snippet "cmd-whoami/output.md" >}}

Если профиль существует, но вход не выполнен, поле `user` будет `null` (JSON) или `"Not logged in"` (текст).

## Смотрите также

- [tgs profile]({{< relref "/reference/commands/profile" >}})
- [Руководство по профилям]({{< relref "/guide/profiles" >}})
