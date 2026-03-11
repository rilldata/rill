#!/usr/bin/env bash
# Copies shared (engine-agnostic) docs into each engine-specific docs directory.
# Run before any Docusaurus build or dev command.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DOCS_DIR="$(dirname "$SCRIPT_DIR")"
SHARED_DIR="$DOCS_DIR/shared"

for target in docs-duckdb docs-clickhouse; do
  for folder in contact guide reference; do
    dest="$DOCS_DIR/$target/$folder"
    rm -rf "$dest"
    cp -R "$SHARED_DIR/$folder" "$dest"
  done
done

echo "Shared docs copied into docs-duckdb and docs-clickhouse."
