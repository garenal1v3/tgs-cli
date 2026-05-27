---
title: Аутентификация
weight: 10
---

# Аутентификация

tgs подключается к Telegram через ваш реальный аккаунт по протоколу MTProto. Аутентификация нужна один раз для каждого профиля.

## Быстрый старт

Импорт сессии из Telegram Desktop (метод по умолчанию):

```bash
tgs login
```

Или выберите другой метод:

```bash
tgs login --type code   # Номер телефона + код из SMS/Telegram
tgs login --type qr     # Сканирование QR-кода с другого устройства
```

## Методы входа

### Импорт из Desktop (по умолчанию)

Импортирует существующую сессию из Telegram Desktop. tgs автоматически определяет директорию с данными Telegram Desktop.

```bash
tgs login                               # автоопределение
tgs login --desktop-dir /path/to/tdata  # указать путь вручную
tgs login --passcode SECRET             # если клиент защищён паролем
```

### Телефон + код

Интерактивный вход с номером телефона. Поддерживает двухфакторную аутентификацию (облачный пароль).

```bash
tgs login --type code
tgs login --type code --phone +71234567890
```

### QR-код

Отображает QR-код в терминале. Отсканируйте его в Telegram на телефоне.

```bash
tgs login --type qr
```

## Вход в конкретный профиль

```bash
tgs login --profile work
tgs login --type code --profile personal
```

Подробнее о мульти-аккаунте — в разделе [Профили](/guide/profiles/).

## Выход

```bash
tgs logout                  # выйти из текущего профиля
tgs logout --profile work   # выйти из конкретного профиля
```
