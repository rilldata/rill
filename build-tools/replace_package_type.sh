#!/usr/bin/env bash

perl -pe "s/\"type\": \"$1\"/\"type\": \"$2\"/g" package.json > package-temp.json && mv package-temp.json package.json
