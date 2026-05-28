---
title: 配置文件
weight: 20
---

# 配置文件

tgs 通过配置文件支持多个 Telegram 账户——类似于 AWS CLI 管理凭证的方式。每个配置文件拥有独立的会话，存储在单独的数据库文件中。

## 配置文件解析顺序

运行任何 tgs 命令时，活跃配置文件按以下优先级确定（从高到低）：

1. 传递给命令的 `--profile` 参数
2. `TGS_PROFILE` 环境变量
3. 当前目录或父目录中的 `.tgs.yaml` 文件
4. `"default"`（默认值）

## 按目录绑定配置文件

您可以通过创建 `.tgs.yaml` 文件将配置文件绑定到特定目录。使用 `tgs profile switch` 可自动完成此操作：

{{< snippet "profiles/switch-yaml.md" >}}

tgs 会沿目录树向上查找 `.tgs.yaml`，因此它在所有子目录中同样生效。这是为不同项目使用不同账户的推荐方式。

## 管理配置文件

### 列出配置文件

{{< snippet "profiles/list.md" >}}

`*` 标记当前活跃的配置文件。默认使用 JSON 格式输出，包含 `active` 字段。

### 切换当前目录的配置文件

{{< snippet "profiles/switch.md" >}}

在当前目录创建或覆盖 `.tgs.yaml` 文件。

### 删除配置文件

{{< snippet "profiles/delete.md" >}}

删除配置文件目录及其会话数据库。无法删除当前活跃的配置文件——请先切换到其他配置文件。

### 查看当前配置文件和账号

{{< snippet "profiles/whoami.md" >}}

`whoami` 从本地存储读取数据——无需连接 Telegram。

## 环境变量

设置 `TGS_PROFILE` 可为单个命令或整个 shell 会话覆盖配置文件：

{{< snippet "profiles/env.md" >}}

`--profile` 参数始终优先于 `TGS_PROFILE`。

## 实用示例

### 临时使用工作账号

### 设置项目目录

### 查看当前配置文件

{{< snippet "profiles/practical.md" >}}

## 完整参考

- [tgs profile]({{< relref "/reference/commands/account/profile" >}}) — list、switch、delete 子命令
- [tgs whoami]({{< relref "/reference/commands/account/whoami" >}}) — 查看当前账户信息
- [环境变量]({{< relref "/reference/environment" >}}) — 所有支持的环境变量
