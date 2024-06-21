import {
  createAdminServiceGetProject,
  V1DeploymentStatus,
} from "@rilldata/web-admin/client";
import {
  PollTimeWhenProjectDeployed,
  PollTimeWhenProjectDeploymentError,
  PollTimeWhenProjectDeploymentPending,
} from "@rilldata/web-admin/features/projects/status/selectors";
import { useValidDashboards } from "@rilldata/web-common/features/dashboards/selectors";
import { refreshResource } from "@rilldata/web-common/features/entity-management/resource-invalidations";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import {
  createRuntimeServiceListResources,
  getRuntimeServiceListResourcesQueryKey,
  runtimeServiceListResources,
  V1ReconcileStatus,
} from "@rilldata/web-common/runtime-client";
import { invalidateMetricsViewData } from "@rilldata/web-common/runtime-client/invalidation";
import type { CreateQueryResult, QueryClient } from "@tanstack/svelte-query";
import { derived } from "svelte/store";

export interface DashboardListItem {
  name: string;
  title?: string;
  description?: string;
  isValid: boolean;
}

export async function listDashboards(
  queryClient: QueryClient,
  instanceId: string,
): Promise<V1Resource[]> {
  // Fetch all resources
  const queryKey = getRuntimeServiceListResourcesQueryKey(instanceId);
  const queryFn = () => runtimeServiceListResources(instanceId);
  const resp = await queryClient.fetchQuery(queryKey, queryFn);

  // Filter for metricsViews client-side (to reduce calls to ListResources)
  const metricsViews = resp.resources.filter((res) => !!res.metricsView);

  return metricsViews;
}

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

export function useDashboardsStatus(instanceId: string) {
  return createRuntimeServiceListResources(
    instanceId,
    {
      kind: ResourceKind.MetricsView,
    },
    {
      query: {
        select: (data): V1DeploymentStatus => {
          let isPending = false;
          let isError = false;
          for (const resource of data.resources) {
            if (
              resource.meta.reconcileStatus !==
              V1ReconcileStatus.RECONCILE_STATUS_IDLE
            ) {
              isPending = true;
              continue;
            }

            if (
              resource.meta.reconcileError ||
              !resource.metricsView?.state?.validSpec
            ) {
              isError = true;
            }
          }

          if (isPending) return V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING;
          if (isError) return V1DeploymentStatus.DEPLOYMENT_STATUS_ERROR;
          return V1DeploymentStatus.DEPLOYMENT_STATUS_OK;
        },

        refetchInterval: (data) => {
          switch (data) {
            case V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING:
              return PollTimeWhenProjectDeploymentPending;

            case V1DeploymentStatus.DEPLOYMENT_STATUS_ERROR:
            case V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED:
              return PollTimeWhenProjectDeploymentError;

            case V1DeploymentStatus.DEPLOYMENT_STATUS_OK:
              return PollTimeWhenProjectDeployed;

            default:
              return PollTimeWhenProjectDeployed;
          }
        },
      },
    },
  );
}

export function listenAndInvalidateDashboards(
  queryClient: QueryClient,
  instanceId: string,
) {
  const store = derived(
    [useDashboardsStatus(instanceId), useValidDashboards(instanceId)],
    (state) => state,
  );

  const dashboards = new Map<string, Date>();

  return store.subscribe(([status, dashboardsResp]) => {
    if (
      // Let through error and ok states
      status.data === V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING ||
      status.data === V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED ||
      !dashboardsResp.data
    )
      return;

    const existingDashboards = new Set<string>();
    for (const [name] of dashboards) {
      existingDashboards.add(name);
    }

    let dashboardChanged = false;

    for (const dashboardResource of dashboardsResp.data) {
      const stateUpdatedOn = new Date(dashboardResource.meta.stateUpdatedOn);

      if (dashboards.has(dashboardResource.meta.name.name)) {
        // if the dashboard existed then check if it was updated since last seen
        const prevStateUpdatedOn = dashboards.get(
          dashboardResource.meta.name.name,
        );
        if (prevStateUpdatedOn.getTime() < stateUpdatedOn.getTime()) {
          // invalidate if it was updated
          refreshResource(queryClient, instanceId, dashboardResource);
          void invalidateMetricsViewData(queryClient, instanceId, false);
          dashboardChanged = true;
        }
      }

      if (!existingDashboards.has(dashboardResource.meta.name.name)) {
        dashboardChanged = true;
      }

      existingDashboards.delete(dashboardResource.meta.name.name);
      dashboards.set(dashboardResource.meta.name.name, stateUpdatedOn);
    }

    // cleanup of older dashboards
    for (const oldName of existingDashboards) {
      dashboards.delete(oldName);
    }

    if (dashboardChanged) {
      // Temporary to refresh useDashboardsV2 from below
      queryClient.resetQueries(
        getRuntimeServiceListResourcesQueryKey(instanceId),
      );
    }
  });
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
        const dashboard = data.resources.find(
          (res) => res.meta.name.name === name,
        );
        const refreshedOn = getDashboardRefreshedOn(dashboard, data.resources);
        return { resource: dashboard, refreshedOn };
      },
    },
  });
}
