import { page } from "$app/stores";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import {
  createRuntimeServiceListResources,
  type V1ListResourcesResponse,
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
): CreateQueryResult<V1Resource[]> {
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
    let hasValidDashboard: boolean;
    if (dashboardName) {
      const dashboard = dashboards.find(
        (res) => res.meta.name.name === dashboardName,
      );
      hasValidDashboard = isValidDashboard(dashboard);
    } else {
      hasValidDashboard = dashboards.some(isValidDashboard);
    }

    if (hasValidDashboard) {
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
      select: (data) => data.resources.filter(isDashboard),
      refetchInterval: calculateRefetchInterval,
    },
  });
}

export function getDashboardToRedirect(
  orgName: string,
  projName: string,
  resources: V1Resource[],
  dashboardName: string | null,
): string | null {
  let dashboardToRedirect: V1Resource | undefined;
  if (dashboardName) {
    dashboardToRedirect = resources.find(
      (res) => res.meta.name.name === dashboardName,
    );
  } else {
    dashboardToRedirect = resources.find(isValidDashboard);
  }
  if (!dashboardToRedirect) return null;

  const resourceRoute =
    dashboardToRedirect.meta?.name?.kind === ResourceKind.Explore
      ? "explore"
      : "canvas";
  return `/${orgName}/${projName}/${resourceRoute}/${dashboardToRedirect.meta.name.name}`;
}

function isDashboard(res: V1Resource) {
  return res.canvas || res.explore;
}

function isValidDashboard(res: V1Resource) {
  return Boolean(res.canvas?.state?.validSpec || res.explore?.state?.validSpec);
}
