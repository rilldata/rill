---
title: How to install a Rill binary
sidebar_label: Installation
sidebar_position: 20
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## macOS

On macOS, we recommend installing `rill` using Brew:

```bash
brew install rilldata/rill-developer/rill
```

Alternatively, you can install `rill` using our installation script:

```bash
curl -s https://cdn.rilldata.com/install.sh | bash
```

Verify that the installation succeeded:
```bash
rill --help
```

## Linux

On Linux, we recommend installing `rill` using the installation script:

```bash
curl -s https://cdn.rilldata.com/install.sh | bash
```

Verify that the installation succeeded:
```bash
rill --help
```

## Nightlies

On both macOS and Linux, you can install the latest nightly build using the installation script:
```bash
curl -s https://cdn.rilldata.com/install.sh | bash -s -- --nightly
```

Note for macOS users: If you previously installed Rill using Brew, the Brew-installed binary will take precedent. You can remove it by running `brew uninstall rill`.

## Manual install

You can download platform-specific binaries from our [releases page on Github](https://github.com/rilldata/rill-developer/releases).
