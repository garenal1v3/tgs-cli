---
title: tgs search global
weight: 11
---

> **Примечание:** Эта документация может отставать от [английской версии](/en/reference/commands/search-global/).

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
| `--channels-only` | | bool | `false` | Искать только в каналах |
| `--groups-only` | | bool | `false` | Искать только в группах |
| `--users-only` | | bool | `false` | Искать только в личных чатах |
| `--folder` | | int | `0` | Искать только в папке с указанным ID |
| `--filter` | | string | | Фильтр по типу сообщения (см. [допустимые значения](/ru/reference/commands/search-messages/#допустимые-значения-фильтров)) |
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

Поиск в определенной папке:

```bash
tgs search global "project update" --folder 3
```

## Вывод

Возвращает JSON с массивом найденных сообщений и полем `cursor` для пагинации. Каждое сообщение содержит отправителя, дату, текст, информацию о чате и ID сообщения.

## Смотрите также

- [tgs search messages]({{< relref "/reference/commands/search-messages" >}})
- [tgs search counters]({{< relref "/reference/commands/search-counters" >}})
- [tgs search calendar]({{< relref "/reference/commands/search-calendar" >}})
