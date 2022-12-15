---
title: How to install Rill
sidebar_label: Install Rill
sidebar_position: 0
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## macOS

On macOS, we recommend installing `rill` using Homebrew:

```
brew install rilldata/tap/rill
```

Alternatively, you can install `rill` using our installation script:

```
curl -s https://cdn.rilldata.com/install.sh | bash
```

Verify that the installation succeeded:
```
rill --help
```

## Linux

On Linux, we recommend installing `rill` using the installation script:

```
curl -s https://cdn.rilldata.com/install.sh | bash
```

Verify that the installation succeeded:
```
rill --help
```

## Nightlies

On both macOS and Linux, you can install the latest nightly build using the installation script:
```
curl -s https://cdn.rilldata.com/install.sh | bash -s -- --nightly
```

Note for macOS users: If you previously installed Rill using `brew`, the brew-managed binary will take precedent. You can remove it by running `brew uninstall rill`.

## Manual install

You can download platform-specific binaries from our [releases page on Github](https://github.com/rilldata/rill-developer/releases). A manual download will not make Rill Developer globally accessible, so you'll need to reference the full path of the binary when executing CLI commands.

### macOS
If you see a warning when opening the Rill macos-arm64 binary you need to change the permissions to make it executable and remove it from Apple Developer identification quarantine.

```
chmod a+x rill
xattr -d com.apple.quarantine ./rill
```