---
title: tgs search counters
weight: 12
---

> **注意：** 本文档可能落后于[英文版本](/en/reference/commands/search-counters/)。

# tgs search counters

获取特定聊天中按类型分组的消息计数（照片、视频、文档等）。

## 用法

```
tgs search counters [flags]
```

此命令不接受位置参数。

## 参数

| 参数 | 缩写 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--chat` | `-c` | string | | 要查询的聊天（用户名、电话或 ID）。**必填。** |
| `--topic` | | int | `0` | 论坛主题 ID |
| `--filters` | | string[] | | 要统计的过滤器类型（逗号分隔；默认：全部）。参见[过滤器值](/zh/reference/commands/search-messages/#过滤器值) |
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

## 输出

返回 JSON，包含对象数组，每个对象包含过滤器类型名称和对应的消息计数。

## 另请参阅

- [tgs search messages]({{< relref "/reference/commands/search-messages" >}})
- [tgs search global]({{< relref "/reference/commands/search-global" >}})
- [tgs search calendar]({{< relref "/reference/commands/search-calendar" >}})
