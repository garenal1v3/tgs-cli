---
title: "环境变量"
weight: 50
---

# 环境变量

tgs 支持通过环境变量进行配置，方便在 CI/CD、Docker 容器或脚本中使用。

## 变量列表

### TGS_API_ID

Telegram API 的应用 ID。

```bash
export TGS_API_ID=12345678
```

可在 [https://my.telegram.org/apps](https://my.telegram.org/apps) 申请。优先级高于配置文件中的设置。

---

### TGS_API_HASH

Telegram API 的应用哈希值。

```bash
export TGS_API_HASH=abcdef1234567890abcdef1234567890
```

与 `TGS_API_ID` 配合使用。可在 [https://my.telegram.org/apps](https://my.telegram.org/apps) 申请。

---

### TGS_PROFILE

指定默认使用的账号配置文件名称。

```bash
export TGS_PROFILE=work
```

优先级低于 `--profile` 命令行参数，高于 `.tgs.yaml` 文件配置。

---

### TGS_CONFIG_DIR

自定义配置文件存储目录。默认路径：

- **Linux**: `~/.config/tgs/`
- **macOS**: `~/Library/Application Support/tgs/`
- **Windows**: `%APPDATA%\tgs\`

```bash
export TGS_CONFIG_DIR=/custom/path/tgs/config
```

---

### TGS_DATA_DIR

自定义数据（会话文件等）存储目录。默认路径：

- **Linux**: `~/.local/share/tgs/`
- **macOS**: `~/Library/Application Support/tgs/`
- **Windows**: `%APPDATA%\tgs\`

```bash
export TGS_DATA_DIR=/custom/path/tgs/data
```

## 配置优先级

各配置来源的优先级从高到低：

1. 命令行参数（如 `--profile`、`--type`）
2. 环境变量（如 `TGS_PROFILE`）
3. `.tgs.yaml` 项目配置文件
4. 全局配置文件（`TGS_CONFIG_DIR` 中）
5. 内置默认值

## 使用示例

### 在 CI/CD 中使用

```bash
export TGS_API_ID=$SECRET_API_ID
export TGS_API_HASH=$SECRET_API_HASH
export TGS_PROFILE=ci-bot
tgs whoami
```

### 在 Docker 中使用

```dockerfile
ENV TGS_API_ID=12345678
ENV TGS_API_HASH=abcdef1234567890abcdef1234567890
ENV TGS_DATA_DIR=/data/tgs
```

## 参考

- [身份验证快速开始](/zh/getting-started/authentication/)
- [多账号管理](/zh/guide/profiles/)
