#!/usr/bin/env bash
DIR="$(dirname $0 )"

if [[ $(npm ls -g mjml | grep mjml | wc -l) -eq 0 ]]; then
  echo "Installing mjml..."
  npm install -g mjml
fi

rm -rf ${DIR}/gen
mkdir -p ${DIR}/gen
mjml ${DIR}/*.mjml -o ${DIR}/gen/
