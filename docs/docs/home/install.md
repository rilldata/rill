---
title: How to install Rill
sidebar_label: Install Rill
sidebar_position: 15
---

## Install Rill

You can install `rill` using our installation script:

```bash
curl https://rill.sh | sh
```

Verify that the installation succeeded:
```bash
rill --help
```

:::tip sharing dashboards in Rill cloud? Clone your git repo first
If you plan to share your dashboards, it is helpful to start by creating a repo in Git. Go to https://github.com/new to create a new repo. Then, run the [Rill install script](#install-rill) in your cloned location locally to make deployment easier. 

More details on deploying Rill via Git in our [Deploy section](../deploy/deploy-dashboard/).
:::

### Rill Version

You can check the current version of rill from the CLI by running:
```bash
rill version
```

### Upgrade to the newest version

To ensure you're on the latest version of Rill, you can upgrade Rill Developer easily via the command line.

```bash
rill upgrade
```

:::info What about Rill Cloud?

Rill Cloud is always on the latest and stable version of Rill Cloud. To check the latest version available, please see our [Releases](https://github.com/rilldata/rill/releases) page.

:::

## Nightly Releases

On both macOS and Linux, you can install the latest nightly build using the installation script:
```bash
curl https://rill.sh | sh -s -- --nightly
```

:::warning macOS users

If you previously installed Rill using `brew`, *the brew-managed binary will take precedent*. You can remove it by running `brew uninstall rill`.

:::

### What is nightly released
The nightly release will give you the most up-to-date version of Rill without having to wait for the official release. As these releases are not fully ready for production, you may encounter some issues.


## Installing a specific version of Rill

Rather than installing the latest version of Rill automatically, you can also install a specific version through the installation script by using the following command (e.g., `v0.40.1`):
```bash
curl https://rill.sh | sh -s -- --version <insert_version_number>
```

:::info Checking the Rill version

To check the precise version of available releases, you can navigate to the [**Releases'**](https://github.com/rilldata/rill/releases) page of our [Rill repo](https://github.com/rilldata/rill). Note that if an invalid or incorrect version is passed to the installation script, you will get prompted with an error to specify a correct version.

:::

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
:::tip Where should Rill be running? 
Please check that you are running the commands in your Linux instance not from your Windows Command Prompt. 

If you are seeing strange behavior in Rill Developer, run the following command from the CLI to see where your project files are being save `echo "$PWD"`. If they are mounted from your Windows drive, you'll need to bring them into the WSL environment. 

:::

With `unzip` installed, you're ready to install Rill. Just run:
```bash
curl https://rill.sh | sh
```

## Manual Install

You can download platform-specific binaries from our [releases page on GitHub](https://github.com/rilldata/rill/releases). A manual download will not make Rill Developer globally accessible, so you'll need to reference the full path of the binary when executing CLI commands.

## Brew Install

On macOS, you can also install Rill using Homebrew. To avoid conflicts, don't mix it with other installation options and always upgrade Rill via `brew`.
```bash
brew install rilldata/tap/rill 
```

## Uninstall Rill

To uninstall Rill, you can use the following command:
```bash
rill uninstall
```
