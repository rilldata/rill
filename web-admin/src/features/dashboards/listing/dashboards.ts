import type { V1GetProjectResponse } from "@rilldata/web-admin/client";
import { V1DeploymentStatus } from "@rilldata/web-admin/client";
import {
  PollTimeDuringError,
  PollTimeDuringReconcile,
  PollTimeWhenProjectReady,
} from "@rilldata/web-admin/features/projects/selectors";
import { refreshResource } from "@rilldata/web-common/features/entity-management/resource-invalidations";
import {
  ResourceKind,
  useFilteredResources,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import {
  V1ReconcileStatus,
  createRuntimeServiceListResources,
} from "@rilldata/web-common/runtime-client";
import { fetchWrapper } from "@rilldata/web-common/runtime-client/fetchWrapper";
import { invalidateMetricsViewData } from "@rilldata/web-common/runtime-client/invalidation";
import type { QueryClient } from "@tanstack/svelte-query";
import Axios from "axios";
import { derived } from "svelte/store";

export interface DashboardListItem {
  name: string;
  title?: string;
  description?: string;
  isValid: boolean;
}

// TODO: use the creator pattern to get rid of the raw call to http endpoint
export async function getDashboardsForProject(
  projectData: V1GetProjectResponse
): Promise<V1Resource[]> {
  // There may not be a prodDeployment if the project was hibernated
  if (!projectData.prodDeployment) {
    return [];
  }

  // Hack: in development, the runtime host is actually on port 8081
  const runtimeHost = projectData.prodDeployment.runtimeHost.replace(
    "localhost:9091",
    "localhost:8081"
  );

  const axios = Axios.create({
    baseURL: runtimeHost,
    headers: {
      Authorization: `Bearer ${projectData.jwt}`,
    },
  });

  // TODO: use resource API
  const catalogEntriesResponse = await axios.get(
    `/v1/instances/${projectData.prodDeployment.runtimeInstanceId}/resources?kind=${ResourceKind.MetricsView}`
  );

  const catalogEntries = catalogEntriesResponse.data?.resources as V1Resource[];

  return catalogEntries.filter((e) => !!e.metricsView);
}

export function useDashboards(instanceId: string) {
  return useFilteredResources(instanceId, ResourceKind.MetricsView);
}

export function useDashboardsStatus(
  instanceId: string,
  project?: V1GetProjectResponse
) {
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
              return PollTimeDuringReconcile;

            case V1DeploymentStatus.DEPLOYMENT_STATUS_ERROR:
            case V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED:
              return PollTimeDuringError;

            case V1DeploymentStatus.DEPLOYMENT_STATUS_OK:
              return PollTimeWhenProjectReady;

            default:
              return PollTimeWhenProjectReady;
          }
        },

        // Do a manual call for project chip. This could be placed where `runtime` is not populated
        ...(project
          ? {
              queryFn: ({ signal }) => {
                // Hack: in development, the runtime host is actually on port 8081
                const host = project.prodDeployment.runtimeHost.replace(
                  "localhost:9091",
                  "localhost:8081"
                );
                const instanceId = project.prodDeployment.runtimeInstanceId;
                const jwt = project.jwt;
                return fetchWrapper({
                  url: `${host}/v1/instances/${instanceId}/resources?kind=${ResourceKind.MetricsView}`,
                  method: "GET",
                  ...(jwt
                    ? {
                        headers: {
                          Authorization: `Bearer ${project.jwt}`,
                          "Content-Type": "application/json",
                        },
                      }
                    : {}),
                  signal,
                });
              },
            }
          : {}),
      },
    }
  );
}

export function listenAndInvalidateDashboards(
  queryClient: QueryClient,
  instanceId: string
) {
  const store = derived(
    [useDashboardsStatus(instanceId), useDashboards(instanceId)],
    (state) => state
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

    for (const dashboardResource of dashboardsResp.data) {
      const stateUpdatedOn = new Date(dashboardResource.meta.stateUpdatedOn);

      if (dashboards.has(dashboardResource.meta.name.name)) {
        // if the dashboard existed then check if it was updated since last seen
        const prevStateUpdatedOn = dashboards.get(
          dashboardResource.meta.name.name
        );
        if (prevStateUpdatedOn.getTime() < stateUpdatedOn.getTime()) {
          // invalidate if it was updated
          refreshResource(queryClient, instanceId, dashboardResource).then(() =>
            invalidateMetricsViewData(queryClient, instanceId, false)
          );
        }
      }

      existingDashboards.delete(dashboardResource.meta.name.name);
      dashboards.set(dashboardResource.meta.name.name, stateUpdatedOn);
    }

    // cleanup of older dashboards
    for (const oldName of existingDashboards) {
      dashboards.delete(oldName);
    }
  });
}

export function useDashboardsV2(instanceId: string) {
  return createRuntimeServiceListResources(instanceId, {
    kind: ResourceKind.MetricsView,
  });
}
