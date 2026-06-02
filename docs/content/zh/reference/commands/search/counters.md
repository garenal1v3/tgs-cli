---
title: tgs search counters
weight: 30
---

# tgs search counters

获取一个或多个聊天中按类型分组的消息计数（照片、视频、文档等）。

## 用法

```
tgs search counters [flags]
```

此命令不接受位置参数。

## 参数

| 参数 | 缩写 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--chat` | `-c` | string | | 要查询的聊天（用户名、电话或 ID）。未设置 `--folder` 时必填。 |
| `--folder` | | string | `""` | 查询此文件夹内的所有聊天（id 或名称）；输出包含按聊天分组的明细及汇总数据 |
| `--topic` | | int | `0` | 论坛主题 ID |
| `--filters` | | string[] | | 要统计的过滤器类型（逗号分隔；默认：全部）。参见[过滤器值](/zh/reference/commands/search/messages/#过滤器值) |
| `--max-wait` | | int | `60` | FLOOD_WAIT 最大等待秒数 |
| `--no-cache` | | bool | `false` | 禁用对等体解析缓存 |
| `--profile` | `-p` | string | | 账户配置文件名称 |

未设置 `--profile` 时的配置文件解析顺序：`TGS_PROFILE` 环境变量 -> `.tgs.yaml` 文件 -> `"default"`。

## 示例

获取频道中所有消息类型的计数：

```bash
tgs search counters -c @golang
```

获取特定类型的计数：

```bash
tgs search counters -c @mygroup --filters photo,video,document
```

获取文件夹内所有聊天的计数：

```bash
tgs search counters --folder Work --filters photo,document
```

## 输出

返回 JSON，包含 `{filter, count}` 数组。

**JSON（默认，单个聊天）：**

```json
{
  "counters": [
    {"filter": "photo", "count": 98},
    {"filter": "video", "count": 45},
    {"filter": "photo-video", "count": 143},
    {"filter": "document", "count": 0},
    {"filter": "url", "count": 188},
    {"filter": "gif", "count": 10},
    {"filter": "voice", "count": 0},
    {"filter": "music", "count": 0},
    {"filter": "round-video", "count": 0},
    {"filter": "geo", "count": 0}
  ]
}
```

**JSON（使用 `--folder`，多个聊天）：**

使用 `--folder` 时，输出包含按聊天分组的明细及汇总数据：

```json
{
  "chats": [
    {
      "chat": {"id": -1001006503122, "type": "channel", "title": "Dev News"},
      "counters": [
        {"filter": "photo", "count": 30},
        {"filter": "document", "count": 12}
      ]
    },
    {
      "chat": {"id": -1001009876543, "type": "supergroup", "title": "Team"},
      "counters": [
        {"filter": "photo", "count": 68},
        {"filter": "document", "count": 41}
      ]
    }
  ],
  "totals": [
    {"filter": "photo", "count": 98},
    {"filter": "document", "count": 53}
  ]
}
```

**文本（`--output text`）：**

```
photo        98
video        45
photo-video  143
document     0
url          188
gif          10
voice        0
music        0
round-video  0
geo          0
```

Telegram 只返回该聊天支持的过滤器类型 —— 频道通常不会返回 `contact`、`pinned`、`mention`、`phone-call`、`chat-photo`。

## 另请参阅

- [tgs search messages]({{< relref "/reference/commands/search/messages" >}})
- [tgs search global]({{< relref "/reference/commands/search/global" >}})
- [tgs search calendar]({{< relref "/reference/commands/search/calendar" >}})
