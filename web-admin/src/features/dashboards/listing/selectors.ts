import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
import { useValidDashboards } from "@rilldata/web-common/features/dashboards/selectors";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";

export function useDashboardsLastUpdated(
  instanceId: string,
  organization: string,
  project: string,
) {
  return derived(
    [
      useValidDashboards(instanceId),
      createAdminServiceGetProject(organization, project),
    ],
    ([dashboardsResp, projResp]) => {
      if (!dashboardsResp.data?.length) {
        if (!projResp.data?.prodDeployment?.updatedOn) return undefined;

        // return project's last updated if there are no dashboards
        return new Date(projResp.data.prodDeployment.updatedOn);
      }

      const max = Math.max(
        ...dashboardsResp.data.map((res) =>
          new Date(res.meta.stateUpdatedOn).getTime(),
        ),
      );
      return new Date(max);
    },
  );
}

/**
 * The DashboardResource is a wrapper around a V1Resource that adds the
 * "refreshedOn" attribute, which is the last time the dashboard was refreshed.
 *
 * If the backend is updated to include this attribute in the V1Resource, this
 * wrapper can be removed.
 */
export interface DashboardResource {
  resource: V1Resource;
  refreshedOn: string;
}

function getDashboardRefreshedOn(
  dashboard: V1Resource,
  allResources: V1Resource[],
): string | undefined {
  if (!dashboard) return undefined;

  const refName = dashboard.meta.refs[0];
  const refTable = allResources.find(
    (r) => r.meta?.name?.name === refName?.name,
  );
  return (
    refTable?.model?.state.refreshedOn || refTable?.source?.state.refreshedOn
  );
}

// This iteration of `useDashboards` returns the above `DashboardResource` type, which includes `refreshedOn`
export function useDashboardsV2(
  instanceId: string,
): CreateQueryResult<DashboardResource[]> {
  return createRuntimeServiceListResources(instanceId, undefined, {
    query: {
      select: (data) => {
        // Filter for Metrics Explorers and Custom Dashboards
        const resources = data.resources.filter(
          (res) => res.metricsView || res.dashboard,
        );
        // Add `refreshedOn` to each resource
        return resources.map((resource) => {
          const refreshedOn = getDashboardRefreshedOn(resource, data.resources);
          return { resource, refreshedOn };
        });
      },
    },
  });
}

// This iteration of `useDashboard` returns the above `DashboardResource` type, which includes `refreshedOn`
export function useDashboardV2(
  instanceId: string,
  name: string,
): CreateQueryResult<DashboardResource> {
  return createRuntimeServiceListResources(instanceId, undefined, {
    query: {
      enabled: !!instanceId && !!name,
      select: (data) => {
        if (!name) return;

        const dashboard = data.resources.find(
          (res) => res.meta.name.name.toLowerCase() === name.toLowerCase(),
        );
        const refreshedOn = getDashboardRefreshedOn(dashboard, data.resources);
        return { resource: dashboard, refreshedOn };
      },
    },
  });
}
