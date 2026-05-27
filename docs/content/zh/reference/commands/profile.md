---
title: tgs profile
weight: 20
---

# tgs profile

管理账户配置文件。

## 用法

{{< snippet "cmd-profile/synopsis.md" >}}

## 子命令

### tgs profile list

列出所有配置文件及其关联的账户信息。

{{< snippet "cmd-profile/list.md" >}}

文本输出中用 `*` 标记当前活跃的配置文件。

JSON 输出为默认格式，包含每个配置文件的 `active` 布尔字段。

---

### tgs profile switch

通过写入 `.tgs.yaml` 文件来设置当前目录的活跃配置文件。

{{< snippet "cmd-profile/switch.md" >}}

在当前目录及所有子目录中运行的 tgs 命令都将使用指定的配置文件。

---

### tgs profile delete

删除配置文件及其本地会话数据库。

{{< snippet "cmd-profile/delete.md" >}}

无法删除当前活跃的配置文件。请先切换到其他配置文件。

## 另请参阅

- [tgs whoami]({{< relref "/reference/commands/whoami" >}})
- [配置文件指南]({{< relref "/guide/profiles" >}})
