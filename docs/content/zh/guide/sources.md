---
title: 来源
weight: 40
---

# 来源

`tgs sources` 让您枚举并查看账户中的 Telegram 对话——频道、超级群组、普通群组、用户和机器人。所有输出均为 JSON 格式，可直接通过 `jq` 处理或提供给 AI 代理。

> **通过 `jq` 解析时：** `tgs` 把诊断日志（`[tgs] retry: …`、`[tgs] FLOOD_WAIT: …`）写入 **stderr**，JSON 写入 **stdout**。在管道传给 `jq` 时请丢弃 stderr，以免解析失败：
>
> ```bash
> tgs sources list --with-stats 2>/dev/null | jq '.sources[].title'
> ```
>
> 避免使用 `2>&1 | jq …`——那会把日志混进 stdout 并破坏 JSON 解析。

## 列出所有来源

最简单的调用方式会返回账户中的所有对话：

```bash
tgs sources list
```

响应为包含 `sources` 数组、`total` 总数和可选 `cursor` 的 JSON 对象：

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
  "total": 287
}
```

> **`total`** 是您账户中的对话总数（应用 `--type` 筛选**之前**）。CLI 保证 `total >= len(sources)`（Telegram 自身的计数器有时已过期，我们会把 `total` 抬升到实际返回的数量）。请勿用 `total` 除以数组长度估算页数——请使用 `cursor` 分页。

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

`first_message` 是尽力而为的：对于广播频道，Telegram MTProto 有时不会返回最早的消息，此时该字段会被省略。`total_messages` 和 `messages_24h` 始终存在。

**性能说明：** `--with-stats` 为每个来源大约触发 4 次 API 调用，且 `messages_24h` 会向后翻阅消息历史（最多约 1000 条）直到跨越 24 小时阈值——超活跃来源会被这个上限截断。对于大型账户，整个运行可能需要几分钟。非活跃来源（最后一条消息超过 7 天）的统计信息会缓存到磁盘，并在多次运行之间复用。

## 查看单个来源

`tgs sources inspect` 提供单个来源的完整信息。适用于您已订阅的来源和您未加入的公开频道：

```bash
# 通过用户名
tgs sources inspect @durov

# 通过数字 Telegram ID —— 使用 `id:` 前缀（裸 `-1001234…` 会被 CLI 的
# 参数解析器当作旗标；用 `id:` 或插入 `--` 分隔符避免这一点）。
tgs sources inspect id:-1001234567890

# `--` 分隔符同样有效，但其后的内容都会被视为位置参数 ——
# 命令的所有旗标必须放在 `--` 之前：
tgs sources inspect --no-stats -- -1001234567890

# 您的收藏夹（Saved Messages）
tgs sources inspect -
```

响应中包含 `list` 中没有的额外字段：`subscribed`、`description`、`creation_date`、`invite_link`。对于广播频道，`creation_date` 反映频道的创建时间；对于用户，该字段会被省略（Telegram 不公开用户注册日期）。

> **数字 ID 的限制：** 数字 ID 仅在该 peer 此前通过 `@username` 或 `+电话号码` 被解析过（这样会把 access_hash 写入本地 peer 缓存）后才能使用。普通的 `tgs sources list` **不会**写入 peer 缓存。如果数字 ID 的 inspect 报错 "peer is not in the local cache"，请先执行一次 `inspect @<username>` 或 `+<phone>`，再用 `id:<n>` 重试。

### 查看未订阅的频道

只要知道 `@username`，您就可以在不加入的情况下查看任何公开频道：

```bash
tgs sources inspect @somebigtechchannel
```

响应中 `subscribed` 将为 `false`，且 `stats` 键将从 JSON 中完全省略。其他显示字段（`title`、`username`、`access`、`members_count`、`description`、`verified` 等）仍会从公开频道信息中填充。

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

当没有更多页面时，JSON 中将**完全不包含 `cursor` 键**（而不是返回值为 `""` 的该键）。上述分页循环通过 `jq -r '.cursor? // empty'` 正确处理了这种情况。

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

## 文件夹

Telegram 允许用户将对话分组到**文件夹**中，即官方客户端顶部的选项卡。`tgs sources folders` 可以枚举您的文件夹，并展开每个文件夹中包含的聊天，与 UI 显示保持一致：

```bash
tgs sources folders
```

默认的"全部聊天"视图不包含在输出中——只有用户自定义的文件夹和共享聊天列表才会出现。

文件夹内容始终包含其**已归档**的聊天：文件夹是一种可跨越归档区的视图，因此 `tgs sources folders`（以及任何带 `--folder` 的命令）都会解析每一个成员，无论其是否已归档。没有单独的开关来启用此行为。

确定目标文件夹后，可将 `--folder` 传递给其他命令，将结果范围限定为该文件夹的内容。`--folder` 接受数字文件夹 ID 或不区分大小写的文件夹名称：

```bash
# 仅列出"Crypto"文件夹中的聊天
tgs sources list --folder Crypto

# 同上，使用数字 ID
tgs sources list --folder 3
```

如需了解如何在搜索命令中使用 `--folder`，请参阅搜索指南中的[按文件夹筛选]({{< relref "/guide/search#按文件夹筛选" >}})部分。

## 完整参考

- [tgs sources list]({{< relref "/reference/commands/sources/list" >}}) — 所有参数和输出字段
- [tgs sources inspect]({{< relref "/reference/commands/sources/inspect" >}}) — 查看单个来源
- [tgs sources folders]({{< relref "/reference/commands/sources/folders" >}}) — 列出文件夹及其内容
