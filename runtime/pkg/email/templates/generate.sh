#!/usr/bin/env bash
npm install -g mjml
DIR="$(dirname $0 )"

rm -rf ${DIR}/gen
mkdir -p ${DIR}/gen
mjml ${DIR}/*.mjml -o ${DIR}/gen/
