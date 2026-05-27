---
title: tgs login
weight: 10
---

# tgs login

Authenticate with Telegram. The session is saved locally and reused by all subsequent commands.

## Synopsis

{{< snippet "cmd-login/synopsis.md" >}}

## Flags

{{< snippet "cmd-login/flags.md" >}}

Profile resolution order when `--profile` is not set: `TGS_PROFILE` env → `.tgs.yaml` file → `"default"`.

## Examples

{{< snippet "cmd-login/examples.md" >}}

## Auth Methods

{{< snippet "cmd-login/methods.md" >}}

## See Also

- [tgs logout]({{< relref "/reference/commands/logout" >}})
- [Authentication guide]({{< relref "/getting-started/authentication" >}})
