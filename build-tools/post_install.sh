#!/usr/bin/env bash

if [ -f "./node_modules/.bin/ts-node-dev" ]; then
  ./node_modules/.bin/ts-node-dev --quiet --project tsconfig.node.json src/cli/post-install.ts
else
  node dist/cli/post-install.js
fi
