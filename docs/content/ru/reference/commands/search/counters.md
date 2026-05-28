---
title: tgs search counters
weight: 30
---

> **Примечание:** Эта документация может отставать от [английской версии](/en/reference/commands/search/counters/).

# tgs search counters

Получение количества сообщений по типам (фото, видео, документы и т.д.) для конкретного чата.

## Использование

```
tgs search counters [flags]
```

Команда не принимает позиционных аргументов.

## Флаги

| Флаг | Сокр. | Тип | По умолчанию | Описание |
|------|-------|-----|--------------|----------|
| `--chat` | `-c` | string | | Чат для запроса (username, телефон или ID). **Обязательный.** |
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

## Вывод

Возвращает JSON с массивом `{filter, count}`.

**JSON (по умолчанию):**

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
