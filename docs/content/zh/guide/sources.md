---
title: 来源
weight: 40
---

> **注意：** 本文档可能落后于[英文版本](/en/guide/sources/)。

# 来源

`tgs sources` 让您枚举并查看账户中的 Telegram 对话——频道、超级群组、普通群组、用户和机器人。所有输出均为 JSON 格式，可直接通过 `jq` 处理或提供给 AI 代理。

## 列出所有来源

最简单的调用方式会返回账户中的所有对话：

```bash
tgs sources list
```

响应为包含 `sources` 数组和 `total` 总数的 JSON 对象：

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
    }
  ],
  "total": 287,
  "cursor": ""
}
```

## 按类型筛选

使用 `--type` 将结果限定为特定类型的对话。可传入单个类型或逗号分隔的多个类型：

```bash
# 仅频道
tgs sources list --type channel

# 频道和超级群组
tgs sources list --type channel,supergroup

# 仅机器人
tgs sources list --type bot
```

支持的类型：`channel`、`supergroup`、`group`、`user`、`bot`。

## 获取完整指标

默认情况下，`tgs sources list` 速度很快，不会发起额外的 API 调用。添加 `--with-stats` 可为每个来源获取消息统计信息：

```bash
tgs sources list --with-stats
```

响应中的每个来源将包含 `stats` 对象：

```json
{
  "stats": {
    "total_messages": 12345,
    "messages_24h": 3,
    "first_message": {"id": 1, "date": "2015-08-26T10:00:00Z"}
  }
}
```

**性能说明：** `--with-stats` 会为每个来源触发约 4 次 API 调用。对于大型账户，这可能需要几分钟。非活跃来源（最后一条消息超过 7 天）的统计信息会缓存到磁盘，在来源收到新消息之前，后续运行将复用缓存。

## 查看单个来源

`tgs sources inspect` 提供单个来源的完整信息。适用于您已订阅的来源和您未加入的公开频道：

```bash
# 通过用户名
tgs sources inspect @durov

# 通过数字 Telegram ID
tgs sources inspect -1001234567890

# 您的收藏夹（Saved Messages）
tgs sources inspect -
```

响应中包含 `list` 中没有的额外字段：`subscribed`、`description`、`creation_date`、`invite_link`。

### 查看未订阅的频道

只要知道 `@username`，您就可以在不加入的情况下查看任何公开频道：

```bash
tgs sources inspect @somebigtechchannel
```

响应中 `subscribed` 将为 `false`，`stats` 将为 `null`。

跳过统计信息以加快速度：

```bash
tgs sources inspect @durov --no-stats
```

## 大型账户的分页处理

如果您有数百个对话，可使用 `--limit` 和 `--cursor` 进行分页访问：

```bash
# 第一页
tgs sources list --limit 50
```

响应中包含 `cursor` 字符串，将其传回以获取下一页：

```bash
tgs sources list --limit 50 --cursor "eyJvIjo1MCwiZCI6MH0"
```

当响应中的 `cursor` 字段为空字符串时，表示已到达最后一页。

Shell 脚本中的典型分页循环：

```bash
cursor=""
while true; do
  if [ -z "$cursor" ]; then
    result=$(tgs sources list --limit 50)
  else
    result=$(tgs sources list --limit 50 --cursor "$cursor")
  fi

  echo "$result" | jq '.sources[]'

  cursor=$(echo "$result" | jq -r '.cursor? // empty')
  [ -z "$cursor" ] && break
done
```

## 已归档对话

已归档的对话默认不出现在列表中。传入 `--archived` 可将其包含在内：

```bash
tgs sources list --archived
```

## 完整参考

- [tgs sources list]({{< relref "/reference/commands/sources/list" >}}) — 所有参数和输出字段
- [tgs sources inspect]({{< relref "/reference/commands/sources/inspect" >}}) — 查看单个来源
