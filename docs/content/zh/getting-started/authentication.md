---
title: 身份验证
weight: 10
---

# 身份验证

tgs 通过 MTProto 协议使用您的真实用户账户连接到 Telegram。每个配置文件只需认证一次——会话保存在本地，后续命令自动复用。

## 快速开始

默认方式是从已有的 Telegram Desktop 客户端导入会话：

{{< snippet "auth/quick-start.md" >}}

或选择其他登录方式：

{{< snippet "auth/methods-alt.md" >}}

## 登录方式

### Desktop 导入（默认）

从 Telegram Desktop 导入现有会话。tgs 会自动检测 macOS、Linux 和 Windows 上的 tdata 目录。

{{< snippet "auth/desktop.md" >}}

无需输入手机号或验证码——会话直接传输。

### 手机验证码

通过手机号进行交互式登录，支持两步验证（云密码）。

{{< snippet "auth/code-login.md" >}}

tgs 会提示输入 Telegram 发送的验证码。如果启用了两步验证，还会要求输入云密码。

### 非交互式登录（适用于 AI 代理）

`code` 方式支持完全非交互的两步登录流程。这是 AI 代理和自动化工具的推荐方式。

{{< snippet "auth/code-login.md" >}}

如果启用了两步验证，请在第 2 步中传入 `--password`。

### 二维码

在终端中显示二维码，使用其他设备上的 Telegram 应用扫描即可。

{{< snippet "auth/qr-login.md" >}}

扫描后 tgs 会自动完成认证。如果启用了两步验证，将提示输入云密码。

## 登录到指定配置文件

使用 `--profile` 可登录到指定的命名配置文件。如果配置文件不存在，会自动创建。

{{< snippet "auth/profile-login.md" >}}

多账户设置详见[配置文件]({{< relref "/guide/profiles" >}})。

## 验证登录

登录后，确认当前活跃账户：

{{< snippet "auth/verify.md" >}}

## 退出登录

{{< snippet "auth/logout.md" >}}

退出登录会在 Telegram 服务器上撤销会话，并清除本地会话数据。

## 完整参数参考

详见 [tgs login]({{< relref "/reference/commands/login" >}}) 和 [tgs logout]({{< relref "/reference/commands/logout" >}}) 的完整参数说明。
