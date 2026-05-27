---
title: tgs logout
weight: 11
---

# tgs logout

Отзывает текущую сессию на серверах Telegram и удаляет локальные данные сессии.

## Синопсис

{{< snippet "cmd-logout/synopsis.md" >}}

## Флаги

{{< snippet "cmd-logout/flags.md" >}}

Порядок определения профиля, если `--profile` не указан: переменная `TGS_PROFILE` -> файл `.tgs.yaml` -> `"default"`.

## Примеры

{{< snippet "cmd-logout/examples.md" >}}

## Примечания

Выход выполняет два шага:

1. Вызывает API Telegram для отзыва токена сессии (сессия становится недействительной на стороне Telegram)
2. Очищает локальную базу данных сессии для профиля

После выхода необходимо выполнить `tgs login` заново, чтобы использовать tgs с этим профилем.

## Смотрите также

- [tgs login]({{< relref "/reference/commands/login" >}})
- [Руководство по аутентификации]({{< relref "/getting-started/authentication" >}})
