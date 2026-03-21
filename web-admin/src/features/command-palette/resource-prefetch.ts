import { type QueryClient } from "@tanstack/svelte-query";
import {
  adminServiceGetProject,
  getAdminServiceGetProjectQueryKey,
} from "@rilldata/web-admin/client";
import {
  getRuntimeServiceListResourcesQueryKey,
  runtimeServiceListResources,
} from "@rilldata/web-common/runtime-client";
import { getCloudRuntimeClient } from "$lib/runtime-client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { buildRoute } from "./route-builders";
import type { SearchableItem } from "./types";

const SEARCHABLE_KINDS = new Set([
  ResourceKind.Explore,
  ResourceKind.Canvas,
  ResourceKind.Report,
  ResourceKind.Alert,
]);

const RESOURCE_KIND_TO_TYPE: Record<string, SearchableItem["type"]> = {
  [ResourceKind.Explore]: "explore",
  [ResourceKind.Canvas]: "canvas",
  [ResourceKind.Report]: "report",
  [ResourceKind.Alert]: "alert",
};

const BATCH_SIZE = 5;
const STALE_TIME = 5 * 60 * 1000;

async function fetchProjectResources(
  queryClient: QueryClient,
  orgName: string,
  projectName: string,
): Promise<SearchableItem[]> {
  try {
    const projectData = await queryClient.fetchQuery({
      queryKey: getAdminServiceGetProjectQueryKey(orgName, projectName),
      queryFn: ({ signal }) =>
        adminServiceGetProject(orgName, projectName, undefined, signal),
      staleTime: STALE_TIME,
    });

    const host = projectData.deployment?.runtimeHost;
    const instanceId = projectData.deployment?.runtimeInstanceId;
    const jwt = projectData.jwt;

    if (!host || !instanceId) return [];

    const client = getCloudRuntimeClient({
      host,
      instanceId,
      jwt: jwt ? { token: jwt } : undefined,
    });

    const resourceData = await queryClient.fetchQuery({
      queryKey: getRuntimeServiceListResourcesQueryKey(instanceId, {}),
      queryFn: ({ signal }) =>
        runtimeServiceListResources(client, {}, { signal }),
      staleTime: STALE_TIME,
    });

    if (!resourceData.resources) return [];

    return resourceData.resources
      .filter(
        (r) =>
          r.meta?.name?.kind &&
          SEARCHABLE_KINDS.has(r.meta.name.kind as ResourceKind),
      )
      .map((r) => {
        const kind = r.meta!.name!.kind! as ResourceKind;
        const name = r.meta!.name!.name!;
        const type = RESOURCE_KIND_TO_TYPE[kind];
        return {
          name,
          type,
          projectName,
          orgName,
          route: buildRoute(type, orgName, projectName, name),
        };
      });
  } catch (e) {
    console.error("[CmdK] fetchProjectResources failed for", projectName, e);
    return [];
  }
}

export async function prefetchAllResources(
  queryClient: QueryClient,
  orgName: string,
  projectNames: string[],
  onProgress: (items: SearchableItem[]) => void,
): Promise<void> {
  const allItems: SearchableItem[] = [];
  const capped = projectNames.slice(0, 20);

  for (let i = 0; i < capped.length; i += BATCH_SIZE) {
    const batch = capped.slice(i, i + BATCH_SIZE);
    const batchResults = await Promise.all(
      batch.map((name) => fetchProjectResources(queryClient, orgName, name)),
    );
    allItems.push(...batchResults.flat());
    onProgress([...allItems]);
  }
}
