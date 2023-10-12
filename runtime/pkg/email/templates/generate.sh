#!/usr/bin/env bash
npm install -g mjml
rm -rf runtime/pkg/email/templates/gen
mkdir -p runtime/pkg/email/templates/gen
mjml runtime/pkg/email/templates/*.mjml -o runtime/pkg/email/templates/gen/
