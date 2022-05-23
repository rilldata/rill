#!/usr/bin/env bash

echo "Building rill-developer"
npm run build

echo "Packaging rill-developer"
npm pack

echo "Publishing rill-developer"
npm publish --access public
# cleanup
rm rilldata-rill-*.tgz
