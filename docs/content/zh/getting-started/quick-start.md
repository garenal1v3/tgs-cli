---
title: 快速开始
weight: 8
---

# 快速开始

五条命令，从零到第一次搜索。本节假设你已安装 tgs——如果还没有，请参阅[安装]({{< relref "/getting-started/installation" >}})。

## 1. 身份验证

默认登录方式会从 Telegram Desktop 导入现有会话——无需手机号或验证码：

{{< snippet "auth/quick-start.md" >}}

其他方式（手机号 + 验证码、二维码、面向 AI 代理的非交互流程）详见[身份验证]({{< relref "/getting-started/authentication" >}})。

## 2. 确认当前账户

{{< snippet "auth/verify.md" >}}

以 JSON 返回当前活跃账户——你的 ID、名称和用户名。

## 3. 列出你的聊天

查看账户中的所有对话——频道、群组、用户、机器人：

{{< snippet "quick-start/sources.md" >}}

加上 `--type channel` 可按类型过滤，或用 `--with-stats` 获取消息统计。详见[来源]({{< relref "/guide/sources" >}})。

## 4. 在某个聊天中搜索

{{< snippet "quick-start/search-chat.md" >}}

`-c` 接受用户名、`t.me` 链接、手机号或数字 ID。用 `--from`、`--filter` 以及 `--after` / `--before` 缩小结果范围。

## 5. 全局搜索

不确定在哪个聊天里？一次搜索全部：

{{< snippet "quick-start/search-global.md" >}}

这就是核心流程。每条命令都输出 JSON——可以管道传给 `jq` 或交给 AI 代理。深入了解请看[搜索指南]({{< relref "/guide/search" >}})。
