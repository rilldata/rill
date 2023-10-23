import type { V1GetProjectResponse } from "@rilldata/web-admin/client";
import {
  ResourceKind,
  useFilteredResources,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  V1Resource,
  createRuntimeServiceListResources,
  type V1CatalogEntry,
  type V1MetricsView,
} from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import Axios from "axios";

export interface DashboardListItem {
  name: string;
  title?: string;
  description?: string;
  isValid: boolean;
}

export async function getDashboardsForProject(
  projectData: V1GetProjectResponse
): Promise<V1MetricsView[]> {
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

  const catalogEntriesResponse = await axios.get(
    `/v1/instances/${projectData.prodDeployment.runtimeInstanceId}/catalog?type=OBJECT_TYPE_METRICS_VIEW`
  );

  const catalogEntries = catalogEntriesResponse.data
    ?.entries as V1CatalogEntry[];

  const dashboards = catalogEntries?.map(
    (entry: V1CatalogEntry) => entry.metricsView
  );

  return dashboards;
}

export function useDashboards(instanceId: string) {
  return useFilteredResources(instanceId, ResourceKind.MetricsView);
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

export function useDashboardsV2(
  instanceId: string
): CreateQueryResult<DashboardResource[]> {
  return createRuntimeServiceListResources(instanceId, undefined, {
    query: {
      select: (data) => {
        const dashboards = data.resources.filter((res) => res.metricsView);
        return dashboards.map((db) => {
          // Extract table name from dashboard metadata
          const refName = db.meta.refs[0];
          const refTable = data.resources.find(
            (r) => r.meta?.name?.name === refName?.name
          );

          // Add the "refreshedOn" attribute
          const refreshedOn =
            refTable?.model?.state.refreshedOn ||
            refTable?.source?.state.refreshedOn;
          return { resource: db, refreshedOn };
        });
      },
    },
  });
}
