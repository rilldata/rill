# Using a Rill binary
The fastest path to installing the Rill application is using our binary. Find the link that is [relevant for your OS and architecture in our most recent release assets](https://github.com/rilldata/rill-developer/releases).

## Safely open Rill on your Mac
If you see a warning when opening the rill-macos-arm64 binary you need to change the permissions to make it executable and remove it from Apple Developer identification quarantine.
```
chmod a+x rill-macos-arm64
xattr -d com.apple.quarantine ./rill-macos-arm64
```
## Safely open Rill on Windows 10
If you see a warning "SmartScreen protected an unrecognized app from starting", [you can fix by by following these instructions here](https://www.windowscentral.com/how-fix-app-has-been-blocked-your-protection-windows-10#open).  In summary:

* Navigate to the file or program that's being blocked by SmartScreen.
* Right-click the file.
* Click Properties.
* Click the checkbox next to Unblock so that a checkmark appears.
* Click Apply.

## CLI commands
To start the application you need to open the terminal and `cd` to the directory where the application is located. You can now use Rill's [CLI](../cli.md) commands by replacing the name of the file you installed with `rill`.
```
cd downloads
rill-macos-arm64 init
rill-macos-arm64 import-source /path/to/data_1.parquet
rill-macos-arm64 start
```
