---
title: Environment Variables
weight: 50
---

# Environment Variables

All tgs environment variables are optional. They override compiled-in defaults or configuration file values.

## Variables

{{< snippet "env/table.md" >}}

## Profile Selection

{{< snippet "env/profile-usage.md" >}}

See [Profile resolution order]({{< relref "/guide/profiles#how-profile-resolution-works" >}}) for the full priority chain.

## API Credentials

tgs ships with API credentials compiled in. Most users do not need to set these. You may want to override them if you are building tgs from source without credentials, or if you want to use your own Telegram application credentials.

{{< snippet "env/api-usage.md" >}}

Obtain API credentials at [my.telegram.org](https://my.telegram.org).

## Custom Directories

{{< snippet "env/dirs-usage.md" >}}
