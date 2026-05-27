---
title: 搜索
weight: 30
---

# 搜索

> **注意：** 本文档可能落后于[英文版本](/en/guide/search/)。

tgs 提供四个搜索子命令，覆盖 Telegram 消息搜索的所有场景：

| 子命令 | 用途 |
|--------|------|
| `tgs search messages` | 在指定聊天中搜索消息 |
| `tgs search global` | 跨所有聊天全局搜索 |
| `tgs search counters` | 按内容类型统计消息数量 |
| `tgs search calendar` | 按日期分组查看搜索结果 |

## 在聊天中搜索

使用 `tgs search messages` 在指定聊天中搜索消息。必须通过 `--chat`/`-c` 指定至少一个聊天：

```bash
tgs search messages "查询关键词" -c @channel_name
```

搜索结果以 JSON 格式输出，包含消息列表、总数和分页游标：

```json
{
  "messages": [
    {
      "id": 12345,
      "chat": {"id": 1234567890, "type": "channel", "title": "频道名称"},
      "from": {"id": 987654321, "first_name": "张", "last_name": "三"},
      "date": "2025-01-15T10:30:00Z",
      "text": "消息内容..."
    }
  ],
  "total": 42,
  "cursor": "eyJvIjoxMjM0NX0"
}
```

## 多个聊天

可以通过逗号分隔或重复 `-c` 参数同时搜索多个聊天：

```bash
# 逗号分隔
tgs search messages "查询" -c @channel1,@channel2,@group1

# 多次指定 -c
tgs search messages "查询" -c @channel1 -c @channel2
```

多聊天搜索会依次查询每个聊天，按日期降序合并结果，并使用多游标 (multi-cursor) 进行分页。

## 按作者筛选

使用 `--from`/`-f` 只返回特定用户发送的消息：

```bash
tgs search messages "关键词" -c @group --from @username
tgs search messages "关键词" -c @group -f 123456789
```

`--from` 支持与 `--chat` 相同的标识符格式（用户名、数字 ID、手机号等）。

## 按内容类型筛选

使用 `--filter` 按消息中的媒体类型筛选。支持以下 15 种过滤器：

| 过滤器 | 说明 |
|--------|------|
| `photo` | 图片 |
| `video` | 视频 |
| `photo-video` | 图片和视频 |
| `document` | 文件/文档 |
| `url` | 包含链接的消息 |
| `gif` | GIF 动图 |
| `voice` | 语音消息 |
| `music` | 音乐文件 |
| `round-video` | 圆形视频消息 |
| `geo` | 地理位置 |
| `contact` | 联系人 |
| `pinned` | 置顶消息 |
| `mention` | 提及当前用户的消息 |
| `phone-call` | 通话记录 |
| `chat-photo` | 聊天头像变更 |

示例：

```bash
# 搜索频道中的所有图片
tgs search messages "" -c @channel --filter photo

# 搜索包含链接的消息
tgs search messages "golang" -c @group --filter url
```

## 按日期筛选

使用 `--after` 和 `--before` 限定搜索的时间范围。支持 `YYYY-MM-DD` 格式和 Unix 时间戳：

```bash
# 搜索 2025 年 1 月之后的消息
tgs search messages "查询" -c @channel --after 2025-01-01

# 搜索特定日期范围内的消息
tgs search messages "查询" -c @channel --after 2025-01-01 --before 2025-06-01

# 使用 Unix 时间戳
tgs search messages "查询" -c @channel --after 1704067200
```

## 论坛主题

对于支持论坛功能的群组/超级群组，使用 `--topic` 限定在特定主题内搜索：

```bash
tgs search messages "bug" -c @dev_group --topic 42
```

`--topic` 接受论坛主题的消息 ID（即该主题的顶部消息 ID）。

## 全局搜索

`tgs search global` 跨所有聊天搜索消息，无需指定具体聊天：

```bash
tgs search global "搜索关键词"
```

### 限定聊天类型

使用互斥的类型标志缩小搜索范围：

```bash
# 仅搜索频道
tgs search global "新闻" --channels-only

# 仅搜索群组
tgs search global "讨论" --groups-only

# 仅搜索私聊
tgs search global "你好" --users-only
```

### 按文件夹筛选

使用 `--folder` 限定在特定文件夹中搜索：

```bash
tgs search global "查询" --folder 3
```

全局搜索同样支持 `--filter`、`--after`、`--before` 参数。

## 内容计数

`tgs search counters` 返回指定聊天中各类型消息的数量统计：

