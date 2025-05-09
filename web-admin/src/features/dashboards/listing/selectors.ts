import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
import { useValidExplores } from "@rilldata/web-common/features/dashboards/selectors";
import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";
import { refetchIntervalStore } from "../../shared/refetch-interval";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";

export function useDashboardsLastUpdated(
  instanceId: string,
  organization: string,
  project: string,
) {
  return derived(
    [
      useValidExplores(instanceId),
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

// This iteration of `useDashboards` returns the above `DashboardResource` type, which includes `refreshedOn`
export function useDashboardsV2(
  instanceId: string,
): CreateQueryResult<DashboardResource[], HTTPError> {
  return createRuntimeServiceListResources(instanceId, undefined, {
    query: {
      select: (data) => {
        // create a map since we are potentially looking up twice per explore
        const allResources = getMapFromArray(
          data.resources,
          (r) => `${r.meta.name.kind}_${r.meta.name.name}`,
        );
        const allDashboards: DashboardResource[] = [];
        // filter canvas dashboards
        const canvasDashboards = data.resources.filter((res) => res.canvas);
        allDashboards.push(
          ...canvasDashboards.map((resource) => {
            // Add `refreshedOn` to each resource
            const refreshedOn = getCanvasRefreshedOn(resource, allResources);
            return { resource, refreshedOn };
          }),
        );
        // filter explores
        const explores = data.resources.filter((res) => res.explore);
        allDashboards.push(
          ...explores.map((resource) => {
            // Add `refreshedOn` to each resource
            const refreshedOn = getExploreRefreshedOn(resource, allResources);
            return { resource, refreshedOn };
          }),
        );
        return allDashboards;
      },
      refetchInterval: refetchIntervalStore.calculatePollingInterval,
    },
  });
}

function getCanvasRefreshedOn(
  dashboard: V1Resource,
  allResources: Map<string, V1Resource>,
): string | undefined {
  if (!dashboard) return undefined;

  // Find the max refresh time of all the resources that are referenced by the components in the dashboard
  const maxRefresh = dashboard.meta.refs
    .map((r) => allResources.get(`${r?.kind}_${r?.name}`))
    .filter((c) => c.meta.refs.length) // filter out resources that don't have refs such as markdown and image
    .map((m) =>
      allResources.get(`${m.meta.refs[0].kind}_${m.meta.refs[0].name}`),
    )
    .map((m) => m?.metricsView?.state?.dataRefreshedOn)
    .reduce((max, c) => (c > max ? c : max));

  return maxRefresh;
}

function getExploreRefreshedOn(
  explore: V1Resource,
  allResources: Map<string, V1Resource>,
): string | undefined {
  if (!explore) return undefined;

  // First, get the metrics view for the explore
  const exploreRefName = explore.meta.refs[0];
  const metricsView = allResources.get(
    `${exploreRefName?.kind}_${exploreRefName?.name}`,
  );
  if (!metricsView) return undefined;

  // Second, get the refreshedOn from the metrics view
  return metricsView?.metricsView?.state?.dataRefreshedOn;
}
