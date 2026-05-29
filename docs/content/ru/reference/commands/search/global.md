---
title: tgs search global
weight: 20
---

> **Примечание:** Эта документация может отставать от [английской версии](/en/reference/commands/search/global/).

# tgs search global

Глобальный поиск сообщений по всем чатам, каналам и группам.

## Использование

```
tgs search global [query] [flags]
```

Аргумент `query` обязателен и задает текст для поиска.

## Флаги

| Флаг | Сокр. | Тип | По умолчанию | Описание |
|------|-------|-----|--------------|----------|
| `--channels-only` | | bool | `false` | Искать только в каналах; несовместим с `--folder` |
| `--groups-only` | | bool | `false` | Искать только в группах; несовместим с `--folder` |
| `--users-only` | | bool | `false` | Искать только в личных чатах; несовместим с `--folder` |
| `--archived` | | bool | `false` | Искать в архиве (папка ID 1); игнорируется при `--folder` |
| `--folder` | | string | `""` | Искать только в этой папке (id или имя), включая архивные чаты; несовместим с `--channels-only`, `--groups-only`, `--users-only` |
| `--filter` | | string | | Фильтр по типу сообщения (см. [допустимые значения](/ru/reference/commands/search/messages/#допустимые-значения-фильтров)) |
| `--after` | | string | | Только сообщения после даты (YYYY-MM-DD или unix timestamp) |
| `--before` | | string | | Только сообщения до даты (YYYY-MM-DD или unix timestamp) |
| `--limit` | `-l` | int | `50` | Максимум сообщений в ответе (1-100) |
| `--cursor` | | string | | Курсор пагинации из предыдущего ответа |
| `--max-wait` | | int | `60` | Максимум секунд ожидания при FLOOD_WAIT |
| `--profile` | `-p` | string | | Имя профиля аккаунта |

Порядок определения профиля, если `--profile` не указан: переменная `TGS_PROFILE` -> файл `.tgs.yaml` -> `"default"`.

## Примеры

Поиск по всем чатам:

```bash
tgs search global "kubernetes"
```

Поиск только в каналах:

```bash
tgs search global "breaking news" --channels-only
```

Поиск с фильтром по дате и ограничением результатов:

```bash
tgs search global "quarterly report" --after 2025-04-01 -l 20
```

Поиск в архиве:

```bash
tgs search global "old project" --archived
```

Поиск в именованной папке:

```bash
tgs search global "project update" --folder Work
```

## Заметки

При использовании `--folder` команда разворачивает список чатов папки и выполняет
веерный запрос (fan-out) — по одному запросу `searchGlobal` на каждый чат с
последующим слиянием результатов. Это необходимо, поскольку Telegram API
`searchGlobal` не принимает список пиров напрямую.

`--folder` нельзя сочетать с `--channels-only`, `--groups-only` и `--users-only`.
Список чатов папки всегда включает её архивные чаты, поэтому `--archived` не
влияет на результат при заданном `--folder` и молча игнорируется.

## Вывод

Та же структура, что у [tgs search messages]({{< relref "/reference/commands/search/messages" >}}#вывод) — массив `messages`, `total` и опциональный `cursor`. Поскольку результаты охватывают много чатов, поле `chat` в каждом сообщении говорит, откуда оно.

**JSON (по умолчанию):**

```json
{
  "messages": [
    {
      "id": 188791,
      "chat": {"id": -1001754252633, "type": "channel", "title": "News", "username": "newschannel"},
      "date": "2026-05-27T10:51:40Z",
      "text": "...",
      "views": 360330,
      "forwards": 283
    },
    {
      "id": 67578,
      "chat": {"id": -1001069896405, "type": "supergroup", "title": "Tech Talk", "username": "techtalk"},
      "from": {"id": 1978176, "first_name": "Ilya", "username": "valkin"},
      "date": "2026-05-27T10:14:00Z",
      "text": "..."
    }
  ],
  "total": 2066,
  "cursor": "eyJvIjo0NzgxNSwiZCI6MCwiciI6MTc3OTg3MDc1Mn0"
}
```

Замечание: курсор глобального поиска содержит дополнительное поле `r` (rate) — передавайте его в `--cursor` без изменений.

## Смотрите также

- [tgs search messages]({{< relref "/reference/commands/search/messages" >}})
- [tgs search counters]({{< relref "/reference/commands/search/counters" >}})
- [tgs search calendar]({{< relref "/reference/commands/search/calendar" >}})
