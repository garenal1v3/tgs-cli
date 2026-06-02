---
title: 安装
weight: 5
---

# 安装

`tgs` 以单个静态二进制文件分发，无任何运行时依赖。选择适合你平台的方式即可——安装后的命令始终是 `tgs`。

## Homebrew（macOS 和 Linux）

在 macOS 和 Linux 上最快的方式：

{{< snippet "install/homebrew.md" >}}

该命令会安装 `tgs` 二进制文件，并可通过 `brew upgrade` 保持更新。

## 预编译二进制文件

从[发布页面](https://github.com/garenal1v3/tgs-cli/releases)下载对应平台的归档文件。归档命名为 `tgs-cli_<os>_<arch>`：

| 平台 | OS | Arch |
|---|---|---|
| macOS（Apple 芯片） | `darwin` | `arm64` |
| macOS（Intel） | `darwin` | `amd64` |
| Linux | `linux` | `amd64` / `arm64` |
| Windows | `windows` | `amd64` / `arm64` |

### macOS 和 Linux

{{< snippet "install/binary-unix.md" >}}

### Windows

{{< snippet "install/binary-windows.md" >}}

## 从源码构建

需要 Go 1.26 或更高版本：

{{< snippet "install/source.md" >}}

## 验证安装

{{< snippet "install/verify.md" >}}

你应该会看到一个包含版本、提交和构建日期的 JSON 对象。接下来请阅读[快速开始]({{< relref "/getting-started/quick-start" >}})。
