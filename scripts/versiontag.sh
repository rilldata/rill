#!/usr/bin/env bash
# Outputs the latest version tag from git history.
# Usage:
#   scripts/versiontag.sh         # e.g. v0.83.8
#   scripts/versiontag.sh --next  # e.g. v0.84.0
set -euo pipefail

TAG=$(git describe --tags "$(git rev-list --tags='v*' --max-count=1)")

if [[ "${1:-}" == "--next" ]]; then
  MAJOR=$(echo "$TAG" | cut -d. -f1)
  MINOR=$(echo "$TAG" | cut -d. -f2)
  echo "${MAJOR}.$((MINOR + 1)).0"
else
  echo "$TAG"
fi
