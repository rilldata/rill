import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import {
  createRuntimeServiceListResources,
  V1ReconcileStatus,
  type V1ListResourcesResponse,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { ConnectError } from "@connectrpc/connect";
import type { CreateQueryResult, Query } from "@tanstack/svelte-query";
import {
  isResourceReconciling,
  MAX_REFETCH_INTERVAL,
  smartRefetchIntervalFunc,
} from "@rilldata/web-admin/lib/refetch-interval-store.ts";

export function useDeployingDashboards(
  client: RuntimeClient,
  orgName: string,
  projName: string,
  deployingDashboard: string | null,
  preCommitSha: string | null,
): CreateQueryResult<{
  redirectPath: string | null;
  dashboardsErrored: boolean;
}> {
  return createRuntimeServiceListResources(
    client,
    {},
    {
      query: {
        select: (data) => {
          const resources = data.resources ?? [];

          // Wait until prod's project parser has advanced past the
          // pre-merge SHA. Until then, the runtime hasn't pulled the new
          // commit, so the dashboard appears idle with stale content and
          // would redirect prematurely.
          if (preCommitSha && parserStillAtSha(resources, preCommitSha)) {
            return {
              redirectPath: null,
              dashboardsErrored: false,
            };
          }

          const dashboards = resources.filter(isDashboard);

          const reconciling = getDashboardsReconciling(
            dashboards,
            deployingDashboard,
          );
          if (reconciling) {
            return {
              redirectPath: null,
              dashboardsErrored: false,
            };
          }

          const dashboardsErrored = getDashboardsErrored(
            dashboards,
            deployingDashboard,
          );
          if (dashboardsErrored) {
            return {
              // Redirect to status page if dashboards errored
              redirectPath: `/${orgName}/${projName}/-/status`,
              dashboardsErrored,
            };
          }

          const dashboard = dashboards.find(
            (res) => res.meta?.name?.name === deployingDashboard,
          );

          // Redirect to home page if no specific dashboard was deployed
          if (!deployingDashboard || !dashboard?.meta?.name) {
            return {
              redirectPath: `/${orgName}/${projName}`,
              dashboardsErrored: false,
            };
          }

          const resourceRoute =
            dashboard.meta.name.kind === ResourceKind.Explore
              ? "explore"
              : "canvas";
          const redirectPath = `/${orgName}/${projName}/${resourceRoute}/${dashboard.meta.name.name}`;
          return {
            redirectPath,
            dashboardsErrored: false,
          };
        },
        refetchInterval: makeDeployingRefetchInterval(preCommitSha),
        enabled: Boolean(client.instanceId && orgName && projName),
      },
    },
  );
}

// Keep polling while we're still waiting for prod's parser to advance
// past the pre-merge SHA. The default smart interval stops as soon as
// nothing is reconciling, which would freeze us during the gap between
// the push and the webhook-triggered parser run.
function makeDeployingRefetchInterval(preCommitSha: string | null) {
  return (
    query: Query<
      V1ListResourcesResponse,
      ConnectError,
      V1ListResourcesResponse,
      readonly unknown[]
    >,
  ): number | false => {
    const smart = smartRefetchIntervalFunc(query);
    if (smart !== false) return smart;

    if (
      preCommitSha &&
      parserStillAtSha(query.state.data?.resources ?? [], preCommitSha)
    ) {
      return MAX_REFETCH_INTERVAL;
    }
    return false;
  };
}

function parserStillAtSha(resources: V1Resource[], preCommitSha: string) {
  const parser = resources.find((r) => r.projectParser);
  // Treat an absent SHA as "still at pre-merge" so we keep polling
  // rather than redirecting to a half-initialized runtime.
  const currentSha = parser?.projectParser?.state?.currentCommitSha ?? "";
  return currentSha === preCommitSha || currentSha === "";
}

function getDashboardsReconciling(
  dashboards: V1Resource[],
  dashboardName: string | null,
) {
  if (dashboardName) {
    const dashboard = dashboards.find(
      (res) => res.meta?.name?.name === dashboardName,
    );
    return dashboard ? isResourceReconciling(dashboard) : false;
  } else {
    return dashboards.some(isResourceReconciling);
  }
}

function getDashboardsErrored(
  dashboards: V1Resource[],
  dashboardName: string | null,
) {
  if (dashboardName) {
    const dashboard = dashboards.find(
      (res) => res.meta?.name?.name === dashboardName,
    );
    return dashboard ? hasErrored(dashboard) : false;
  } else {
    return dashboards.length > 0 && dashboards.some(hasErrored);
  }
}

function isDashboard(res: V1Resource) {
  return res.canvas || res.explore;
}

function hasErrored(res: V1Resource) {
  return (
    res.meta?.reconcileStatus === V1ReconcileStatus.RECONCILE_STATUS_IDLE &&
    !!res.meta?.reconcileError
  );
}
