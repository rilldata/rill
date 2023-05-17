#!/bin/bash
skipUIBuild="${SKIP_UI_BUILD:-false}"

if ! $skipUIBuild ; then
    ECHO "Building UI"
    make cli.prepare -C ../
else
    ECHO "Skipping UI build"
fi
ECHO Building e2e test Rill
go build -o ./test/rill-e2e-test ../cli/main.go