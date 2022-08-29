#!/usr/bin/env bash
set -e

# NOTE
# This script is a temporary hack that releases librillsql for macOS on ARM.
# It replicates the steps in .github/workflows/sql-release.yml. We'll run it
# manually until Github Actions gets a runner for macOS on ARM.

# PREREQUISITES
# - Must run on an ARM Mac
# - Must run from repo root
# - Must have gsutil (Google Cloud SDK) installed and authenticated
# - Must have upload access to the pkg.rilldata.com bucket

# Get platform details
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Check platform
if [ $OS != "darwin" ] || [ $ARCH != "arm64" ]; then
    echo "This script only runs on macOS with an arm64 chip"
    exit 1
fi

# Get output name
TARGET=macos-arm64
VERSION=$(mvn help:evaluate -Dexpression=project.version -q -DforceStdout -pl sql)

# Run native build and create archive
mvn package -Pnative-lib
rm sql/target/*.txt
zip -j librillsql-$TARGET.zip sql/target/librillsql.* sql/target/graal_isolate.*

# Upload and remove archive
gsutil cp librillsql-$TARGET.zip gs://pkg.rilldata.com/rillsql/releases/v$VERSION/
rm librillsql-$TARGET.zip
