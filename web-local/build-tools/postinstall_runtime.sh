#!/usr/bin/env bash
set -e

# Hardcoded runtime version to install
RUNTIME_VERSION="354368fd945ff064105e8afd6b7ba693673d9637"

# Targets dist/runtime as the output directory
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
OUTPUT_DIR=$SCRIPT_DIR/../dist/runtime

# Get OS and ARCH from env (if set) or as current platform
OS=${OS:-$(uname -s | tr '[:upper:]' '[:lower:]')}
ARCH=${ARCH:-$(uname -m)}
if [ $ARCH == "x86_64" ]; then
   ARCH="amd64"
fi

# Map platform to runtime release
if [ $OS == "darwin" ] && [ $ARCH == "amd64" ]; then
   TARGET="macos-amd64"
elif [ $OS == "darwin" ] && [ $ARCH == "arm64" ]; then
   TARGET="macos-arm64"
elif [ $OS == "linux" ] && [ $ARCH == "amd64" ]; then
   TARGET="linux-amd64"
elif [ $OS == "windows" ] && [ $ARCH == "amd64" ]; then
   TARGET="windows-amd64"
else
    echo "Platform not supported: os=$OS arch=$ARCH"
    exit 1
fi

# Install runtime
mkdir -p "$OUTPUT_DIR"
cd "$OUTPUT_DIR"
curl -Lso runtime.zip https://storage.googleapis.com/pkg.rilldata.com/runtime/releases/$RUNTIME_VERSION/runtime-$TARGET.zip
unzip -q -o runtime.zip
rm runtime.zip
