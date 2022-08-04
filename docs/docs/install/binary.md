# Using a Rill binary
Our binary is the fastest path to installing Rill Developer. Download the latest package that is relevant for your OS and architecture:

- [macos-arm64](https://storage.googleapis.com/pkg.rilldata.com/rill-developer-example/binaries/0.7/macos-arm64/rill)
- [macos-x64](https://storage.googleapis.com/pkg.rilldata.com/rill-developer-example/binaries/0.7/macos-x64/rill)
- [linux-x64](https://storage.googleapis.com/pkg.rilldata.com/rill-developer-example/binaries/0.7/linux-x64/rill)
- [win-x64](https://storage.googleapis.com/pkg.rilldata.com/rill-developer-example/binaries/0.7/win-x64/rill.exe)

## CLI commands
To start the application you need to open the terminal and `cd` to the directory where the application is located. You can now use Rill's [CLI](../cli.md) commands.
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

## Safely open Rill on Windows 10
If you see a warning "SmartScreen protected an unrecognized app from starting", [you can fix by by following these instructions here](https://www.windowscentral.com/how-fix-app-has-been-blocked-your-protection-windows-10#open).  In summary:

* Navigate to the file or program that's being blocked by SmartScreen.
* Right-click the file.
* Click Properties.
* Click the checkbox next to Unblock so that a checkmark appears.
* Click Apply.

