import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# How to install a Rill binary
Our install script is the fastest way to download Rill Developer and make it a globally accessible binary:
```
curl -s https://cdn.rilldata.com/install.sh | bash
```

Nightly build:
```
curl -s https://cdn.rilldata.com/install.sh | bash -s -- --nightly
```

Nightly builds for M1 Mac (arm64) use [Apples’s Rosetta emulator](https://support.apple.com/en-us/HT211861), which may be less performant than native weekly builds. 

Alternatively, you can manually download the latest binary that is relevant for your OS and architecture.

<Tabs >
  <TabItem label="MacOS" value="mac">

- [macos-arm64](https://cdn.rilldata.com/rill/latest/macos-arm64/rill) (~180mb)
- [macos-x64](https://cdn.rilldata.com/rill/latest/macos-x64/rill) (~180mb) 

If you see a warning when opening the rill macos-arm64 binary you need to change the permissions to make it executable and remove it from Apple Developer identification quarantine.
```
cd ~/Downloads
chmod a+x rill
xattr -d com.apple.quarantine ./rill
```
Unlike the script-based installation, a manual download will not make Rill Developer globally accessible, so you'll need to reference the full path of the binary when executing CLI commands.  
    
  </TabItem>
  <TabItem label="Windows" value="win">

:::note
Rill is temporarily unavailable on Windows. We're working on bringing Rill back to Windows soon.
:::


If you see a warning "SmartScreen protected an unrecognized app from starting," you can fix by following the [instructions here](https://www.windowscentral.com/how-fix-app-has-been-blocked-your-protection-windows-10#open). In summary:
- Navigate to the file or program that's being blocked by SmartScreen.
- Right-click the file.
- Click Properties.
- Click the checkbox next to Unblock so that a checkmark appears.
- Click Apply.


  </TabItem>
  <TabItem label="Linux" value="linux">

[linux-x64](https://cdn.rilldata.com/rill/latest/linux-x64/rill) (~180mb)

Unlike the script-based installation, a manual download will not make Rill Developer globally accessible, so you'll need to reference the full path of the binary when executing CLI commands.  

  </TabItem>
</Tabs>
