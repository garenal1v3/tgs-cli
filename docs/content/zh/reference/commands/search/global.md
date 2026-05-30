---
title: tgs search global
weight: 20
---

> **注意：** 本文档可能落后于[英文版本](/en/reference/commands/search/global/)。

# tgs search global

在所有聊天、频道和群组中全局搜索消息。

## 用法

```
tgs search global [query] [flags]
```

`query` 参数为必填项，指定要搜索的文本。

## 参数

| 参数 | 缩写 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--channels-only` | | bool | `false` | 仅在频道中搜索；与 `--folder` 不兼容 |
| `--groups-only` | | bool | `false` | 仅在群组中搜索；与 `--folder` 不兼容 |
| `--users-only` | | bool | `false` | 仅在私聊中搜索；与 `--folder` 不兼容 |
| `--archived` | | bool | `false` | 在归档中搜索（文件夹 ID 1）；设置 `--folder` 时忽略 |
| `--folder` | | string | `""` | 仅在此文件夹中搜索（id 或名称），包含已归档的聊天；与 `--channels-only`、`--groups-only`、`--users-only` 不兼容 |
| `--filter` | | string | | 消息类型过滤器（参见[过滤器值](/zh/reference/commands/search/messages/#过滤器值)） |
| `--after` | | string | | 仅返回此日期之后的消息（YYYY-MM-DD 或 unix 时间戳） |
| `--before` | | string | | 仅返回此日期之前的消息（YYYY-MM-DD 或 unix 时间戳） |
| `--limit` | `-l` | int | `50` | 最大返回消息数（1-100） |
| `--cursor` | | string | | 上次响应中的分页游标 |
| `--max-wait` | | int | `60` | FLOOD_WAIT 最大等待秒数 |
| `--profile` | `-p` | string | | 账户配置文件名称 |

未设置 `--profile` 时的配置文件解析顺序：`TGS_PROFILE` 环境变量 -> `.tgs.yaml` 文件 -> `"default"`。

## 示例

在所有聊天中搜索：

```bash
tgs search global "kubernetes"
```

仅在频道中搜索：

```bash
tgs search global "breaking news" --channels-only
```

使用日期过滤和结果限制进行搜索：

```bash
tgs search global "quarterly report" --after 2025-04-01 -l 20
```

在归档中搜索：

```bash
tgs search global "old project" --archived
```

在指定文件夹中搜索：

```bash
tgs search global "project update" --folder Work
```

## 说明

使用 `--folder` 时，`tgs` 会解析文件夹的聊天列表并执行扇出（fan-out）——
对每个聊天发送一个 `searchGlobal` 请求，然后合并结果。这是因为 Telegram
的 `searchGlobal` API 不直接接受对等体列表。

`--folder` 不能与 `--channels-only`、`--groups-only` 或 `--users-only` 同时使用。
文件夹的聊天列表始终包含其已归档的聊天，因此设置 `--folder` 时 `--archived`
不会产生任何效果，会被静默忽略。

## 输出

与 [tgs search messages]({{< relref "/reference/commands/search/messages" >}}#输出) 结构相同 —— `messages` 数组、`total` 和可选的 `cursor`。由于结果跨越多个聊天，每条消息的 `chat` 字段指明来源。

**JSON（默认）：**

```json
{
  "messages": [
    {
      "id": 188791,
      "chat": {"id": -1001754252633, "type": "channel", "title": "News", "username": "newschannel"},
      "date": "2026-05-27T10:51:40Z",
      "text": "...",
      "views": 360330,
      "forwards": 283
    },
    {
      "id": 67578,
      "chat": {"id": -1001069896405, "type": "supergroup", "title": "Tech Talk", "username": "techtalk"},
      "from": {"id": 1978176, "first_name": "Ilya", "username": "valkin"},
      "date": "2026-05-27T10:14:00Z",
      "text": "..."
    }
  ],
  "total": 2066,
  "cursor": "eyJvIjo0NzgxNSwiZCI6MCwiciI6MTc3OTg3MDc1Mn0"
}
```

注意：全局搜索的游标包含额外的 `r`（rate）字段 —— 直接原样传给 `--cursor` 即可。

## 另请参阅

- [tgs search messages]({{< relref "/reference/commands/search/messages" >}})
- [tgs search counters]({{< relref "/reference/commands/search/counters" >}})
- [tgs search calendar]({{< relref "/reference/commands/search/calendar" >}})
