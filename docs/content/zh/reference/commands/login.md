---
title: tgs login
weight: 10
---

# tgs login

对 Telegram 进行身份验证。会话保存在本地，后续所有命令自动复用。

## 用法

{{< snippet "cmd-login/synopsis.md" >}}

## 参数

{{< snippet "cmd-login/flags.md" >}}

未设置 `--profile` 时的配置文件解析顺序：`TGS_PROFILE` 环境变量 → `.tgs.yaml` 文件 → `"default"`。

## 示例

{{< snippet "cmd-login/examples.md" >}}

## 认证方式

{{< snippet "cmd-login/methods.md" >}}

## 另请参阅

- [tgs logout]({{< relref "/reference/commands/logout" >}})
- [身份验证指南]({{< relref "/getting-started/authentication" >}})
