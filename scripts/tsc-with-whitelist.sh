#!/usr/bin/env bash

whitelist="
web-admin/src/features/alerts/selectors.ts: error TS18048
web-admin/src/features/billing/issues/getMessageForTrialPlan.ts: error TS18048
web-admin/src/features/dashboards/listing/selectors.ts: error TS18048
web-admin/src/features/scheduled-reports/selectors.ts: error TS18048
web-admin/src/routes/[organization]/-/console/+layout.ts: error TS2307
web-admin/src/routes/[organization]/-/settings/+layout.ts: error TS2307
web-admin/src/routes/[organization]/-/settings/+page.ts: error TS2307
web-admin/src/routes/[organization]/-/settings/billing/+page.ts: error TS2307
web-admin/src/routes/[organization]/-/settings/billing/payment/+page.ts: error TS2307
web-admin/src/routes/[organization]/-/settings/billing/upgrade/+page.ts: error TS2307
web-admin/src/routes/[organization]/-/settings/usage/+page.ts: error TS2307
web-admin/src/routes/[organization]/-/upgrade-callback/+page.ts: error TS2307
web-admin/src/routes/[organization]/[project]/-/open-query/+page.ts: error TS2307
web-common/src/components/editor/line-status/line-number-gutter.svelte.ts: error TS2322
web-common/src/components/editor/line-status/line-number-gutter.svelte.ts: error TS2339
web-common/src/components/editor/line-status/line-status-gutter.ts: error TS2339
web-common/src/components/editor/line-status/state.ts: error TS2322
web-common/src/features/dashboards/time-controls/time-control-store.spec.ts: error TS2322
web-common/src/features/metrics-views/column-selectors.ts: error TS18048
web-common/src/runtime-client/v2/gen/connector-service.ts: error TS2707
web-common/src/runtime-client/v2/gen/runtime-service.ts: error TS2707
web-admin/src/client/gen/default/default.ts: error TS2707
web-common/src/lib/formatters.ts: error TS18046
web-common/src/lib/number-formatting/utils/format-with-order-of-magnitude.spec.ts: error TS2345
web-common/src/lib/time/comparisons/index.ts: error TS2322
web-common/src/lib/time/config.ts: error TS2322
"

# Run TypeScript compiler and find all distinct error per file
# NOTE: this is the command to run to update the whitelist above
unique_errors=$(npx tsc --noEmit | grep "error TS" | sed 's/([^()]*)//g' | sed 's/^\([^:]*:[^:]*\):.*$/\1/' | sort -u)

new_errors=$(echo "$unique_errors" | grep -v -Fx -f <(echo "$whitelist"))

# Check if 'new_errors' is not empty
if [ -n "$new_errors" ]; then
    echo "New TypeScript errors found:"
    echo "$new_errors"
    exit 1  # Exit with error code
else
    echo "No new TypeScript errors detected."
    exit 0  # Exit without error
fi
