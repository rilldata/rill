#!/bin/bash
skipUIBuild=false;

while getopts "su" arg; do
  case $arg in
    s) # Specify p value.
      skipUIBuild=true;;
  esac
done

if ! $skipUIBuild ; then
    ECHO "Building UI"
    make cli.prepare -C ../..
else
    ECHO "Skipping UI build"
fi
ECHO Building e2e test Rill
go build -o rill-e2e-test ../../cli/main.go