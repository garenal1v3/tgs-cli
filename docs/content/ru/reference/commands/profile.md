---
title: tgs profile
weight: 20
---

# tgs profile

Управление профилями аккаунтов.

## Подкоманды

### tgs profile list

Список всех профилей с информацией об аккаунтах.

```bash
tgs profile list
tgs profile list --output json
```

Активный профиль отмечен звёздочкой (`*`).

### tgs profile switch

Устанавливает активный профиль для текущей директории, записывая `.tgs.yaml`.

```bash
tgs profile switch <name>
```

Пример:

```bash
tgs profile switch work
```

### tgs profile delete

Удаляет профиль и все его данные сессии. Нельзя удалить активный профиль.

```bash
tgs profile delete <name>
```

Пример:

```bash
tgs profile delete old-account
```