```bash
tgs search counters -c @channel
```

输出示例：

```json
{
  "counters": [
    {"filter": "photo", "count": 1523},
    {"filter": "video", "count": 342},
    {"filter": "document", "count": 89},
    {"filter": "url", "count": 4210}
  ]
}
```

默认统计所有 14 种过滤器类型。可以通过 `--filters` 指定只统计部分类型：

```bash
tgs search counters -c @channel --filters photo,video,document
```

`counters` 同样支持 `--topic` 参数，可以统计特定论坛主题内的消息。

## 日历视图

`tgs search calendar` 按日期分组返回搜索结果的分布情况。需要同时指定 `--chat` 和 `--filter`：

```bash
tgs search calendar -c @channel --filter photo
```

输出示例：

```json
{
  "periods": [
    {"date": "2025-05-20", "count": 15, "min_msg_id": 8901, "max_msg_id": 8920},
    {"date": "2025-05-19", "count": 8, "min_msg_id": 8880, "max_msg_id": 8900}
  ],
  "total": 1523
}
```

每个时段包含日期、消息数量以及该日期范围内的最小和最大消息 ID，便于后续精确定位。

## 游标分页

当搜索结果超过单页限制（默认 50 条，最大 100 条）时，响应中会包含 `cursor` 字段。将其传递给下一次请求以获取后续页面：

```bash
# 第一页
tgs search messages "查询" -c @channel --limit 20

# 使用返回的游标获取下一页
tgs search messages "查询" -c @channel --limit 20 --cursor "eyJvIjoxMjM0NX0"
```

当没有更多结果时，响应中不再包含 `cursor` 字段。

游标是 base64 编码的 JSON 对象，包含偏移信息。单聊天搜索和多聊天搜索使用不同的游标格式，但对调用方来说是透明的——只需原样传回即可。

全局搜索 (`tgs search global`) 同样支持 `--cursor` 分页。

## 聊天解析

`--chat`/`-c`、`--from`/`-f` 等参数接受多种格式来标识 Telegram 聊天或用户：

| 格式 | 示例 | 说明 |
|------|------|------|
| `@用户名` | `@durov` | 带 `@` 前缀的用户名 |
| 用户名 | `durov` | 不带 `@` 的用户名 |
| 正数 ID | `123456789` | 用户 ID |
| 负数 ID | `-1001234567890` | 超级群组/频道 ID |
| 负数 ID | `-123456` | 普通群组 ID |
| 手机号 | `+79001234567` | 带 `+` 前缀，至少 7 位数字 |
| t.me 链接 | `t.me/durov` | Telegram 短链接 |
| 完整链接 | `https://t.me/durov` | 带协议的 Telegram 链接 |
| telegram.me | `telegram.me/durov` | 旧域名格式 |

数字 ID 的解析规则：
- 正数 → 用户
- 小于 `-1000000000000` → 超级群组/频道（实际 ID = 绝对值 - 1000000000000）
- 其他负数 → 普通群组（实际 ID = 绝对值）

> **提示：** 邀请链接（`t.me/+hash`）不支持用于搜索。

## 速率限制处理

Telegram API 对请求频率有限制。当触发 `FLOOD_WAIT` 时，tgs 会自动等待后重试：

```
[tgs] FLOOD_WAIT: waiting 5s before retry (attempt 1/4)
```

### --max-wait

`--max-wait` 参数控制 tgs 愿意等待的最长秒数（默认 60 秒）：

```bash
# 最多等待 120 秒
tgs search messages "查询" -c @channel --max-wait 120

# 不等待，立即返回错误
tgs search messages "查询" -c @channel --max-wait 0
```

如果 Telegram 要求的等待时间超过 `--max-wait`，tgs 将立即返回 `FLOOD_WAIT_N` 错误而不再等待。

除 FLOOD_WAIT 外，tgs 对临时性错误使用指数退避重试（最多 3 次），对永久性错误（如 `PEER_ID_INVALID`、`USERNAME_NOT_OCCUPIED` 等）立即返回。

## 完整参考

- [tgs search messages]({{< relref "/reference/commands/search-messages" >}}) — 聊天内搜索命令参考
- [tgs search global]({{< relref "/reference/commands/search-global" >}}) — 全局搜索命令参考
- [tgs search counters]({{< relref "/reference/commands/search-counters" >}}) — 计数命令参考
- [tgs search calendar]({{< relref "/reference/commands/search-calendar" >}}) — 日历视图命令参考
