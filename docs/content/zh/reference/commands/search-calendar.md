---
title: tgs search calendar
weight: 13
---

> **注意：** 本文档可能落后于[英文版本](/en/reference/commands/search-calendar/)。

# tgs search calendar

获取特定聊天和过滤器类型的按日期分组的消息搜索结果。

## 用法

```
tgs search calendar [flags]
```

此命令不接受位置参数。

## 参数

| 参数 | 缩写 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--chat` | `-c` | string | | 要查询的聊天（用户名、电话或 ID）。**必填。** |
| `--filter` | | string | | 消息类型过滤器（参见[过滤器值](/zh/reference/commands/search-messages/#过滤器值)）。**必填。** |
| `--max-wait` | | int | `60` | FLOOD_WAIT 最大等待秒数 |
| `--no-cache` | | bool | `false` | 禁用对等体解析缓存 |
| `--profile` | `-p` | string | | 账户配置文件名称 |

未设置 `--profile` 时的配置文件解析顺序：`TGS_PROFILE` 环境变量 -> `.tgs.yaml` 文件 -> `"default"`。

## 示例

获取频道中照片消息的日历：

```bash
tgs search calendar -c @travel_photos --filter photo
```

获取群组中共享文档的日历：

```bash
tgs search calendar -c @dev_team --filter document
```

## 输出

返回 JSON，包含日期条目数组。每个条目包含日期和该日期匹配消息的数量，以及一个代表性消息 ID。

## 另请参阅

- [tgs search messages]({{< relref "/reference/commands/search-messages" >}})
- [tgs search global]({{< relref "/reference/commands/search-global" >}})
- [tgs search counters]({{< relref "/reference/commands/search-counters" >}})
