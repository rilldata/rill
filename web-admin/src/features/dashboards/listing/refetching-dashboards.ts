import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import {
  createRuntimeServiceListResources,
  type V1ListResourcesResponse,
  V1ReconcileStatus,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper.ts";
import type { CreateQueryResult, Query } from "@tanstack/svelte-query";

const InitialParseRefetchInterval = 2000;
const ParseRefetchBackoffFactor = 1.5;
const MaxParseRefetchInterval = 2000 * Math.pow(ParseRefetchBackoffFactor, 4);

const InitialDashboardsRefetchInterval = 2000;
const DashboardsRefetchBackoffFactor = 2;
const MaxDashboardsRefetchInterval =
  2000 * Math.pow(ParseRefetchBackoffFactor, 4);

export function useRefetchingDashboards(
  instanceId: string,
  dashboardName: string | null,
): CreateQueryResult<{
  dashboards: V1Resource[];
  dashboard: V1Resource | null;
  dashboardsReconciling: boolean;
  dashboardsErrored: boolean;
}> {
  let parseRefetchInterval = InitialParseRefetchInterval;
  let dashboardRefetchInterval = InitialDashboardsRefetchInterval;

  function calculateRefetchInterval(
    query: Query<V1ListResourcesResponse, HTTPError>,
  ): number | false {
    if (query.state.error) return false;
    const resources = query.state.data?.resources ?? [];
    if (resources.length <= 1) {
      parseRefetchInterval = Math.min(
        parseRefetchInterval * ParseRefetchBackoffFactor,
        MaxParseRefetchInterval,
      );

      return parseRefetchInterval;
    }

    const dashboards = resources.filter(isDashboard);
    const dashboard = getDashboard(dashboards, dashboardName);
    const hasValidDashboard = dashboard ? isValidDashboard(dashboard) : false;
    const dashboardsReconciling = !hasValidDashboard
      ? getDashboardsReconciling(dashboards, dashboardName)
      : false;
    const dashboardsErrored = !hasValidDashboard
      ? getDashboardsErrored(dashboards, dashboardName)
      : false;

    const stopRefetch =
      hasValidDashboard || !dashboardsReconciling || dashboardsErrored;
    if (stopRefetch) {
      parseRefetchInterval = InitialParseRefetchInterval;
      dashboardRefetchInterval = InitialDashboardsRefetchInterval;
      return false;
    }

    dashboardRefetchInterval = Math.min(
      dashboardRefetchInterval * DashboardsRefetchBackoffFactor,
      MaxDashboardsRefetchInterval,
    );
    return dashboardRefetchInterval;
  }

  return createRuntimeServiceListResources(instanceId, undefined, {
    query: {
      select: (data) => {
        const resources = data.resources ?? [];
        const dashboards = resources.filter(isDashboard);
        const dashboard = getDashboard(dashboards, dashboardName) ?? null;
        const hasValidDashboard = dashboard
          ? isValidDashboard(dashboard)
          : false;
        const dashboardsReconciling = !hasValidDashboard
          ? getDashboardsReconciling(dashboards, dashboardName)
          : false;
        const dashboardsErrored = !hasValidDashboard
          ? getDashboardsErrored(dashboards, dashboardName)
          : false;

        return {
          dashboards,
          dashboard,
          dashboardsReconciling,
          dashboardsErrored,
        };
      },
      refetchInterval: calculateRefetchInterval,
    },
  });
}

export function getDashboardToRedirect(
  orgName: string,
  projName: string,
  dashboard: V1Resource,
): string {
  const resourceRoute =
    dashboard.meta?.name?.kind === ResourceKind.Explore ? "explore" : "canvas";
  return `/${orgName}/${projName}/${resourceRoute}/${dashboard.meta.name.name}`;
}

function getDashboard(
  dashboards: V1Resource[],
  dashboardName: string | null,
): V1Resource | undefined {
  let dashboard: V1Resource | undefined;
  if (dashboardName) {
    dashboard = dashboards.find((res) => res.meta.name.name === dashboardName);
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
      (res) => res.meta.name.name === dashboardName,
    );
    return dashboard ? isReconciling(dashboard) : false;
  } else {
    return dashboards.some(isReconciling);
  }
}

function getDashboardsErrored(
  dashboards: V1Resource[],
  dashboardName: string | null,
) {
  if (dashboardName) {
    const dashboard = dashboards.find(
      (res) => res.meta.name.name === dashboardName,
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

function isReconciling(res: V1Resource) {
  return (
    res.meta?.reconcileStatus === V1ReconcileStatus.RECONCILE_STATUS_PENDING ||
    res.meta?.reconcileStatus === V1ReconcileStatus.RECONCILE_STATUS_RUNNING
  );
}

function hasErrored(res: V1Resource) {
  return (
    res.meta?.reconcileStatus === V1ReconcileStatus.RECONCILE_STATUS_IDLE &&
    !!res.meta?.reconcileError
  );
}
