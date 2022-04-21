#!/usr/bin/env bash

echo "Building rill-developer"
npm run build

echo "Packaging rill-developer"
npm pack

echo "Publishing rill-developer"
npm publish --access public --dry-run
# cleanup
rm rilldata-rill-developer-*.tgz
