---
title: tgs sources folders
weight: 30
---

# tgs sources folders

枚举用户的 Telegram **文件夹**（UI 中的对话筛选器），并按 Telegram 客户端的方式
解析每个文件夹的内容——将置顶/包含/排除的对等体及自动规则（联系人、群组等）
应用到完整的对话列表。

默认的"所有聊天"文件夹不会出现在输出中；只有用户自定义的文件夹和共享聊天列表才会显示。

## 用法

{{< snippet "cmd-sources-folders/synopsis.md" >}}

## 参数

{{< snippet "cmd-sources-folders/flags.md" >}}

未设置 `--profile` 时的配置文件解析顺序：`TGS_PROFILE` 环境变量 -> `.tgs.yaml` 文件 -> `"default"`。

## 示例

{{< snippet "cmd-sources-folders/examples.md" >}}

## 输出

{{< snippet "cmd-sources-folders/output.md" >}}

### 文件夹字段

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | int | Telegram 对话筛选器 ID |
| `kind` | string | `custom` 表示用户自定义文件夹；`chatlist` 表示共享聊天列表 |
| `title` | string | 文件夹显示名称 |
| `emoticon` | string | 文件夹表情符号；未设置时省略 |
| `has_my_invites` | bool | 聊天列表包含您创建的邀请链接；仅在为 `true` 时出现 |
| `chats_count` | int | `chats` 数组中的聊天数量 |
| `chats` | array | 已解析的聊天条目（字段结构与 `sources list` 来源相同） |
| `total` | int | 返回的文件夹总数 |

`chats` 内的聊天条目字段结构与
[tgs sources list]({{< relref "/reference/commands/sources/list" >}}) 中的来源相同。

## 注意事项

- 文件夹内容**始终包含已归档的聊天**。Telegram 文件夹是一种可跨越归档区的
  横向视图，因此命令会同时遍历主对话列表和归档列表，解析每一个成员——
  与您在官方客户端中打开该文件夹时所见完全一致。这里没有 `--archived`
  参数：它是多余的。
- 文件夹内容不分页。文件夹总数很少（Telegram 上限约 30 个），
  每个文件夹的聊天在一次响应中全部返回。

## 另请参阅

- [tgs sources list]({{< relref "/reference/commands/sources/list" >}})
- [tgs sources inspect]({{< relref "/reference/commands/sources/inspect" >}})
