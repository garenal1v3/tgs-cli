---
title: "tgs profile"
weight: 30
---

# tgs profile

管理 tgs 账号配置文件。

## 用法

```
tgs profile <子命令> [参数]
```

## 子命令

| 子命令 | 描述 |
|--------|------|
| `list` | 列出所有配置文件 |
| `switch <名称>` | 切换到指定配置文件 |
| `delete <名称>` | 删除指定配置文件 |

---

## tgs profile list

列出所有已创建的配置文件。

### 用法

```
tgs profile list
```

### 示例

```bash
tgs profile list
```

示例输出：

```json
[
  {"name": "default", "active": true},
  {"name": "work", "active": false},
  {"name": "personal", "active": false}
]
```

`active` 字段表示当前活跃的配置文件。

---

## tgs profile switch

切换全局活跃配置文件。切换后，后续所有命令默认使用该配置文件（除非被 `--profile` 或 `TGS_PROFILE` 覆盖）。

### 用法

```
tgs profile switch <名称>
```

### 示例

```bash
tgs profile switch work
tgs profile switch default
```

---

## tgs profile delete

删除指定配置文件及其所有会话数据。

### 用法

```
tgs profile delete <名称>
```

### 示例

```bash
tgs profile delete work
```

> **注意**：此操作不可逆。删除后需重新运行 `tgs login --profile <名称>` 才能恢复该配置文件。

## 相关命令

- [tgs login](/zh/reference/commands/login/) — 创建或更新配置文件的登录
- [tgs whoami](/zh/reference/commands/whoami/) — 查看当前配置文件的账号信息
- [多账号管理指南](/zh/guide/profiles/)
