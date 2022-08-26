#!/usr/bin/env bash
set -e

PROTO_FILE_NAME=$1
INPUT_DIR=$2
OUTPUT_DIR=$3
echo "Generating Java protobuf classes in $OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"
protoc "$PROTO_FILE_NAME" --java_out="$OUTPUT_DIR" --proto_path="$INPUT_DIR"
echo "Protobuf builder classes generated"
