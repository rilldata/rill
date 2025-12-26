import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import {
  createRuntimeServiceListResources,
  V1ReconcileStatus,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import {
  isResourceReconciling,
  smartRefetchIntervalFunc,
} from "@rilldata/web-admin/lib/refetch-interval-store.ts";

export function useDeployingDashboards(
  instanceId: string,
  orgName: string,
  projName: string,
  deployingDashboard: string | null,
): CreateQueryResult<{
  redirectPath: string | null;
  dashboardsErrored: boolean;
}> {
  return createRuntimeServiceListResources(instanceId, undefined, {
    query: {
      select: (data) => {
        const resources = data.resources ?? [];
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
          const dashboardsErrored = getDashboardsErrored(
            dashboards,
            deployingDashboard,
          );
          return {
            // Redirect to status page is dashboards errored
            redirectPath: dashboardsErrored
              ? `/${orgName}/${projName}/-/status`
              : null,
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
      refetchInterval: smartRefetchIntervalFunc,
    },
  });
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
    return dashboards.every(hasErrored);
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
