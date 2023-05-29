---
title: How to install Rill
sidebar_label: Install  
sidebar_position: 11
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## Releases

You can install `rill` using our installation script:

```
curl -s https://cdn.rilldata.com/install.sh | bash
```

Verify that the installation succeeded:
```
rill --help
```

## Nightly Releases

On both macOS and Linux, you can install the latest nightly build using the installation script:
```
curl -s https://cdn.rilldata.com/install.sh | bash -s -- --nightly
```

Note for macOS users: If you previously installed Rill using `brew`, the brew-managed binary will take precedent. You can remove it by running `brew uninstall rill`.


## Rill on Windows using WSL

To install Rill on Windows, you'll first need to install WSL and one dependency in your WSL environment. To install WSL, please refer to [Microsoft's documentation](https://learn.microsoft.com/en-us/windows/wsl/install).

We have verified that Rill runs on Ubuntu 22.04 LTS. Other distributions and versions may work, but are not tested. You can install Ubuntu 22.04 LTS with the following PowerShell command:

```
wsl --install -d Ubuntu-22.04

```

Once you have installed WSL and logged in to your Linux instance, you just need to install the `unzip` package to use Rill's `curl` installer. This can be done from the Linux command line with the following commands:

```
sudo apt-get update
sudo apt-get install unzip

```

With `unzip` installed, you're ready to install Rill. Just run:

```
curl -s <https://cdn.rilldata.com/install.sh> | bash

```

## Manual install

You can download platform-specific binaries from our [releases page on Github](https://github.com/rilldata/rill-developer/releases). A manual download will not make Rill Developer globally accessible, so you'll need to reference the full path of the binary when executing CLI commands.

## Frequently Asked Questions 
### Rill cannot be opened because it is from an unidentified developer.
This occurs when Rill binary is downloaded via browser. You need to change the permissions to make it executable and remove it from Apple Developer identification quarantine. 
Below CLI commands will help you to do that: 
```
chmod a+x rill
xattr -d com.apple.quarantine ./rill
```

### Error - This macOS version is not supported. Please upgrade.
Rill uses duckDB internally which requires a newer [macOS version](https://github.com/duckdb/duckdb/issues/3824). 
Please upgrade your macOS version to 10.14 or higher.