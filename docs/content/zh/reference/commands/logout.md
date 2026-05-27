---
title: "tgs logout"
weight: 20
---

# tgs logout

退出当前 Telegram 账户登录。

## 用法

```
tgs logout [参数]
```

## 描述

`tgs logout` 终止与 Telegram 的会话，并删除本地存储的会话数据。退出后，使用 tgs 需要重新运行 `tgs login`。

## 参数

| 参数 | 简写 | 类型 | 默认值 | 描述 |
|------|------|------|--------|------|
| `--profile` | `-p` | string | `default` | 要退出的账号配置文件名称 |

## 示例

### 退出默认配置文件

```bash
tgs logout
```

### 退出指定配置文件

```bash
tgs logout --profile work
```

## 相关命令

- [tgs login](/zh/reference/commands/login/) — 登录账户
- [tgs whoami](/zh/reference/commands/whoami/) — 查看当前登录账号
- [tgs profile delete](/zh/reference/commands/profile/) — 删除配置文件及其所有数据
