---
title: "身份验证"
weight: 10
---

# 身份验证

在使用 tgs 之前，您需要通过身份验证将其连接到您的 Telegram 账户。tgs 支持三种登录方式，适用于不同的使用场景。

## 快速开始

最简单的登录方式是从现有的 Telegram Desktop 客户端导入会话：

```bash
tgs login
```

这将自动检测您的 Telegram Desktop 数据目录并导入会话。成功后，您可以立即开始使用 tgs。

验证登录是否成功：

```bash
tgs whoami
```

## 三种登录方式

### 方式一：从 Telegram Desktop 导入（默认）

如果您已安装 Telegram Desktop，这是最便捷的方式。tgs 会直接读取本地会话文件，无需重新输入凭证。

```bash
tgs login --type desktop
```

如果 tgs 找不到您的 Telegram Desktop 数据目录，可以手动指定路径：

```bash
tgs login --type desktop --desktop-dir /path/to/tdata
```

如果您的 Telegram Desktop 设置了本地密码（passcode），请提供：

```bash
tgs login --type desktop --passcode 您的密码
```

### 方式二：手机验证码登录

通过向您的手机号码发送验证码进行身份验证。适用于没有安装 Telegram Desktop 的环境。

```bash
tgs login --type code --phone +79001234567
```

tgs 会提示您输入收到的验证码。如果账户启用了两步验证，还会要求输入密码。

### 方式三：二维码登录

通过手机扫描二维码进行身份验证。适合在终端中快速登录。

```bash
tgs login --type qr
```

tgs 会在终端中显示二维码，使用已登录的 Telegram 手机应用扫描即可。

## 首次登录

首次运行 tgs 时，您还需要提供 Telegram API 凭证。可以通过环境变量设置：

```bash
export TGS_API_ID=12345678
export TGS_API_HASH=abcdef1234567890abcdef1234567890
tgs login
```

或者在配置文件中设置。详见[环境变量参考](/zh/reference/environment/)。

API 凭证可在 [https://my.telegram.org/apps](https://my.telegram.org/apps) 申请。

## 下一步

- [管理多个账号](/zh/guide/profiles/) — 使用配置文件在多个账号之间切换
- [tgs login 命令参考](/zh/reference/commands/login/) — 完整的参数说明
