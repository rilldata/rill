#!/bin/bash
skipUIBuild="${SKIP_UI_BUILD:-false}"

if ! $skipUIBuild ; then
    echo "Building UI"
    make cli.prepare -C ../
else
    echo "Skipping UI build"
fi
echo Building e2e test Rill
go build -o ./test/rill-e2e-test ../cli/main.go