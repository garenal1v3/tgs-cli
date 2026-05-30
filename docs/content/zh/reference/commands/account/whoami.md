---
title: tgs whoami
weight: 20
---

# tgs whoami

显示当前配置文件名称和账户信息。从本地存储读取——无需连接 Telegram。

## 用法

{{< snippet "cmd-whoami/synopsis.md" >}}

## 参数

{{< snippet "cmd-whoami/flags.md" >}}

未设置 `--profile` 时的配置文件解析顺序：`TGS_PROFILE` 环境变量 → `.tgs.yaml` 文件 → `"default"`。

## 示例

{{< snippet "cmd-whoami/examples.md" >}}

## 输出

**文本输出 (`--output text`)：**

{{< snippet "cmd-whoami/output.md" >}}

如果配置文件存在但未登录，`user` 字段在 JSON 中为 `null`，在文本中显示为 `"Not logged in"`。

## 另请参阅

- [tgs profile]({{< relref "/reference/commands/account/profile" >}})
- [配置文件指南]({{< relref "/guide/profiles" >}})
