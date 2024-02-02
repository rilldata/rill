---
title: How to install Rill
sidebar_label: Install  
sidebar_position: 10
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## Releases

You can install `rill` using our installation script:

```bash
curl https://rill.sh | sh
```

Verify that the installation succeeded:
```bash
rill --help
```

## Nightly Releases

On both macOS and Linux, you can install the latest nightly build using the installation script:
```bash
curl -s https://rill.sh | sh -s -- --nightly
```

Note for macOS users: If you previously installed Rill using `brew`, the brew-managed binary will take precedent. You can remove it by running `brew uninstall rill`.

## Rill on Windows using WSL

To install Rill on Windows, you'll first need to install WSL and one dependency in your WSL environment. To install WSL, please refer to [Microsoft's documentation](https://learn.microsoft.com/en-us/windows/wsl/install).

We have verified that Rill runs on Ubuntu 22.04 LTS. Other distributions and versions may work, but are not tested. You can install Ubuntu 22.04 LTS with the following PowerShell command:
```bash
wsl --install -d Ubuntu-22.04
```

Once you have installed WSL and logged in to your Linux instance, you just need to install the `unzip` package to use Rill's `curl` installer. This can be done from the Linux command line with the following commands:
```bash
sudo apt-get update
sudo apt-get install unzip
```

With `unzip` installed, you're ready to install Rill. Just run:
```bash
curl -s <https://rill.sh> | sh
```

## Alternative Install Options

## Manual Install

You can download platform-specific binaries from our [releases page on Github](https://github.com/rilldata/rill/releases). A manual download will not make Rill Developer globally accessible, so you'll need to reference the full path of the binary when executing CLI commands.

## Brew Install

On macOS, you can also install Rill using Homebrew. To avoid conflicts, don't mix it with other installation options and always upgrade Rill via `brew`.
```bash
brew install rilldata/tap/rill 
```

## Frequently Asked Questions

### How do I upgrade Rill to the latest version?
If you installed Rill using the installation script described above, you can upgrade by running `rill upgrade` or by re-running the installation script.

### Rill cannot be opened because it is from an unidentified developer.
This occurs when Rill binary is downloaded via the browser. You need to change the permissions to make it executable and remove it from Apple Developer identification quarantine. 
Below CLI commands will help you to do that: 
```bash
chmod a+x rill
xattr -d com.apple.quarantine ./rill
```

### Error - This macOS version is not supported. Please upgrade.
Rill uses duckDB internally which requires a newer [macOS version](https://github.com/duckdb/duckdb/issues/3824). 
Please upgrade your macOS version to 10.14 or higher.