---
title: tgs search counters
weight: 12
---

> **Примечание:** Эта документация может отставать от [английской версии](/en/reference/commands/search-counters/).

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
| `--filters` | | string[] | | Типы фильтров для подсчета (через запятую; по умолчанию: все). См. [допустимые значения](/ru/reference/commands/search-messages/#допустимые-значения-фильтров) |
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

Возвращает JSON с массивом объектов, каждый из которых содержит название типа фильтра и соответствующее количество сообщений.

## Смотрите также

- [tgs search messages]({{< relref "/reference/commands/search-messages" >}})
- [tgs search global]({{< relref "/reference/commands/search-global" >}})
- [tgs search calendar]({{< relref "/reference/commands/search-calendar" >}})
