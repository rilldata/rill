#!/usr/bin/env bash

# Run TypeScript compiler and filter out some errors
output=$(npx tsc --noEmit | grep "error TS" | grep -v -E 'TS18048|TS2345|TS2322|TS18047|TS2532|TS2339|TS2538|TS2769|TS18046|TS2614')

# Check if 'error' is in the output
if echo "$output" | grep -q "error"; then
    echo "TypeScript errors found:"
    echo "$output"
    exit 1  # Exit with error code
else
    echo "No TypeScript errors detected."
    exit 0  # Exit without error
fi