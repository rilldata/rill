#!/usr/bin/env bash
set -uo pipefail

# In CI, fail fast on first error. Locally, run exhaustively by default.
# Override with FAIL_FAST=true or FAIL_FAST=false.
if [[ -z "${FAIL_FAST:-}" ]]; then
  FAIL_FAST="${CI:-false}"
fi

if [[ "$FAIL_FAST" == "true" ]]; then
  set -e
else
  exit_code=0
fi

# This script mirrors the original GitHub Action, but can also be run locally with parity.
# In CI, ADMIN/LOCAL/COMMON are passed from dorny/paths-filter.
# Locally, if they are not set, we compute them from `git diff`.

filter_defaults_if_unset() {
  # If any of ADMIN/LOCAL/COMMON are set, assume caller is controlling behavior.
  if [[ -n "${ADMIN:-}" || -n "${LOCAL:-}" || -n "${COMMON:-}" ]]; then
    ADMIN="${ADMIN:-false}"
    LOCAL="${LOCAL:-false}"
    COMMON="${COMMON:-false}"
    return
  fi

  # Local mode: compute changes relative to a base ref.
  # BASE defaults to origin/main to match typical PR base; override as needed.
  BASE="${BASE:-origin/main}"
  HEAD="${HEAD:-HEAD}"

  git rev-parse --verify "$BASE" >/dev/null 2>&1 || git fetch --all --prune >/dev/null 2>&1 || true

  changed="$(git diff --name-only "${BASE}...${HEAD}" || true)"

  match_admin="false"
  match_local="false"
  match_common="false"

  while IFS= read -r f; do
    [[ -z "$f" ]] && continue

    # Mirrors the workflow filters exactly:
    # admin:  .github/workflows/web-test.yml OR web-admin/**
    # local:  .github/workflows/web-test.yml OR web-local/**
    # common: .github/workflows/web-test.yml OR web-common/**
    if [[ "$f" == ".github/workflows/web-test.yml" ]]; then
      match_admin="true"
      match_local="true"
      match_common="true"
      continue
    fi
    [[ "$f" == web-admin/*  ]] && match_admin="true"
    [[ "$f" == web-local/*  ]] && match_local="true"
    [[ "$f" == web-common/* ]] && match_common="true"
  done <<< "$changed"

  ADMIN="$match_admin"
  LOCAL="$match_local"
  COMMON="$match_common"
}

filter_defaults_if_unset

echo "Web code quality checks"
echo "filters: admin=$ADMIN local=$LOCAL common=$COMMON"

echo ""
echo "== NPM Install =="
# https://typicode.github.io/husky/how-to.html#ci-server-and-docker
HUSKY=0 npm install

if [[ "$COMMON" == "true" ]]; then
  echo ""
  echo "== lint and type checks for web common =="
  cd web-common
  npx svelte-kit sync
  cd ..
  npx eslint web-common --quiet || exit_code=$?
  npx svelte-check --workspace web-common --no-tsconfig --ignore "src/features/dashboards/time-series/MetricsTimeSeriesCharts.svelte,src/features/dashboards/time-series/MeasureChart.svelte,src/features/dashboards/time-controls/TimeControls.svelte,src/components/data-graphic/elements/GraphicContext.svelte,src/components/data-graphic/guides/Axis.svelte,src/components/data-graphic/guides/DynamicallyPlacedLabel.svelte,src/components/data-graphic/guides/Grid.svelte,src/components/data-graphic/compositions/timestamp-profile/TimestampDetail.svelte,src/components/data-graphic/marks/Area.svelte,src/components/data-graphic/marks/ChunkedLine.svelte,src/components/data-graphic/marks/HistogramPrimitive.svelte,src/components/data-graphic/marks/Line.svelte,src/components/data-graphic/marks/MultiMetricMouseoverLabel.svelte,src/features/column-profile/column-types/details/SummaryNumberPlot.svelte,src/stories/Tooltip.stories.svelte,src/lib/number-formatting/__stories__/NumberFormatting.stories.svelte" || exit_code=$?
fi

if [[ "$LOCAL" == "true" ]]; then
  echo ""
  echo "== lint and type checks for web local =="
  cd web-local
  npx svelte-kit sync
  cd ..
  npx eslint web-local --quiet || exit_code=$?
  npm run check -w web-local || exit_code=$?
fi

if [[ "$ADMIN" == "true" ]]; then
  echo ""
  echo "== lint and type checks for web admin =="
  cd web-admin
  npx svelte-kit sync
  cd ..
  npx eslint web-admin --quiet || exit_code=$?
  npx svelte-check --workspace web-admin --no-tsconfig || exit_code=$?
fi

echo ""
echo "== type check non-svelte files (with temporary whitelist) =="
bash ./scripts/tsc-with-whitelist.sh || exit_code=$?

# Exit with failure if any check failed (only relevant when not in fail-fast mode)
exit "${exit_code:-0}"
