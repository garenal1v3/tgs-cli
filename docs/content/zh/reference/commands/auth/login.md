---
title: tgs login
weight: 10
---

# tgs login

对 Telegram 进行身份验证。会话保存在本地，后续所有命令自动复用。

## 用法

{{< snippet "cmd-login/synopsis.md" >}}

## 参数

{{< snippet "cmd-login/flags.md" >}}

未设置 `--profile` 时的配置文件解析顺序：`TGS_PROFILE` 环境变量 → `.tgs.yaml` 文件 → `"default"`。

## 示例

{{< snippet "cmd-login/examples.md" >}}

## 认证方式

{{< snippet "cmd-login/methods.md" >}}

## 输出

登录成功时：

**JSON（默认）：**

```json
{"profile":"default","user":{"id":261054642,"phone":"+79001234567","username":"alice","first_name":"Alice","last_name":"Doe"}}
```

**文本（`--output text`）：**

```
Logged in as Alice Doe (id: 261054642, profile: default)
```

在两步 `code` 流程中，第一次调用（不带 `--code`）会发送验证码：

```json
{"profile":"default","status":"code_sent"}
```

```
Verification code sent (profile: default). Re-run with --code to complete login.
```

失败时命令以非零退出码结束并将错误写入 stderr。常见原因：验证码无效、QR 令牌过期、2FA 密码错误，或会话已存在（需先执行 `tgs logout`）。

## 另请参阅

- [tgs logout]({{< relref "/reference/commands/auth/logout" >}})
- [身份验证指南]({{< relref "/getting-started/authentication" >}})
