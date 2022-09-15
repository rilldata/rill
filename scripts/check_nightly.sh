#!/usr/bin/env bash

if [[ ${TRAVIS_EVENT_TYPE} = "cron" ]]; then
  echo "Found Nightly Cronjob..."

  short_sha=$(git rev-parse --short HEAD)
  version=$(cat package.json | jq -r '.version')
  nightly="${version}-nightly-${short_sha}"

  echo "Updating nightly version to ${nightly}"
  cat package.json | jq -r '.version = "'${nightly}'"' > package.json.tmp
  mv package.json.tmp package.json
fi
