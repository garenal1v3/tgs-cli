---
title: tgs search counters
weight: 30
---

> **Примечание:** Эта документация может отставать от [английской версии](/en/reference/commands/search/counters/).

# tgs search counters

Получение количества сообщений по типам (фото, видео, документы и т.д.) для одного или нескольких чатов.

## Использование

```
tgs search counters [flags]
```

Команда не принимает позиционных аргументов.

## Флаги

| Флаг | Сокр. | Тип | По умолчанию | Описание |
|------|-------|-----|--------------|----------|
| `--chat` | `-c` | string | | Чат для запроса (username, телефон или ID). Обязателен, если не задан `--folder`. |
| `--folder` | | string | `""` | Запросить все чаты в этой папке (id или имя); вывод включает разбивку по чатам и агрегированные итоги |
| `--topic` | | int | `0` | ID темы форума |
| `--filters` | | string[] | | Типы фильтров для подсчета (через запятую; по умолчанию: все). См. [допустимые значения](/ru/reference/commands/search/messages/#допустимые-значения-фильтров) |
| `--max-wait` | | int | `60` | Максимум секунд ожидания при FLOOD_WAIT |
| `--no-cache` | | bool | `false` | Отключить кеш разрешения пиров |
| `--profile` | `-p` | string | | Имя профиля аккаунта |

Порядок определения профиля, если `--profile` не указан: переменная `TGS_PROFILE` -> файл `.tgs.yaml` -> `"default"`.

## Примеры

Получить количество сообщений всех типов в канале:

```bash
tgs search counters -c @golang
```

Получить количество для определенных типов:

```bash
tgs search counters -c @mygroup --filters photo,video,document
```

Счётчики для всех чатов в папке:

```bash
tgs search counters --folder Work --filters photo,document
```

## Вывод

Возвращает JSON с массивом `{filter, count}`.

**JSON (по умолчанию, один чат):**

```json
{
  "counters": [
    {"filter": "photo", "count": 98},
    {"filter": "video", "count": 45},
    {"filter": "photo-video", "count": 143},
    {"filter": "document", "count": 0},
    {"filter": "url", "count": 188},
    {"filter": "gif", "count": 10},
    {"filter": "voice", "count": 0},
    {"filter": "music", "count": 0},
    {"filter": "round-video", "count": 0},
    {"filter": "geo", "count": 0}
  ]
}
```

**JSON (с `--folder`, несколько чатов):**

При использовании `--folder` вывод содержит разбивку по чатам и агрегированные итоги:

```json
{
  "chats": [
    {
      "chat": {"id": 1006503122, "type": "channel", "title": "Dev News"},
      "counters": [
        {"filter": "photo", "count": 30},
        {"filter": "document", "count": 12}
      ]
    },
    {
      "chat": {"id": 1009876543, "type": "supergroup", "title": "Team"},
      "counters": [
        {"filter": "photo", "count": 68},
        {"filter": "document", "count": 41}
      ]
    }
  ],
  "totals": [
    {"filter": "photo", "count": 98},
    {"filter": "document", "count": 53}
  ]
}
```

**Текст (`--output text`):**

```
photo        98
video        45
photo-video  143
document     0
url          188
gif          10
voice        0
music        0
round-video  0
geo          0
```

Telegram возвращает только те типы фильтров, которые поддерживает данный чат — каналы обычно не отдают `contact`, `pinned`, `mention`, `phone-call`, `chat-photo`.

## Смотрите также

- [tgs search messages]({{< relref "/reference/commands/search/messages" >}})
- [tgs search global]({{< relref "/reference/commands/search/global" >}})
- [tgs search calendar]({{< relref "/reference/commands/search/calendar" >}})
