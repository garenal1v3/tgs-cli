---
title: tgs profile
weight: 10
---

# tgs profile

管理账户配置文件。

## 用法

{{< snippet "cmd-profile/synopsis.md" >}}

## 子命令

### tgs profile list

列出所有配置文件及其关联的账户信息。

{{< snippet "cmd-profile/list.md" >}}

**JSON（默认）：**

```json
[
  {"name":"default","phone":"+79001234567","username":"alice","id":261054642,"active":true},
  {"name":"work","phone":"+79001230000","username":"alice_work","id":342178890,"active":false}
]
```

**文本（`--output text`）：**

```
* default (@alice)
  work (@alice_work)
```

文本输出中用 `*` 标记当前活跃的配置文件，JSON 中标记为 `"active": true`。若未缓存 username 和 phone，则只显示配置文件名。

---

### tgs profile switch

通过写入 `.tgs.yaml` 文件来设置当前目录的活跃配置文件。

{{< snippet "cmd-profile/switch.md" >}}

**JSON（默认）：**

```json
{"profile":"work","status":"switched","dir":"/Users/alice/projects"}
```

**文本（`--output text`）：**

```
Switched to profile "work" (wrote .tgs.yaml in /Users/alice/projects)
```

在当前目录及所有子目录中运行的 tgs 命令都将使用指定的配置文件。

---

### tgs profile delete

删除配置文件及其本地会话数据库。

{{< snippet "cmd-profile/delete.md" >}}

**JSON（默认）：**

```json
{"profile":"old","status":"deleted"}
```

**文本（`--output text`）：**

```
Deleted profile "old"
```

无法删除当前活跃的配置文件。请先切换到其他配置文件。

## 另请参阅

- [tgs whoami]({{< relref "/reference/commands/account/whoami" >}})
- [配置文件指南]({{< relref "/guide/profiles" >}})
