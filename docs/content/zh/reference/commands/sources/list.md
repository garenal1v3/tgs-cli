---
title: tgs sources list
weight: 10
---

> **注意：** 本文档可能落后于[英文版本](/en/reference/commands/sources/list/)。

# tgs sources list

枚举您账户中的所有 Telegram 对话——频道、超级群组、普通群组、用户和机器人。

## 用法

```
tgs sources list [flags]
```

## 参数

| 参数 | 缩写 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--type` | | string[] | （全部） | 按类型筛选：`channel`、`supergroup`、`group`、`user`、`bot`（可重复使用，支持逗号分隔） |
| `--with-stats` | | bool | `false` | 为每个来源获取扩展统计信息（总消息数/24h/首条消息 + 完整信息） |
| `--limit` | `-l` | int | `0` | 最大返回记录数（1-500，0=全部） |
| `--cursor` | | string | | 上次响应中的分页游标 |
| `--archived` | | bool | `false` | 包含已归档的对话 |
| `--max-wait` | | int | `60` | FLOOD_WAIT 最大等待秒数 |
| `--no-cache` | | bool | `false` | 禁用对等体和统计缓存 |
| `--profile` | `-p` | string | | 账户配置文件名称 |

未设置 `--profile` 时的配置文件解析顺序：`TGS_PROFILE` 环境变量 -> `.tgs.yaml` 文件 -> `"default"`。

## 示例

列出所有对话：

```bash
tgs sources list
```

仅列出频道和超级群组：

```bash
tgs sources list --type channel,supergroup
```

获取每个来源的统计信息（较慢——每个来源约 4 次 API 调用）：

```bash
tgs sources list --with-stats
```

包含已归档对话：

```bash
tgs sources list --archived
```

分页遍历大型账户：

```bash
# 第一页
tgs sources list --limit 50

# 使用上次响应中的游标获取下一页
tgs sources list --limit 50 --cursor "eyJvIjo1MCwiZCI6MH0"
```

## 输出

返回 JSON 对象，包含 `sources` 数组、`total` 总数和用于分页的 `cursor`。

```json
{
  "sources": [
    {
      "id": -1001234567890,
      "type": "channel",
      "title": "Durov's Channel",
      "username": "durov",
      "access": "public",
      "members_count": 1234567,
      "verified": true,
      "unread_count": 0,
      "last_message": {"id": 4321, "date": "2026-05-28T08:15:00Z"}
    },
    {
      "id": -1009876543210,
      "type": "supergroup",
      "title": "Go Programming",
      "username": "golang",
      "access": "public",
      "members_count": 78432,
      "has_topics": true,
      "unread_count": 5,
      "last_message": {"id": 120450, "date": "2026-05-28T07:42:11Z"}
    }
  ],
  "total": 287,
  "cursor": "eyJvIjo1MCwiZCI6MH0"
}
```

使用 `--with-stats` 时，每个来源还包含 `stats` 对象：

```json
{
  "stats": {
    "total_messages": 12345,
    "messages_24h": 3,
    "first_message": {"id": 1, "date": "2015-08-26T10:00:00Z"}
  }
}
```

### 来源字段

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | int | Telegram 对等体 ID（频道/群组为负数） |
| `type` | string | 以下之一：`channel`、`supergroup`、`group`、`user`、`bot` |
| `title` | string | 显示名称 |
| `username` | string | 不含 `@` 的公开用户名；未设置时省略 |
| `access` | string | `public` 或 `private`；用户类型省略 |
| `members_count` | int | 成员/订阅者数量；不可用时省略 |
| `description` | string | 简介或关于信息；仅在完整信息下存在 |
| `has_comments` | bool | 频道有关联讨论组；仅在为 `true` 时出现 |
| `linked_chat_id` | int | 关联讨论组的 ID；仅在适用时出现 |
| `creation_date` | string | 频道/群组创建时间的 UTC RFC3339 时间戳；仅对频道/群组在 `--with-stats` 或 `inspect` 时返回 |
| `invite_link` | string | 主要邀请链接；可用时存在 |
| `first_name` | string | 用户/机器人名；频道/群组省略 |
| `last_name` | string | 用户姓氏；未设置时省略 |
| `phone` | string | E.164 格式号码（无 `+`）；非联系人可见时省略 |
| `verified` | bool | Telegram 官方认证账户；仅在为 `true` 时出现 |
| `scam` | bool | 被 Telegram 标记为诈骗；仅在为 `true` 时出现 |
| `fake` | bool | 被 Telegram 标记为虚假；仅在为 `true` 时出现 |
| `restricted` | bool | 在某些地区受限；仅在为 `true` 时出现 |
| `restricted_reason` | string | 自由格式的限制原因；不存在时省略 |
| `deleted` | bool | 已删除的用户账户；仅在为 `true` 时出现 |
| `archived` | bool | 对话已归档；仅在为 `true` 时出现 |
| `pinned` | bool | 对话已置顶；仅在为 `true` 时出现 |
| `saved` | bool | 这是"收藏夹"（"Saved Messages"）；仅在为 `true` 时出现 |
| `gigagroup` | bool | 广播群组（gigagroup）；仅在为 `true` 时出现 |
| `has_topics` | bool | 带论坛主题的超级群组；仅在为 `true` 时出现 |
| `unread_count` | int | 未读消息数（始终存在，包括 `0`） |
| `last_message` | object | 最近一条消息的 `{id, date}` |
| `stats` | object | 仅在使用 `--with-stats` 时存在 |
| `stats_error` | string | 获取统计失败时的简短 Telegram 错误码（例如 `CHANNEL_PRIVATE`）；成功时省略 |

当没有更多结果时，`cursor` 键**从 JSON 中完全省略**（不会以 `""` 形式出现）。请使用 `jq -r '.cursor? // empty'` 干净地终止循环。

### 性能说明

默认情况下，`tgs sources list` 速度很快——直接从账户列表读取对话。添加 `--with-stats` 会为每个来源额外触发约 4 次 API 调用：`channels.getFullChannel` / `messages.getFullChat` / `users.getFullUser` 获取完整信息，`messages.search` 获取总数，外加分页的 `messages.getHistory` 用于 24 小时计数和首条消息。对于有数百个对话的账户，整个运行可能需要几分钟。非活跃来源（最后一条消息超过 7 天）的统计信息会缓存到磁盘并在后续运行中复用。

### 统计准确性

- `total_messages` 是该对话的服务端消息计数。
- `messages_24h` 向后翻阅最近的消息（最多约 1000 条）并统计 24 小时窗口内的消息；超活跃来源会被该步进上限截断。
- `first_message` 是尽力而为：对于广播频道 Telegram API 可能不返回最早消息，此时该字段会被省略。

## 另请参阅

- [tgs sources inspect]({{< relref "/reference/commands/sources/inspect" >}})
