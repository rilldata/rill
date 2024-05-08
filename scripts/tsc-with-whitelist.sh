#!/usr/bin/env bash

whitelist="
web-admin/src/features/alerts/selectors.ts: error TS18048
web-admin/src/features/alerts/selectors.ts: error TS2345
web-admin/src/features/dashboards/listing/selectors.ts: error TS18048
web-admin/src/features/dashboards/listing/selectors.ts: error TS2322
web-admin/src/features/dashboards/listing/selectors.ts: error TS2345
web-admin/src/features/dashboards/listing/selectors.ts: error TS2769
web-admin/src/features/errors/error-utils.ts: error TS18048
web-admin/src/features/errors/error-utils.ts: error TS2322
web-admin/src/features/help/initPylonChat.ts: error TS2322
web-admin/src/features/help/initPylonWidget.ts: error TS18047
web-admin/src/features/projects/selectors.ts: error TS18048
web-admin/src/features/scheduled-reports/get-dashboard-state-for-report.ts: error TS18048
web-admin/src/features/scheduled-reports/get-dashboard-state-for-report.ts: error TS2322
web-admin/src/features/scheduled-reports/get-dashboard-state-for-report.ts: error TS2345
web-admin/src/features/scheduled-reports/get-dashboard-state-for-report.ts: error TS2769
web-admin/src/features/scheduled-reports/selectors.ts: error TS18048
web-admin/src/features/scheduled-reports/selectors.ts: error TS2345
web-admin/src/features/view-as-user/clearViewedAsUser.ts: error TS18047
web-admin/src/features/view-as-user/clearViewedAsUser.ts: error TS2322
web-admin/src/features/view-as-user/setViewedAsUser.ts: error TS2322
web-admin/src/routes/[organization]/[project]/-/dashboards/+page.ts: error TS2307
web-common/src/components/button-group/ButtonGroup.spec.ts: error TS2345
web-common/src/components/data-graphic/actions/mouse-position-to-domain-action-factory.ts: error TS2322
web-common/src/components/data-graphic/actions/outline.ts: error TS18047
web-common/src/components/data-graphic/actions/outline.ts: error TS2345
web-common/src/components/data-graphic/marks/segment.ts: error TS2345
web-common/src/components/data-graphic/utils.ts: error TS2362
web-common/src/components/data-graphic/utils.ts: error TS2363
web-common/src/components/date-picker/datetime.ts: error TS18047
web-common/src/components/date-picker/datetime.ts: error TS2322
web-common/src/components/date-picker/datetime.ts: error TS2339
web-common/src/components/date-picker/datetime.ts: error TS2345
web-common/src/components/date-picker/datetime.ts: error TS2538
web-common/src/components/date-picker/util.ts: error TS18047
web-common/src/components/editor/indent-guide/index.ts: error TS2345
web-common/src/components/editor/line-status/line-number-gutter.ts: error TS2322
web-common/src/components/editor/line-status/line-number-gutter.ts: error TS2339
web-common/src/components/editor/line-status/line-status-gutter.ts: error TS2339
web-common/src/components/editor/line-status/state.ts: error TS2322
web-common/src/components/notifications/notificationStore.ts: error TS2322
web-common/src/features/dashboards/dashboard-utils.ts: error TS18048
web-common/src/features/dashboards/dashboard-utils.ts: error TS2322
web-common/src/features/dashboards/granular-access-policies/resetSelectedMockUserAfterNavigate.ts: error TS18047
web-common/src/features/dashboards/granular-access-policies/updateDevJWT.ts: error TS2322
web-common/src/features/dashboards/granular-access-policies/useDashboardPolicyCheck.ts: error TS2345
web-common/src/features/dashboards/granular-access-policies/useMockUsers.ts: error TS2345
web-common/src/features/dashboards/pivot/util.ts: error TS18047
web-common/src/features/dashboards/pivot/util.ts: error TS2322
web-common/src/features/dashboards/proto-state/dashboard-url-state.spec.ts: error TS2345
web-common/src/features/dashboards/proto-state/dashboard-url-state.ts: error TS2345
web-common/src/features/dashboards/proto-state/toProto.ts: error TS2322
web-common/src/features/dashboards/selectors.ts: error TS18048
web-common/src/features/dashboards/selectors.ts: error TS2322
web-common/src/features/dashboards/selectors.ts: error TS2345
web-common/src/features/dashboards/selectors/index.ts: error TS2345
web-common/src/features/dashboards/show-hide-selectors.spec.ts: error TS2345
web-common/src/features/dashboards/show-hide-selectors.ts: error TS2322
web-common/src/features/dashboards/state-managers/selectors/dashboard-queries.ts: error TS2322
web-common/src/features/dashboards/state-managers/state-managers.ts: error TS2345
web-common/src/features/dashboards/stores/dashboard-store-defaults.ts: error TS18048
web-common/src/features/dashboards/stores/dashboard-store-defaults.ts: error TS2322
web-common/src/features/dashboards/stores/dashboard-store-defaults.ts: error TS2345
web-common/src/features/dashboards/stores/dashboard-store-defaults.ts: error TS2538
web-common/src/features/dashboards/stores/dashboard-store-defaults.ts: error TS2769
web-common/src/features/dashboards/stores/dashboard-stores.spec.ts: error TS2345
web-common/src/features/dashboards/stores/dashboard-stores.ts: error TS18048
web-common/src/features/dashboards/stores/dashboard-stores.ts: error TS2322
web-common/src/features/dashboards/stores/dashboard-stores.ts: error TS2345
web-common/src/features/dashboards/time-controls/time-control-store.spec.ts: error TS18048
web-common/src/features/dashboards/time-controls/time-control-store.spec.ts: error TS2322
web-common/src/features/dashboards/time-controls/time-control-store.spec.ts: error TS2345
web-common/src/features/dashboards/time-controls/time-control-store.ts: error TS18048
web-common/src/features/dashboards/time-controls/time-control-store.ts: error TS2322
web-common/src/features/dashboards/time-controls/time-control-store.ts: error TS2345
web-common/src/features/dashboards/time-controls/time-control-store.ts: error TS2769
web-common/src/features/dashboards/time-controls/time-range-store.ts: error TS18048
web-common/src/features/dashboards/time-controls/time-range-utils.ts: error TS2322
web-common/src/features/dashboards/time-series/multiple-dimension-queries.ts: error TS18048
web-common/src/features/dashboards/time-series/multiple-dimension-queries.ts: error TS2345
web-common/src/features/entity-management/file-artifacts-store.ts: error TS18048
web-common/src/features/entity-management/file-artifacts-store.ts: error TS2345
web-common/src/features/entity-management/file-artifacts-store.ts: error TS2532
web-common/src/features/entity-management/file-artifacts-store.ts: error TS2538
web-common/src/features/entity-management/resource-invalidations.ts: error TS18048
web-common/src/features/entity-management/resource-invalidations.ts: error TS2345
web-common/src/features/entity-management/resource-invalidations.ts: error TS2538
web-common/src/features/entity-management/resource-selectors.ts: error TS18048
web-common/src/features/entity-management/resource-selectors.ts: error TS2345
web-common/src/features/entity-management/resource-status-utils.ts: error TS2322
web-common/src/features/entity-management/resources-store.ts: error TS18048
web-common/src/features/entity-management/resources-store.ts: error TS2322
web-common/src/features/entity-management/resources-store.ts: error TS2345
web-common/src/features/entity-management/resources-store.ts: error TS2538
web-common/src/features/entity-management/watch-files-client.ts: error TS18048
web-common/src/features/entity-management/watch-files-client.ts: error TS2345
web-common/src/features/metrics-views/column-selectors.ts: error TS18048
web-common/src/features/metrics-views/errors.ts: error TS2322
web-common/src/features/metrics-views/errors.ts: error TS2345
web-common/src/features/metrics-views/metrics-internal-store.ts: error TS18048
web-common/src/features/metrics-views/metrics-internal-store.ts: error TS2345
web-common/src/features/metrics-views/workspace/editor/create-placeholder.ts: error TS2322
web-common/src/features/models/selectors.ts: error TS18048
web-common/src/features/models/selectors.ts: error TS2345
web-common/src/features/models/utils/embedded.ts: error TS18048
web-common/src/features/models/utils/get-table-references/index.ts: error TS18048
web-common/src/features/models/utils/get-table-references/index.ts: error TS2322
web-common/src/features/models/utils/get-table-references/index.ts: error TS2532
web-common/src/features/models/workspace/inspector/utils.ts: error TS18048
web-common/src/features/project/selectors.ts: error TS2345
web-common/src/features/project/shorthand-title/index.spec.ts: error TS2345
web-common/src/features/sources/createModel.ts: error TS2345
web-common/src/features/sources/group-uris.ts: error TS18048
web-common/src/features/sources/group-uris.ts: error TS2322
web-common/src/features/sources/modal/file-upload.ts: error TS2322
web-common/src/features/sources/modal/file-upload.ts: error TS2345
web-common/src/features/sources/selectors.ts: error TS18048
web-common/src/features/sources/selectors.ts: error TS2322
web-common/src/features/sources/selectors.ts: error TS2345
web-common/src/features/welcome/is-project-initialized.ts: error TS18048
web-common/src/layout/navigation/navigation-utils.ts: error TS2345
web-common/src/lib/actions/command-click-action.ts: error TS2345
web-common/src/lib/actions/shift-click-action.ts: error TS2345
web-common/src/lib/actions/truncate-middle-text.ts: error TS18047
web-common/src/lib/actions/truncate-middle-text.ts: error TS2345
web-common/src/lib/formatters.ts: error TS18046
web-common/src/lib/number-formatting/utils/format-with-order-of-magnitude.spec.ts: error TS2345
web-common/src/lib/store-utils/local-storage.ts: error TS2345
web-common/src/lib/time/comparisons/index.ts: error TS2322
web-common/src/lib/time/grains/index.spec.ts: error TS2345
web-common/src/lib/time/ranges/index.ts: error TS18048
web-common/src/lib/time/ranges/index.ts: error TS2345
web-common/src/lib/url-utils.ts: error TS2345
web-common/src/metrics/service/ServiceBase.ts: error TS18046
web-common/src/metrics/service/ServiceBase.ts: error TS18048
web-common/src/runtime-client/fetchWrapper.ts: error TS2345
web-common/src/runtime-client/http-request-queue/Heap.ts: error TS2322
web-common/src/runtime-client/http-request-queue/Heap.ts: error TS2345
web-common/src/runtime-client/http-request-queue/Heap.ts: error TS2538
web-common/src/runtime-client/http-request-queue/HttpRequestQueue.ts: error TS18048
web-common/src/runtime-client/http-request-queue/HttpRequestQueue.ts: error TS2345
web-common/src/runtime-client/http-request-queue/HttpRequestQueue.ts: error TS2532
web-common/src/runtime-client/http-request-queue/HttpRequestQueueTypes.ts: error TS18048
web-common/src/runtime-client/http-request-queue/HttpRequestQueueTypes.ts: error TS2322
web-common/src/runtime-client/invalidation.ts: error TS18048
web-common/src/runtime-client/invalidation.ts: error TS2345
web-common/src/runtime-client/watch-request-client.ts: error TS2322
web-common/vite.config.ts: error TS2339
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