---
title: tgs logout
weight: 20
---

# tgs logout

在 Telegram 服务器上撤销当前会话，并清除本地会话数据。

## 用法

{{< snippet "cmd-logout/synopsis.md" >}}

## 参数

{{< snippet "cmd-logout/flags.md" >}}

未设置 `--profile` 时的配置文件解析顺序：`TGS_PROFILE` 环境变量 → `.tgs.yaml` 文件 → `"default"`。

## 示例

{{< snippet "cmd-logout/examples.md" >}}

## 输出

**JSON（默认）：**

```json
{"profile":"default","status":"logged_out"}
```

**文本（`--output text`）：**

```
Logged out (profile: default)
```

若配置文件未登录，命令以非零退出码结束并将错误写入 stderr。

## 说明

退出登录执行两个步骤：

1. 调用 Telegram API 撤销会话令牌（会话在 Telegram 服务器端失效）
2. 清除该配置文件的本地会话数据库

退出后，必须重新运行 `tgs login` 才能继续使用该配置文件。

## 另请参阅

- [tgs login]({{< relref "/reference/commands/auth/login" >}})
- [身份验证指南]({{< relref "/getting-started/authentication" >}})
