---
title: Quick Start
weight: 8
---

# Quick Start

From zero to your first search in five commands. This assumes tgs is already installed — if not, see [Installation]({{< relref "/getting-started/installation" >}}).

## 1. Authenticate

The default login imports your existing session from Telegram Desktop — no phone number or code required:

{{< snippet "auth/quick-start.md" >}}

Other methods (phone + code, QR, a non-interactive flow for AI agents) are covered in [Authentication]({{< relref "/getting-started/authentication" >}}).

## 2. Confirm who you are

{{< snippet "auth/verify.md" >}}

Returns the active account as JSON — your ID, name, and username.

## 3. List your chats

See every dialog in your account — channels, groups, users, bots:

{{< snippet "quick-start/sources.md" >}}

Add `--type channel` to filter by kind, or `--with-stats` for message counts. See [Sources]({{< relref "/guide/sources" >}}).

## 4. Search inside a chat

{{< snippet "quick-start/search-chat.md" >}}

`-c` accepts a username, `t.me` link, phone number, or numeric ID. Narrow results with `--from`, `--filter`, and `--after` / `--before`.

## 5. Search across everything

Not sure which chat? Search all of them at once:

{{< snippet "quick-start/search-global.md" >}}

That's the core loop. Every command outputs JSON — pipe it into `jq` or hand it to an AI agent. Go deeper in the [Search guide]({{< relref "/guide/search" >}}).
