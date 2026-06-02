---
title: Installation
weight: 5
---

# Installation

`tgs` ships as a single static binary with no runtime dependencies. Pick whichever method fits your platform — the installed command is always `tgs`.

## Homebrew (macOS & Linux)

The quickest way on macOS and Linux:

{{< snippet "install/homebrew.md" >}}

This installs the `tgs` binary and keeps it up to date via `brew upgrade`.

## Prebuilt binaries

Grab a release archive for your platform from the [releases page](https://github.com/garenal1v3/tgs-cli/releases). Archives are named `tgs-cli_<os>_<arch>`:

| Platform | OS | Arch |
|---|---|---|
| macOS (Apple Silicon) | `darwin` | `arm64` |
| macOS (Intel) | `darwin` | `amd64` |
| Linux | `linux` | `amd64` / `arm64` |
| Windows | `windows` | `amd64` / `arm64` |

### macOS & Linux

{{< snippet "install/binary-unix.md" >}}

### Windows

{{< snippet "install/binary-windows.md" >}}

## Build from source

Requires Go 1.26 or newer:

{{< snippet "install/source.md" >}}

## Verifying the install

{{< snippet "install/verify.md" >}}

You should see a JSON object with the version, commit, and build date. Next, head to [Quick Start]({{< relref "/getting-started/quick-start" >}}).
