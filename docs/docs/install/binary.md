import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Using a Rill binary
Our binary script is the fastest path to installing Rill Developer. Use this command to get started:
```
curl -s https://cdn.rilldata.com/install.sh | bash
```

Nightly build:
```
curl -s [https://cdn.rilldata.com/install.sh](https://cdn.rilldata.com/install.sh) | bash -s -- --nightly
```

We don’t have nightly builds for M1 Mac (arm64). Instead, you will be using Apples’s Rosetta 2 emulation. You might notice a performance difference compared to our releases.


Alternatively you can manually download the latest binary that is relevant for your OS and architecture:

<Tabs >
  <TabItem label="MacOS" value="mac">

- [macos-arm64](https://cdn.rilldata.com/rill/latest/macos-arm64/rill) (~180mb)
- [macos-x64](https://cdn.rilldata.com/rill/latest/macos-x64/rill) (~180mb) 

## Safely open Rill on your Mac
If you see a warning when opening the rill macos-arm64 binary you need to change the permissions to make it executable and remove it from Apple Developer identification quarantine.
```
cd downloads
chmod a+x rill
xattr -d com.apple.quarantine ./rill
```

  </TabItem>
  <TabItem label="Windows" value="win">

:::note
Rill is temporarily unavailable on Windows. We're working on bringing Rill back to Windows soon.
:::

## Safely open Rill on Windows 10

If you see a warning "SmartScreen protected an unrecognized app from starting," you can fix by following the [instructions here](https://www.windowscentral.com/how-fix-app-has-been-blocked-your-protection-windows-10#open). In summary:
- Navigate to the file or program that's being blocked by SmartScreen.
- Right-click the file.
- Click Properties.
- Click the checkbox next to Unblock so that a checkmark appears.
- Click Apply.


  </TabItem>
  <TabItem label="Linux" value="linux">

[linux-x64](https://cdn.rilldata.com/rill/latest/linux-x64/rill) (~180mb)

  </TabItem>
</Tabs>

## Nuance for manually download binaries
Installing the Rill binary manually doesn't give you the ablity to use Rill globally. Instead, you should open the terminal and `cd` to the directory where the application is located. You can now use Rill's [CLI](../cli.md) commands as expected.
```
cd downloads
rill init
rill import-source /path/to/data_1.parquet
rill start
```
