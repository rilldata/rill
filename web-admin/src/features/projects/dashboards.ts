import type { V1GetProjectResponse } from "@rilldata/web-admin/client";
import {
  ResourceKind,
  useFilteredResources,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import Axios from "axios";

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
