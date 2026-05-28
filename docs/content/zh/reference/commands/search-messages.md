---
title: tgs search messages
weight: 10
---

> **注意：** 本文档可能落后于[英文版本](/en/reference/commands/search-messages/)。

# tgs search messages

在一个或多个聊天、频道或群组中搜索消息。

## 用法

```
tgs search messages [query] [flags]
```

`query` 参数为必填项，指定要搜索的文本。

## 参数

| 参数 | 缩写 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--chat` | `-c` | string[] | | 要搜索的聊天（用户名、电话、ID；可重复使用，支持逗号分隔）。**必填。** |
| `--from` | `-f` | string | | 按发送者筛选（用户名、电话或 ID） |
| `--filter` | | string | | 消息类型过滤器（参见[过滤器值](#过滤器值)） |
| `--after` | | string | | 仅返回此日期之后的消息（YYYY-MM-DD 或 unix 时间戳） |
| `--before` | | string | | 仅返回此日期之前的消息（YYYY-MM-DD 或 unix 时间戳） |
| `--topic` | | int | `0` | 论坛主题 ID |
| `--limit` | `-l` | int | `50` | 最大返回消息数（1-100） |
| `--cursor` | | string | | 上次响应中的分页游标 |
| `--max-wait` | | int | `60` | FLOOD_WAIT 最大等待秒数 |
| `--no-cache` | | bool | `false` | 禁用对等体解析缓存 |
| `--include-comments` | | bool | `false` | 同时搜索每个频道关联的讨论组 |
| `--profile` | `-p` | string | | 账户配置文件名称 |

未设置 `--profile` 时的配置文件解析顺序：`TGS_PROFILE` 环境变量 -> `.tgs.yaml` 文件 -> `"default"`。

### 过滤器值

`photo`、`video`、`photo-video`、`document`、`url`、`gif`、`voice`、`music`、`round-video`、`geo`、`contact`、`pinned`、`mention`、`phone-call`、`chat-photo`。

## 示例

在频道中搜索关键词：

```bash
tgs search messages "release notes" -c @golang
```

跨多个聊天搜索：

```bash
tgs search messages "deployment" -c @devops_team -c @infrastructure
```

按发送者和消息类型筛选：

```bash
tgs search messages "config" -c @mygroup --from @alice --filter document
```

在日期范围内搜索：

```bash
tgs search messages "outage" -c @incidents --after 2025-01-01 --before 2025-06-01
```

分页浏览结果：

```bash
tgs search messages "bug" -c @dev -l 10 --cursor "eyJvIjo1MCwiZCI6MH0"
```

## 输出

返回 JSON，包含匹配消息的数组和用于分页的 `cursor` 字段。每条消息包含发送者、日期、文本、聊天信息和消息 ID。

## 另请参阅

- [tgs search global]({{< relref "/reference/commands/search-global" >}})
- [tgs search counters]({{< relref "/reference/commands/search-counters" >}})
- [tgs search calendar]({{< relref "/reference/commands/search-calendar" >}})
