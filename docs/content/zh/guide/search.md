---
title: 搜索
weight: 30
---

# 搜索

tgs 提供四个搜索子命令，它们直接对应 Telegram 的 MTProto 搜索方法。所有输出均为 JSON，便于通过管道传入 `jq`、交给 AI 智能体或以编程方式处理。

## 在聊天中搜索

最常见的操作：在指定聊天中搜索消息。

```bash
tgs search messages "docker compose" -c @devops_notes
```

查询参数为必填项。`-c` / `--chat` 标志指定在哪个聊天中搜索。所有支持的格式请参见下方的[聊天解析](#聊天解析)。

## 多个聊天

通过重复该标志或使用逗号分隔的值传入多个聊天：

```bash
# 重复标志
tgs search messages "release" -c @frontend -c @backend

# 逗号分隔
tgs search messages "release" -c @frontend,@backend
```

tgs 会依次对每个聊天执行搜索，并将结果合并为一个 JSON 响应。

## 按作者筛选

使用 `--from` / `-f` 只显示来自特定发送者的消息：

```bash
tgs search messages "bug" -c @dev_chat --from @alice
```

`--from` 的值接受与 `--chat` 相同的格式（用户名、数字 ID、手机号或 `t.me` 链接）。

## 按内容类型筛选

`--filter` 标志将结果限定为特定的消息类型。共有 15 种可用过滤器：

| 过滤器 | 说明 |
|---|---|
| `photo` | 图片 |
| `video` | 视频 |
| `photo-video` | 图片和视频 |
| `document` | 文件和文档 |
| `url` | 包含链接的消息 |
| `gif` | GIF 动图 |
| `voice` | 语音消息 |
| `music` | 音频文件 / 音乐 |
| `round-video` | 圆形视频消息 |
| `geo` | 位置和实时位置 |
| `contact` | 共享的联系人 |
| `pinned` | 置顶消息 |
| `mention` | 提及当前用户的消息 |
| `phone-call` | 通话记录 |
| `chat-photo` | 聊天头像变更 |

示例 —— 查找频道中的所有文档：

```bash
tgs search messages "" -c @project_files --filter document
```

空查询（`""`）配合过滤器会返回该类型的所有消息。

## 按日期筛选

使用 `--after` 和 `--before` 将结果限定在某个日期范围内。两者都接受两种格式：

- **日期字符串**：`YYYY-MM-DD`（解释为 UTC 午夜）
- **Unix 时间戳**：自纪元以来的原始秒数

```bash
# 2025 年 1 月的消息
tgs search messages "deploy" -c @ops --after 2025-01-01 --before 2025-02-01

# 使用 Unix 时间戳
tgs search messages "deploy" -c @ops --after 1704067200 --before 1706745600
```

两个标志均为可选，且可独立使用。

## 论坛主题

对于启用了论坛模式的超级群组，使用 `--topic` 按主题 ID 在特定主题内搜索：

```bash
tgs search messages "error" -c @dev_forum --topic 42
```

不带 `--topic` 时，搜索会覆盖论坛中的所有主题。

## 搜索频道评论

Telegram 频道可以关联一个讨论组，每个帖子下方的评论都存放在该讨论组中。使用 `--include-comments` 可在一条命令中自动同时搜索频道及其讨论组 —— 无需查找讨论组的用户名：

```bash
tgs search messages "release" -c @somechannel --include-comments
```

`--chat` 中的每个频道都会被检查是否存在关联讨论组；若存在，该讨论组会被透明地加入搜索。对于没有评论的频道，此标志不会产生影响。

结果会合并帖子和评论并按日期排序。每条评论的 `reply_to_msg_id` 字段指向其所回复的消息，便于重建讨论线索。

## 全局搜索

使用 `tgs search global` 一次性搜索您的所有聊天：

```bash
tgs search global "meeting notes"
```

### 按聊天类型缩小范围

将全局搜索限定为特定的聊天类型：

```bash
# 仅频道
tgs search global "announcement" --channels-only

# 仅群组
tgs search global "discussion" --groups-only

# 仅私聊
tgs search global "hey" --users-only
```

这些标志在实际使用中是互斥的 —— Telegram 应用其中第一个被设置的标志。

### 搜索归档区

传入 `--archived`，将全局搜索限定为 Telegram 内置归档文件夹中的聊天：

```bash
tgs search global "old thread" --archived
```

全局搜索同样支持 `--filter`、`--after`、`--before`、`--limit` 和 `--cursor`，与 `search messages` 相同。

## 内容计数

获取某个聊天中按类型统计的消息数量明细，而无需获取消息本身：

```bash
tgs search counters -c @mychannel
```

这会返回所有 15 种过滤器类型的计数。若只查询特定类型：

```bash
tgs search counters -c @mychannel --filters photo,video,document
```

计数同样支持论坛群组的 `--topic`。

## 日历视图

获取按日期分组的搜索结果 —— 便于了解特定内容是何时发布的：

```bash
tgs search calendar -c @mychannel --filter photo
```

日历查询同时需要 `--chat` 和 `--filter`。响应包含一个日期列表，每个日期对应消息数量以及消息 ID 范围（`min_msg_id`、`max_msg_id`）。

## 按文件夹筛选

所有搜索子命令均接受 `--folder <id|名称>`，它将搜索范围限定为用户自定义 Telegram 文件夹中的聊天。`--folder` 接受数字文件夹 ID 或不区分大小写的文件夹名称：

```bash
# 搜索“Work”文件夹中所有聊天的消息
tgs search messages "release notes" --folder Work

# 统计“Crypto”文件夹中各聊天的媒体数量
tgs search counters --folder Crypto

# 限定在文件夹范围内的日历视图
tgs search calendar --folder Crypto --filter photo
```

在底层，`tgs` 会将文件夹解析为其成员聊天，然后对所有聊天扇出 (fan-out) 请求，合并结果的方式与多聊天搜索相同。解析出的聊天列表始终包含文件夹的**已归档**聊天 —— 文件夹是一种可跨越归档区的视图 —— 因此设置 `--folder` 时无需 `--archived`（且会被忽略）。

> **关于 `tgs search global` 的说明：** 原先 `search global` 上的 `--folder` 标志（通过整数 ID 选择 Telegram 主文件夹或归档）已被 `--archived` 取代。`--folder` 标志现在在所有子命令中统一表示用户自定义文件夹。

要浏览您的文件夹并查找其 ID 或准确名称，请使用 `tgs sources folders`。

## 游标分页

tgs 使用基于游标的分页，而非基于偏移量的分页。这与 Telegram 的原生方式一致，可避免在分页过程中有新消息到达时出现跳过或重复的结果。

每个搜索响应都包含一个 `cursor` 字段（若没有更多结果则为空字符串）。将其通过 `--cursor` 传回以获取下一页：

```bash
# 第一页（默认：50 条结果）
tgs search messages "update" -c @news --limit 10

# 使用上一个响应中的游标获取下一页
tgs search messages "update" -c @news --limit 10 --cursor "eyJvIjo1MCwiZCI6MH0"
```

`--limit` 标志控制页面大小（1-100，默认 50）。

脚本中典型的分页循环：

```bash
cursor=""
while true; do
  if [ -z "$cursor" ]; then
    result=$(tgs search messages "query" -c @chat --limit 20)
  else
    result=$(tgs search messages "query" -c @chat --limit 20 --cursor "$cursor")
  fi

  echo "$result" | jq '.messages[]'

  cursor=$(echo "$result" | jq -r '.cursor? // empty')
  [ -z "$cursor" ] && break
done
```

## 聊天解析

tgs 接受多种格式来标识聊天、用户和群组：

| 格式 | 示例 | 说明 |
|---|---|---|
| `@username` | `@durov` | 带 @ 前缀的用户名 |
| `username` | `durov` | 不带 @ 前缀的用户名 |
| 数字 ID | `123456789` | 正数的用户/聊天 ID |
| 负数 ID | `-1001234567890` | 超级群组/频道 ID（Telegram 的原生格式） |
| `t.me` 链接 | `t.me/durov` | Telegram 链接（带或不带 `https://`） |
| 手机号 | `+79001234567` | 国际手机号格式（7+ 位数字） |

所有这些都可用于 `--chat`、`--from` 以及任何接受 peer 引用的标志。

> **提示：** 邀请链接（`t.me/+hash`）不支持用于搜索。

tgs 会在本地缓存已解析的 peer，以避免冗余的 API 调用。如果怀疑数据陈旧，使用 `--no-cache` 可绕过缓存。

## 速率限制处理

Telegram 通过 `FLOOD_WAIT` 错误强制实施速率限制，该错误会指定重试前必须等待的秒数。tgs 会自动处理：

1. 收到 `FLOOD_WAIT` 时，tgs 向 stderr 打印警告并休眠所需的时长。
2. 等待之后，重试该请求。
3. 如果所需等待时间超过 `--max-wait`（默认：60 秒），tgs 会立即返回错误而非阻塞。

```bash
# 允许最多 120 秒的 flood wait
tgs search messages "query" -c @bigchannel --max-wait 120

# 快速失败 —— 等待不超过 5 秒
tgs search messages "query" -c @bigchannel --max-wait 5
```

对于临时性网络错误，tgs 会以指数退避加抖动重试最多 3 次。永久性错误（无效 peer、权限不足等）永远不会重试。

## 完整参考

- [tgs search messages]({{< relref "/reference/commands/search/messages" >}}) — 在指定聊天中搜索
- [tgs search global]({{< relref "/reference/commands/search/global" >}}) — 跨所有聊天搜索
- [tgs search counters]({{< relref "/reference/commands/search/counters" >}}) — 按类型统计消息数量明细
- [tgs search calendar]({{< relref "/reference/commands/search/calendar" >}}) — 按日期分组的搜索结果
