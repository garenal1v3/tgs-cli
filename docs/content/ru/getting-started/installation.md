---
title: Установка
weight: 5
---

# Установка

`tgs` поставляется как один статический бинарник без зависимостей. Выберите способ под вашу платформу — установленная команда в любом случае называется `tgs`.

## Homebrew (macOS и Linux)

Самый быстрый способ на macOS и Linux:

{{< snippet "install/homebrew.md" >}}

Команда устанавливает бинарник `tgs` и обновляет его через `brew upgrade`.

## Готовые бинарники

Скачайте архив релиза под вашу платформу со [страницы релизов](https://github.com/garenal1v3/tgs-cli/releases). Архивы называются `tgs-cli_<os>_<arch>`:

| Платформа | OS | Arch |
|---|---|---|
| macOS (Apple Silicon) | `darwin` | `arm64` |
| macOS (Intel) | `darwin` | `amd64` |
| Linux | `linux` | `amd64` / `arm64` |
| Windows | `windows` | `amd64` / `arm64` |

### macOS и Linux

{{< snippet "install/binary-unix.md" >}}

### Windows

{{< snippet "install/binary-windows.md" >}}

## Сборка из исходников

Требуется Go 1.26 или новее:

{{< snippet "install/source.md" >}}

## Проверка установки

{{< snippet "install/verify.md" >}}

В ответ должен прийти JSON с версией, коммитом и датой сборки. Дальше — [Быстрый старт]({{< relref "/getting-started/quick-start" >}}).
