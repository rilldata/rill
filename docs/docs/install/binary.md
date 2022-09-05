# Using a Rill binary
Our binary script is the fastest path to installing Rill Developer. Use this command to get started:
```
curl -s https://cdn.rilldata.com/install.sh | bash
```
Alternatively you can manually download the latest binary that is relevant for your OS and architecture:

- [macos-arm64](https://cdn.rilldata.com/rill/latest/macos-arm64/rill)
- [macos-x64](https://cdn.rilldata.com/rill/latest/macos-x64/rill)
- [linux-x64](https://cdn.rilldata.com/rill/latest/linux-x64/rill)
<!-- - [win-x64](https://cdn.rilldata.com/rill/latest/win-x64/rill.exe) -->

_Note: Rill is temporarily unavailable on Windows. We're working on bringing Rill back to Windows soon._

## Nuance for manually download binaries
Installing the Rill binary manually doesn't give you the ablity to use Rill globally. Instead, you should open the terminal and `cd` to the directory where the application is located. You can now use Rill's [CLI](../cli.md) commands as expected.
```
cd downloads
rill init
rill import-source /path/to/data_1.parquet
rill start
```

## Safely open Rill on your Mac
If you see a warning when opening the rill macos-arm64 binary you need to change the permissions to make it executable and remove it from Apple Developer identification quarantine.
```
cd downloads
chmod a+x rill
xattr -d com.apple.quarantine ./rill
```
<!-- 
## Safely open Rill on Windows 10
If you see a warning "SmartScreen protected an unrecognized app from starting", [you can fix by by following these instructions here](https://www.windowscentral.com/how-fix-app-has-been-blocked-your-protection-windows-10#open).  In summary:

* Navigate to the file or program that's being blocked by SmartScreen.
* Right-click the file.
* Click Properties.
* Click the checkbox next to Unblock so that a checkmark appears.
* Click Apply. -->
