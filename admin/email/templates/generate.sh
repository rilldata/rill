#!/usr/bin/env bash
npm install -g mjml
rm -rf admin/email/templates/gen
mkdir -p admin/email/templates/gen
mjml admin/email/templates/*.mjml -o admin/email/templates/gen/
