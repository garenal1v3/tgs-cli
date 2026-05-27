---
title: tgs profile
weight: 20
---

# tgs profile

Управление профилями аккаунтов.

## Синопсис

{{< snippet "cmd-profile/synopsis.md" >}}

## Подкоманды

### tgs profile list

Список всех профилей с информацией об аккаунтах.

{{< snippet "cmd-profile/list.md" >}}

Активный профиль в текстовом выводе отмечен символом `*`.

В формате JSON каждый профиль содержит поле `active` типа boolean.

---

### tgs profile switch

Устанавливает активный профиль для текущей директории, записывая файл `.tgs.yaml`.

{{< snippet "cmd-profile/switch.md" >}}

Все команды tgs, выполняемые в этой директории (и вложенных), будут использовать указанный профиль.

---

### tgs profile delete

Удаляет профиль и его локальную базу данных сессии.

{{< snippet "cmd-profile/delete.md" >}}

Нельзя удалить текущий активный профиль. Сначала переключитесь на другой:

## Смотрите также

- [tgs whoami]({{< relref "/reference/commands/whoami" >}})
- [Руководство по профилям]({{< relref "/guide/profiles" >}})
