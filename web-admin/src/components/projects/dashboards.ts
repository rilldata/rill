import type { V1GetProjectResponse } from "@rilldata/web-admin/client";
import {
  createRuntimeServiceListCatalogEntries,
  createRuntimeServiceListFiles,
} from "@rilldata/web-common/runtime-client";
import type { V1CatalogEntry } from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import Axios from "axios";
import { derived, get, Readable } from "svelte/store";

export interface DashboardListItem {
  name: string;
  title?: string;
  isValid: boolean;
}

export async function getDashboardsForProject(
  projectData: V1GetProjectResponse
): Promise<DashboardListItem[]> {
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

  // get all valid and invalid dashboards
  const filesRequest = axios.get(
    `/v1/instances/${projectData.prodDeployment.runtimeInstanceId}/files?glob=dashboards/*.yaml`
  );

  // get the valid dashboards
  const catalogEntriesRequest = axios.get(
    `/v1/instances/${projectData.prodDeployment.runtimeInstanceId}/catalog?type=OBJECT_TYPE_METRICS_VIEW`
  );

  const [filesResponse, catalogEntriesResponse] = await Promise.all([
    filesRequest,
    catalogEntriesRequest,
  ]);

  const filePaths = filesResponse.data?.paths;
  const catalogEntries = catalogEntriesResponse.data?.entries;

  // compose the dashboard list items
  const dashboardListItems = getDashboardListItemsFromFilesAndCatalogEntries(
    filePaths,
    catalogEntries
  );

  return dashboardListItems;
}

export function getDashboardListItemsFromFilesAndCatalogEntries(
  filePaths: string[],
  catalogEntries: V1CatalogEntry[]
): DashboardListItem[] {
  const dashboardListings = filePaths?.map((path: string) => {
    const name = path.replace("/dashboards/", "").replace(".yaml", "");
    const catalogEntry = catalogEntries?.find(
      (entry: V1CatalogEntry) => entry.path === path
    );
    const title = catalogEntry?.metricsView?.label;
    // invalid dashboards are not in the catalog
    const isValid = !!catalogEntry;
    return {
      name,
      title,
      isValid,
    };
  });

  return dashboardListings;
}

export function useDashboardListItems(
  instanceId: string,
  project: CreateQueryResult<V1GetProjectResponse>
): Readable<{
  items: DashboardListItem[];
  success: boolean;
}> {
  let isProfiling = false;
  if (project) {
    const status = get(project)?.data?.prodDeployment?.status;
    if (
      status === "DEPLOYMENT_STATUS_PENDING" ||
      status === "DEPLOYMENT_STATUS_RECONCILING"
    ) {
      isProfiling = true;
    }
  }

  return derived(
    [
      createRuntimeServiceListFiles(
        instanceId,
        {
          glob: "dashboards/*.yaml",
        },
        {
          query: {
            placeholderData: undefined,
            enabled: !isProfiling && !!project && !!instanceId,
          },
        }
      ),
      createRuntimeServiceListCatalogEntries(
        instanceId,
        {
          type: "OBJECT_TYPE_METRICS_VIEW",
        },
        {
          query: {
            placeholderData: undefined,
            enabled: !isProfiling && !!project && !!instanceId,
          },
        }
      ),
    ],
    ([dashboardFiles, dashboardCatalogEntries]) => {
      if (!dashboardFiles.isSuccess || !dashboardCatalogEntries.isSuccess)
        return {
          success: false,
          items: [],
        };

      return {
        success: true,
        items: getDashboardListItemsFromFilesAndCatalogEntries(
          dashboardFiles?.data?.paths ?? [],
          dashboardCatalogEntries?.data?.entries ?? []
        ),
      };
    }
  );
}
