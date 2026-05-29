---
title: tgs search messages
weight: 10
---

> **Примечание:** Эта документация может отставать от [английской версии](/en/reference/commands/search/messages/).

# tgs search messages

Поиск сообщений в одном или нескольких чатах, каналах или группах.

## Использование

```
tgs search messages [query] [flags]
```

Аргумент `query` обязателен и задает текст для поиска.

## Флаги

| Флаг | Сокр. | Тип | По умолчанию | Описание |
|------|-------|-----|--------------|----------|
| `--chat` | `-c` | string[] | | Чат для поиска (username, телефон, ID; можно указывать несколько раз или через запятую). Обязателен, если не задан `--folder`. |
| `--folder` | | string | `""` | Искать во всех чатах этой папки (id или имя); пиры папки суммируются с `--chat` |
| `--from` | `-f` | string | | Фильтр по отправителю (username, телефон или ID) |
| `--filter` | | string | | Фильтр по типу сообщения (см. [допустимые значения](#допустимые-значения-фильтров)) |
| `--after` | | string | | Только сообщения после даты (YYYY-MM-DD или unix timestamp) |
| `--before` | | string | | Только сообщения до даты (YYYY-MM-DD или unix timestamp) |
| `--topic` | | int | `0` | ID темы форума |
| `--limit` | `-l` | int | `50` | Максимум сообщений в ответе (1-100) |
| `--cursor` | | string | | Курсор пагинации из предыдущего ответа |
| `--max-wait` | | int | `60` | Максимум секунд ожидания при FLOOD_WAIT |
| `--no-cache` | | bool | `false` | Отключить кеш разрешения пиров |
| `--include-comments` | | bool | `false` | Также искать в дискуссионной группе каждого канала |
| `--profile` | `-p` | string | | Имя профиля аккаунта |

Порядок определения профиля, если `--profile` не указан: переменная `TGS_PROFILE` -> файл `.tgs.yaml` -> `"default"`.

### Допустимые значения фильтров

`photo`, `video`, `photo-video`, `document`, `url`, `gif`, `voice`, `music`, `round-video`, `geo`, `contact`, `pinned`, `mention`, `phone-call`, `chat-photo`.

## Примеры

Поиск по ключевому слову в канале:

```bash
tgs search messages "release notes" -c @golang
```

Поиск в нескольких чатах:

```bash
tgs search messages "deployment" -c @devops_team -c @infrastructure
```

Фильтр по отправителю и типу сообщения:

```bash
tgs search messages "config" -c @mygroup --from @alice --filter document
```

Поиск в диапазоне дат:

```bash
tgs search messages "outage" -c @incidents --after 2025-01-01 --before 2025-06-01
```

Поиск по всем чатам в папке:

```bash
tgs search messages "announcement" --folder Work
```

Пагинация результатов:

```bash
tgs search messages "bug" -c @dev -l 10 --cursor "eyJvIjo1MCwiZCI6MH0"
```

## Вывод

Возвращает JSON с массивом найденных сообщений, полем `total` и опциональным `cursor` для пагинации.

**JSON (по умолчанию):**

```json
{
  "messages": [
    {
      "id": 520,
      "chat": {"id": -1001006503122, "type": "channel", "title": "Pavel Durov"},
      "date": "2026-05-23T13:26:07Z",
      "text": "WhatsApp encryption is a giant fraud...",
      "media": {"type": "webpage"},
      "views": 1128733,
      "forwards": 10748,
      "replies": 540,
      "reactions": [
        {"emoji": "👍", "count": 8412},
        {"emoji": "🔥", "count": 2103},
        {"emoji": "❤", "count": 991}
      ]
    }
  ],
  "total": 285,
  "cursor": "eyJvIjo1MjAsImQiOjB9"
}
```

**Текст (`--output text`):**

```
[2026-05-23 13:26:07] Pavel Durov: WhatsApp encryption is a giant fraud... [👁 1128733 ↻ 10748 💬 540 👍 8412 🔥 2103 ❤ 991]
```

### Поля сообщения

| Поле | Тип | Описание |
|---|---|---|
| `id` | int | ID сообщения внутри чата |
| `chat` | object | `{id, type, title?, username?}` — `id` в форме Bot API (отрицательный для `channel`/`supergroup`/`group`, положительный для `private`), поэтому его можно сразу передать в `--chat`; `type` это `channel`, `supergroup`, `group` или `private` |
| `from` | object | Отправитель (отсутствует для постов канала, присутствует для групп и комментариев) |
| `date` | string | UTC-таймштамп в формате RFC3339 |
| `text` | string | Текст сообщения |
| `media` | object | Тип и метаданные вложения; отсутствует для текстовых сообщений |
| `reply_to_msg_id` | int | ID сообщения, на которое отвечают; в discussion-группах указывает на forwarded-пост канала |
| `topic_id` | int | ID темы форума (если применимо) |
| `views`, `forwards`, `replies` | int | Счётчики активности (для постов канала) |
| `reactions` | array | Массив `{emoji, count}`; кастомные эмодзи — как `custom:<doc_id>`, платные — `⭐` |
| `cursor` | string | Курсор пагинации; отсутствует, если результатов больше нет |

## Смотрите также

- [tgs search global]({{< relref "/reference/commands/search/global" >}})
- [tgs search counters]({{< relref "/reference/commands/search/counters" >}})
- [tgs search calendar]({{< relref "/reference/commands/search/calendar" >}})
