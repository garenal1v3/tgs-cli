---
title: tgs whoami
weight: 20
---

# tgs whoami

Show the current profile name and account information. Reads from local storage — no Telegram connection required.

## Synopsis

{{< snippet "cmd-whoami/synopsis.md" >}}

## Flags

{{< snippet "cmd-whoami/flags.md" >}}

Profile resolution order when `--profile` is not set: `TGS_PROFILE` env → `.tgs.yaml` file → `"default"`.

## Examples

{{< snippet "cmd-whoami/examples.md" >}}

## Output

**Text output (`--output text`):**

{{< snippet "cmd-whoami/output.md" >}}

If the profile exists but is not logged in, the `user` field will be `null` (JSON) or `"Not logged in"` (text).

## See Also

- [tgs profile]({{< relref "/reference/commands/account/profile" >}})
- [Profiles guide]({{< relref "/guide/profiles" >}})
