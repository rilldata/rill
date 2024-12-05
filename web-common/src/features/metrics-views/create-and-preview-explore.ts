import { goto } from "$app/navigation";
import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";
import { waitUntil } from "../../lib/waitUtils";
import type { V1Resource } from "../../runtime-client";
import { fileArtifacts } from "../entity-management/file-artifacts";
import { ResourceKind } from "../entity-management/resource-selectors";
import { createResourceFile } from "../file-explorer/new-files";

export async function createAndPreviewExplore(
  queryClient: QueryClient,
  instanceId: string,
  metricsViewResource: V1Resource,
) {
  // Create the Explore file
  const filePath = await createResourceFile(
    ResourceKind.Explore,
    metricsViewResource,
  );

  // Wait until the Explore resource is ready
  const fileArtifact = fileArtifacts.getFileArtifact(filePath);
  const resource = fileArtifact.getResource(queryClient, instanceId);
  await waitUntil(() => get(resource).data !== undefined, 10000);
  const name = get(resource).data?.meta?.name?.name;
  if (!name) throw new Error("Failed to create an Explore resource");

  await goto(`/files${filePath}`);
}
