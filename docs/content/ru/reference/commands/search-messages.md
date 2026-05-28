---
title: tgs search messages
weight: 10
---

> **Примечание:** Эта документация может отставать от [английской версии](/en/reference/commands/search-messages/).

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
| `--chat` | `-c` | string[] | | Чат для поиска (username, телефон, ID; можно указывать несколько раз или через запятую). **Обязательный.** |
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

Пагинация результатов:

```bash
tgs search messages "bug" -c @dev -l 10 --cursor "eyJvIjo1MCwiZCI6MH0"
```

## Вывод

Возвращает JSON с массивом найденных сообщений и полем `cursor` для пагинации. Каждое сообщение содержит отправителя, дату, текст, информацию о чате и ID сообщения.

## Смотрите также

- [tgs search global]({{< relref "/reference/commands/search-global" >}})
- [tgs search counters]({{< relref "/reference/commands/search-counters" >}})
- [tgs search calendar]({{< relref "/reference/commands/search-calendar" >}})
