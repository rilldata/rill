# using a binary
The fastest path to installing the application is using our binary. Find the link that is relevant for your OS and architecture in our most recent [release assets](https://github.com/rilldata/rill-developer/releases).

If you see a warning opening the rill-macos-arm64 binary you need to change the permissions of the binary to make it executable and remove it from apple developer identification quarantine.
```
chmod a+x rill-macos-arm64
xattr -d com.apple.quarantine ./rill-macos-arm64
```
