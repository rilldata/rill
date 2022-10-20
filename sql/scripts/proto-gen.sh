#!/usr/bin/env bash
set -e

# Generates protocol buffer Java classes using protoc. Protoc must already be installed.
# For install instructions, see: http://google.github.io/proto-lens/installing-protoc.html.
# 
# NOTE: We generate protocol buffer files relative to the repo root (i.e. proto_path=..).
# Generating from root is necessary to cleanly cross-import in the runtime.

mkdir -p target/generated-sources/annotations
mkdir -p target/generated-sources/requests
protoc --proto_path=.. --java_out="target/generated-sources/annotations" sql/src/main/proto/ast.proto
protoc --proto_path=.. --java_out="target/generated-sources/requests" sql/src/main/proto/requests.proto
