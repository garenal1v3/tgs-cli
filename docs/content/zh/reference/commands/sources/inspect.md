---
title: tgs sources inspect
weight: 20
---

> **注意：** 本文档可能落后于[英文版本](/en/reference/commands/sources/inspect/)。

# tgs sources inspect

获取单个 Telegram 来源的详细信息。适用于您已订阅的来源，也适用于您未加入的公开频道。

## 用法

```
tgs sources inspect <ref> [flags]
```

`ref` 参数用于标识目标来源，支持以下格式：

| 格式 | 示例 | 说明 |
|---|---|---|
| `@username` | `@durov` | 带 `@` 前缀的公开用户名 |
| `username` | `durov` | 不带前缀的公开用户名 |
| 数字 ID | `-1001234567890` | Telegram 对等体 ID |
| 电话号码 | `+79001234567` | 用于联系人或您自己的账户 |
| `-` 或 `@me` | `-` | 收藏夹（"Saved Messages"） |

## 参数

| 参数 | 缩写 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--no-stats` | | bool | `false` | 跳过扩展统计（总消息数/24h/首条消息） |
| `--no-cache` | | bool | `false` | 禁用对等体和统计缓存 |
| `--max-wait` | | int | `60` | FLOOD_WAIT 最大等待秒数 |
| `--profile` | `-p` | string | | 账户配置文件名称 |

未设置 `--profile` 时的配置文件解析顺序：`TGS_PROFILE` 环境变量 -> `.tgs.yaml` 文件 -> `"default"`。

## 示例

通过用户名查看频道：

```bash
tgs sources inspect @durov
```

通过数字 ID 查看群组：

```bash
tgs sources inspect -1009876543210
```

查看收藏夹：

```bash
tgs sources inspect -
```

查看您未订阅的频道：

```bash
tgs sources inspect @somechannel
```

跳过扩展统计：

```bash
tgs sources inspect @durov --no-stats
```

使用指定账户配置文件：

```bash
tgs sources inspect @golang --profile work
```

## 输出

返回单个 JSON 对象，包含来源的完整信息。`subscribed` 字段表示您的账户是否为该来源的成员。

```json
{
  "id": -1001234567890,
  "type": "channel",
  "title": "Durov's Channel",
  "username": "durov",
  "access": "public",
  "members_count": 1234567,
  "has_comments": true,
  "linked_chat_id": -1009876543210,
  "verified": true,
  "unread_count": 0,
  "last_message": {"id": 4321, "date": "2026-05-28T08:15:00Z"},
  "subscribed": false,
  "description": "Pavel Durov's official channel.",
  "creation_date": "2015-08-26T10:00:00Z",
  "invite_link": "https://t.me/+abc",
  "stats": {
    "total_messages": 12345,
    "messages_24h": 3,
    "first_message": {"id": 1, "date": "2015-08-26T10:00:00Z"}
  }
}
```

对于未订阅的频道，`subscribed` 为 `false`，`stats` 为 `null`。

### 额外字段（仅限 inspect）

| 字段 | 类型 | 说明 |
|---|---|---|
| `subscribed` | bool | 您的账户是否订阅了该来源 |
| `description` | string | 完整描述/关于信息 |
| `creation_date` | string | 聊天创建时间（RFC3339 UTC 时间戳） |
| `invite_link` | string | 主要邀请链接；可用时存在 |

布尔字段（`verified`、`scam`、`fake`、`restricted`、`archived`、`pinned`、`saved`、`deleted`、`has_topics`、`gigagroup`）仅在值为 `true` 时出现。

## 另请参阅

- [tgs sources list]({{< relref "/reference/commands/sources/list" >}})
