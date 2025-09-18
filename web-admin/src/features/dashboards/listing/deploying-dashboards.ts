import { isResourceReconciling } from "@rilldata/web-admin/lib/refetch-interval-store.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import {
  createRuntimeServiceListResources,
  V1ReconcileStatus,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";

export function useDeployingDashboards(
  instanceId: string,
  orgName: string,
  projName: string,
  deploying: boolean,
  deployingDashboard: string | null,
): CreateQueryResult<{
  redirectToDashboardPath: string | null;
  dashboardsReconciling: boolean;
  dashboardsErrored: boolean;
}> {
  return createRuntimeServiceListResources(instanceId, undefined, {
    query: {
      select: (data) => {
        if (!deploying) {
          return {
            redirectToDashboardPath: null,
            dashboardsReconciling: false,
            dashboardsErrored: false,
          };
        }

        const resources = data.resources ?? [];
        const dashboards = resources.filter(isDashboard);
        const dashboard = getDashboard(dashboards, deployingDashboard) ?? null;
        const hasValidDashboard = dashboard
          ? isValidDashboard(dashboard)
          : false;

        if (!hasValidDashboard || !dashboard?.meta?.name?.name) {
          return {
            redirectToDashboardPath: null,
            dashboardsReconciling: getDashboardsReconciling(
              dashboards,
              deployingDashboard,
            ),
            dashboardsErrored: getDashboardsErrored(
              dashboards,
              deployingDashboard,
            ),
          };
        }

        const resourceRoute =
          dashboard.meta?.name?.kind === ResourceKind.Explore
            ? "explore"
            : "canvas";
        const redirectToDashboardPath = `/${orgName}/${projName}/${resourceRoute}/${dashboard.meta.name.name}`;

        return {
          redirectToDashboardPath,
          dashboardsReconciling: false,
          dashboardsErrored: false,
        };
      },
    },
  });
}

function getDashboard(
  dashboards: V1Resource[],
  dashboardName: string | null,
): V1Resource | undefined {
  let dashboard: V1Resource | undefined;
  if (dashboardName) {
    dashboard = dashboards.find(
      (res) => res.meta?.name?.name === dashboardName,
    );
  } else {
    dashboard =
      dashboards.find((res) => res.canvas?.state?.validSpec) ??
      dashboards.find((res) => res.explore?.state?.validSpec);
  }

  return dashboard;
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

function isValidDashboard(res: V1Resource) {
  return Boolean(res.canvas?.state?.validSpec || res.explore?.state?.validSpec);
}

function hasErrored(res: V1Resource) {
  return (
    res.meta?.reconcileStatus === V1ReconcileStatus.RECONCILE_STATUS_IDLE &&
    !!res.meta?.reconcileError
  );
}
