---
title: 环境变量
weight: 50
---

# 环境变量

所有 tgs 环境变量均为可选。它们会覆盖编译时的默认值或配置文件中的值。

## 变量列表

{{< snippet "env/table.md" >}}

## 配置文件选择

{{< snippet "env/profile-usage.md" >}}

完整优先级链详见[配置文件解析顺序]({{< relref "/guide/profiles#how-profile-resolution-works" >}})。

## API 凭证

tgs 内置了编译时的 API 凭证，大多数用户无需设置。如果您从源码构建且未包含凭证，或者希望使用自己的 Telegram 应用凭证，可以进行覆盖。

{{< snippet "env/api-usage.md" >}}

在 [my.telegram.org](https://my.telegram.org) 获取 API 凭证。

## 自定义目录

{{< snippet "env/dirs-usage.md" >}}
