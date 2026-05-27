---
title: "多账号管理"
weight: 20
---

# 多账号管理

tgs 支持多个 Telegram 账号，通过配置文件（profiles）进行管理。每个配置文件独立存储会话和设置，类似于 AWS CLI 的工作方式。

## 创建配置文件

登录时通过 `--profile` 参数指定配置文件名称：

```bash
tgs login --profile work
tgs login --profile personal
tgs login --profile client-acme
```

如果不指定，则使用默认配置文件 `default`：

```bash
tgs login  # 等同于 tgs login --profile default
```

## 切换配置文件

### 使用命令切换

```bash
tgs profile switch work
```

切换后，后续所有命令都将使用 `work` 配置文件。

### 使用 --profile 参数

在任意命令中临时指定配置文件，不影响全局设置：

```bash
tgs whoami --profile personal
tgs search "关键词" --profile work
```

## 查看所有配置文件

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

## 删除配置文件

```bash
tgs profile delete work
```

此操作会删除该配置文件的会话数据。操作不可逆，请谨慎使用。

## 通过 .tgs.yaml 配置

可以在项目目录中创建 `.tgs.yaml` 文件来指定默认配置文件。tgs 会从当前目录向上查找该文件：

```yaml
# .tgs.yaml
profile: work
```

这样在该项目目录下运行 tgs 时，会自动使用 `work` 配置文件，无需每次指定 `--profile`。

## 通过环境变量配置

设置 `TGS_PROFILE` 环境变量来指定配置文件：

```bash
export TGS_PROFILE=work
tgs whoami  # 使用 work 配置文件
```

## 配置文件优先级

当多种方式同时存在时，tgs 按以下优先级解析配置文件（从高到低）：

1. `--profile` 命令行参数
2. `TGS_PROFILE` 环境变量
3. 当前目录或父目录中的 `.tgs.yaml`
4. 默认值 `default`

## 参考

- [tgs profile 命令参考](/zh/reference/commands/profile/)
- [tgs login 命令参考](/zh/reference/commands/login/)
- [环境变量参考](/zh/reference/environment/)
