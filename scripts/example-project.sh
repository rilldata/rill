#!/usr/bin/env bash

if [ -z ${PROJECT_BASE} ]; then
  PROJECT_BASE=".."
fi

npm run cli --silent -- init-example --project ${PROJECT_BASE}
