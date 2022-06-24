---
sidebar_position: 1
---

# using a binary
The fastest path to installing the application is using our binary. Find the link that is relevant for your OS and architecture in our most recent [release assets](https://github.com/rilldata/rill-developer/releases).

If you see a warning when opening the rill-macos-arm64 binary you need to change the permissions to make it executable and remove it from apple developer identification quarantine.
```
chmod a+x rill-macos-arm64
xattr -d com.apple.quarantine ./rill-macos-arm64
```

# CLI commands
To start the application you need to open the terminal and `cd` to the directory where the application is located. You can now use Rill's [CLI](https://github.com/rilldata/rill-developer/blob/main/docs/cli.md) commands by replacing the name of the file you installed with `rill`.
```
cd downloads
rill-macos-arm64 init
rill-macos-arm64 import-source /path/to/data_1.parquet
rill-macos-arm64 start
```
