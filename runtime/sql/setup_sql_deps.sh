#!/usr/bin/env bash
set -e

if [[ $SKIP_SQL_DEPS == "true" ]]; then
  echo "Skipping SQL dependencies installation"
  exit 0
fi

# Version of librillsql to download
RILLSQL_VERSION="0.1.6"

# Get platform details
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
if [ $ARCH == "x86_64" ]; then
   ARCH="amd64"
fi

# Override platform details if GOOS and/or GOARCH are set
OS=${GOOS:-$OS}
ARCH=${GOARCH:-$ARCH}

# Map platform to librillsql release
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

# Targets runtime/sql/deps/ as the output directory
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
OUTPUT_DIR=$SCRIPT_DIR/deps/${OS}_${ARCH}
if [ -d "$OUTPUT_DIR" ] && [[ $* != -f ]]; then
   exit 0
fi

# Download librillsql
mkdir -p $OUTPUT_DIR
cd $OUTPUT_DIR
curl -Lo librillsql.zip https://storage.googleapis.com/pkg.rilldata.com/rillsql/releases/v$RILLSQL_VERSION/librillsql-$TARGET.zip
unzip -o librillsql.zip
rm librillsql.zip
