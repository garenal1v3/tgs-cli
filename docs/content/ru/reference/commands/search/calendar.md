---
title: tgs search calendar
weight: 40
---

> **Примечание:** Эта документация может отставать от [английской версии](/en/reference/commands/search/calendar/).

# tgs search calendar

Получение результатов поиска сообщений, сгруппированных по дате, для конкретного чата и типа фильтра.

## Использование

```
tgs search calendar [flags]
```

Команда не принимает позиционных аргументов.

## Флаги

| Флаг | Сокр. | Тип | По умолчанию | Описание |
|------|-------|-----|--------------|----------|
| `--chat` | `-c` | string | | Чат для запроса (username, телефон или ID). **Обязательный.** |
| `--filter` | | string | | Фильтр по типу сообщения (см. [допустимые значения](/ru/reference/commands/search/messages/#допустимые-значения-фильтров)). **Обязательный.** |
| `--max-wait` | | int | `60` | Максимум секунд ожидания при FLOOD_WAIT |
| `--no-cache` | | bool | `false` | Отключить кеш разрешения пиров |
| `--profile` | `-p` | string | | Имя профиля аккаунта |

Порядок определения профиля, если `--profile` не указан: переменная `TGS_PROFILE` -> файл `.tgs.yaml` -> `"default"`.

## Примеры

Календарь фотографий в канале:

```bash
tgs search calendar -c @travel_photos --filter photo
```

Календарь документов в группе:

```bash
tgs search calendar -c @dev_team --filter document
```

## Вывод

Возвращает JSON с массивом записей по датам и общим количеством `total`.

**JSON (по умолчанию):**

```json
{
  "periods": [
    {"date": "2026-05-12", "count": 2, "min_msg_id": 510, "max_msg_id": 511},
    {"date": "2026-05-10", "count": 1, "min_msg_id": 508, "max_msg_id": 508},
    {"date": "2025-12-23", "count": 1, "min_msg_id": 466, "max_msg_id": 466}
  ],
  "total": 98
}
```

**Текст (`--output text`):**

```
2026-05-12  2
2026-05-10  1
2025-12-23  1

total: 98
```

Каждая запись охватывает сообщения за указанную дату (UTC). `min_msg_id` и `max_msg_id` позволяют достать конкретные сообщения через `tgs search messages --cursor` или любой сторонний инструмент, принимающий диапазон ID.

## Смотрите также

- [tgs search messages]({{< relref "/reference/commands/search/messages" >}})
- [tgs search global]({{< relref "/reference/commands/search/global" >}})
- [tgs search counters]({{< relref "/reference/commands/search/counters" >}})
