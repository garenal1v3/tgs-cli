---
title: "tgs login"
weight: 10
---

# tgs login

对 Telegram 账户进行身份验证。

## 用法

```
tgs login [参数]
```

## 描述

`tgs login` 将 tgs 连接到您的 Telegram 账户。支持三种身份验证方式：从 Telegram Desktop 导入会话（默认）、手机验证码登录，以及二维码登录。

会话数据存储在本地，后续命令无需重新登录。

## 参数

| 参数 | 简写 | 类型 | 默认值 | 描述 |
|------|------|------|--------|------|
| `--type` | `-T` | string | `desktop` | 身份验证方式：`desktop`、`code`、`qr` |
| `--desktop-dir` | `-d` | string | — | 自定义 Telegram Desktop tdata 目录路径 |
| `--passcode` | — | string | — | Telegram Desktop 本地密码 |
| `--phone` | — | string | — | 手机号码（用于验证码登录） |
| `--profile` | `-p` | string | `default` | 账号配置文件名称 |

## 示例

### 从 Telegram Desktop 导入（默认）

```bash
tgs login
```

自动检测 Telegram Desktop 数据目录并导入会话。

### 指定 tdata 目录

```bash
tgs login --type desktop --desktop-dir /home/user/.local/share/TelegramDesktop/tdata
```

### 带本地密码的 Desktop 导入

```bash
tgs login --type desktop --passcode 我的密码
```

### 手机验证码登录

```bash
tgs login --type code --phone +79001234567
```

tgs 会提示输入发送到手机的验证码。如启用了两步验证，还需输入密码。

### 二维码登录

```bash
tgs login --type qr
```

在终端显示二维码，用已登录的 Telegram 手机应用扫描。

### 登录到指定配置文件

```bash
tgs login --profile work
tgs login --type code --phone +79001234567 --profile personal
```

## 相关命令

- [tgs logout](/zh/reference/commands/logout/) — 退出登录
- [tgs whoami](/zh/reference/commands/whoami/) — 查看当前登录账号
- [tgs profile](/zh/reference/commands/profile/) — 管理配置文件
