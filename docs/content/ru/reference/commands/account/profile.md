---
title: tgs profile
weight: 10
---

# tgs profile

Управление профилями аккаунтов.

## Синопсис

{{< snippet "cmd-profile/synopsis.md" >}}

## Подкоманды

### tgs profile list

Список всех профилей с информацией об аккаунтах.

{{< snippet "cmd-profile/list.md" >}}

**JSON (по умолчанию):**

```json
[
  {"name":"default","phone":"+79001234567","username":"alice","id":261054642,"active":true},
  {"name":"work","phone":"+79001230000","username":"alice_work","id":342178890,"active":false}
]
```

**Текст (`--output text`):**

```
* default (@alice)
  work (@alice_work)
```

Активный профиль отмечен `*` в текстовом виде и `"active": true` в JSON. Если ни username, ни телефон не закешированы, выводится только имя профиля.

---

### tgs profile switch

Устанавливает активный профиль для текущей директории, записывая файл `.tgs.yaml`.

{{< snippet "cmd-profile/switch.md" >}}

**JSON (по умолчанию):**

```json
{"profile":"work","status":"switched","dir":"/Users/alice/projects"}
```

**Текст (`--output text`):**

```
Switched to profile "work" (wrote .tgs.yaml in /Users/alice/projects)
```

Все команды tgs, выполняемые в этой директории (и вложенных), будут использовать указанный профиль.

---

### tgs profile delete

Удаляет профиль и его локальную базу данных сессии.

{{< snippet "cmd-profile/delete.md" >}}

**JSON (по умолчанию):**

```json
{"profile":"old","status":"deleted"}
```

**Текст (`--output text`):**

```
Deleted profile "old"
```

Нельзя удалить текущий активный профиль. Сначала переключитесь на другой.

## Смотрите также

- [tgs whoami]({{< relref "/reference/commands/account/whoami" >}})
- [Руководство по профилям]({{< relref "/guide/profiles" >}})
