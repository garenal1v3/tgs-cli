---
title: tgs sources inspect
weight: 20
---

> **Примечание:** Эта документация может отставать от [английской версии](/en/reference/commands/sources/inspect/).

# tgs sources inspect

Получить подробную информацию об одном источнике Telegram. Работает как для подписанных, так и для публичных каналов, на которые вы не подписаны.

## Использование

```
tgs sources inspect <ref> [flags]
```

Аргумент `ref` идентифицирует целевой источник. Поддерживаемые форматы:

| Формат | Пример | Описание |
|---|---|---|
| `@username` | `@durov` | Публичный username с префиксом `@` |
| `username` | `durov` | Публичный username без префикса |
| Числовой ID | `-1001234567890` | Telegram peer ID |
| Номер телефона | `+79001234567` | Для контактов и собственного аккаунта |
| `-` или `@me` | `-` | «Избранное» ("Saved Messages") |

## Флаги

| Флаг | Сокр. | Тип | По умолчанию | Описание |
|------|-------|-----|--------------|----------|
| `--no-stats` | | bool | `false` | Пропустить расширенную статистику (всего/за 24ч/первое сообщение) |
| `--no-cache` | | bool | `false` | Отключить кеш пиров и статистики |
| `--max-wait` | | int | `60` | Максимум секунд ожидания при FLOOD_WAIT |
| `--profile` | `-p` | string | | Имя профиля аккаунта |

Порядок определения профиля, если `--profile` не указан: переменная `TGS_PROFILE` -> файл `.tgs.yaml` -> `"default"`.

## Примеры

Изучить канал по username:

```bash
tgs sources inspect @durov
```

Изучить группу по числовому ID:

```bash
tgs sources inspect -1009876543210
```

Изучить своё «Избранное»:

```bash
tgs sources inspect -
```

Изучить канал, на который вы не подписаны:

```bash
tgs sources inspect @somechannel
```

Пропустить расширенную статистику:

```bash
tgs sources inspect @durov --no-stats
```

Использовать конкретный профиль аккаунта:

```bash
tgs sources inspect @golang --profile work
```

## Вывод

Возвращает один JSON-объект с полным описанием источника. Поле `subscribed` указывает, является ли ваш аккаунт участником.

```json
{
  "id": -1001234567890,
  "type": "channel",
  "title": "Durov's Channel",
  "username": "durov",
  "access": "public",
  "members_count": 1234567,
  "has_comments": true,
  "linked_chat_id": -1009876543210,
  "verified": true,
  "unread_count": 0,
  "last_message": {"id": 4321, "date": "2026-05-28T08:15:00Z"},
  "subscribed": false,
  "description": "Pavel Durov's official channel.",
  "creation_date": "2015-08-26T10:00:00Z",
  "invite_link": "https://t.me/+abc",
  "stats": {
    "total_messages": 12345,
    "messages_24h": 3,
    "first_message": {"id": 1, "date": "2015-08-26T10:00:00Z"}
  }
}
```

Для каналов, на которые вы не подписаны, `subscribed` равно `false`, а `stats` — `null`.

### Дополнительные поля (только в inspect)

| Поле | Тип | Описание |
|---|---|---|
| `subscribed` | bool | Подписан ли ваш аккаунт на этот источник |
| `description` | string | Полное описание / «О чате» |
| `creation_date` | string | UTC-таймштамп RFC3339 — дата создания чата |
| `invite_link` | string | Основная ссылка-приглашение; присутствует, если доступна |

Булевы поля (`verified`, `scam`, `fake`, `restricted`, `archived`, `pinned`, `saved`, `deleted`, `has_topics`, `gigagroup`) присутствуют только если они равны `true`.

## Смотрите также

- [tgs sources list]({{< relref "/reference/commands/sources/list" >}})
