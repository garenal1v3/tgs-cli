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
| `id:<n>` | `id:-1001234567890` | Telegram peer ID (рекомендуется для отрицательных ID — pflag воспринимает голый `-N` как флаг) |
| Числовой ID | `12345` | Положительный peer ID. Для отрицательных используйте форму `id:` выше или вставьте `--` (например `inspect -- -1001234567890`) |
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

Изучить группу по числовому ID (используйте форму `id:`, чтобы CLI не съел минус как флаг):

```bash
tgs sources inspect id:-1009876543210
# или с разделителем `--` (внимание: флаги команды должны идти ДО `--`,
# так как всё после разделителя становится позиционным аргументом):
tgs sources inspect --no-stats -- -1009876543210
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

Для публичных каналов, на которые вы не подписаны, `subscribed` равно `false`, а ключ `stats` полностью отсутствует в JSON. Отображаемые поля (`title`, `username`, `access`, `members_count`, `description`, `verified`, …) при этом заполняются из публичной информации канала.

### Дополнительные поля (только в inspect)

| Поле | Тип | Описание |
|---|---|---|
| `subscribed` | bool | Подписан ли ваш аккаунт на этот источник |
| `description` | string | Полное описание / «О чате» |
| `creation_date` | string | UTC-таймштамп RFC3339 — дата создания чата/канала. Возвращается только для каналов/супергрупп/групп — Telegram не отдаёт дату регистрации пользователей |
| `invite_link` | string | Основная ссылка-приглашение; присутствует, если доступна |

Булевы поля (`verified`, `scam`, `fake`, `restricted`, `archived`, `pinned`, `saved`, `deleted`, `has_topics`, `gigagroup`) присутствуют только если они равны `true`.

### Замечания по флагам

`--no-cache` отключает peer-кеш (BoltDB по пути `$XDG_DATA_HOME/tgs/profiles/<profile>/cache.db`, по умолчанию это `~/.local/share/tgs/profiles/<profile>/cache.db` на Linux/macOS) — обходится и кэш username-резолва, и snapshot отображаемых полей. Рядом лежит `stats_cache.db`, на который опирается `tgs sources list --with-stats`; `inspect` его не использует.

## Смотрите также

- [tgs sources list]({{< relref "/reference/commands/sources/list" >}})
