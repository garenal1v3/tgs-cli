---
title: "tgs whoami"
weight: 40
---

# tgs whoami

显示当前登录的 Telegram 账号信息。

## 用法

```
tgs whoami [参数]
```

## 描述

`tgs whoami` 返回当前配置文件对应的 Telegram 账号信息，包括用户 ID、用户名和手机号码。可用于验证登录状态或确认当前使用的账号。

## 参数

| 参数 | 简写 | 类型 | 默认值 | 描述 |
|------|------|------|--------|------|
| `--profile` | `-p` | string | `default` | 要查询的账号配置文件名称 |

## 示例

### 查看默认配置文件的账号

```bash
tgs whoami
```

示例输出：

```json
{
  "id": 123456789,
  "username": "myusername",
  "first_name": "张",
  "last_name": "三",
  "phone": "+79001234567"
}
```

### 查看指定配置文件的账号

```bash
tgs whoami --profile work
```

## 相关命令

- [tgs login](/zh/reference/commands/login/) — 登录账户
- [tgs logout](/zh/reference/commands/logout/) — 退出登录
- [tgs profile](/zh/reference/commands/profile/) — 管理配置文件
