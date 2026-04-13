import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";
import { previewModeStore } from "../../layout/preview-mode-store";
import { waitUntil } from "../../lib/waitUtils";
import type { V1Resource } from "../../runtime-client";
import type { RuntimeClient } from "../../runtime-client/v2";
import { fileArtifacts } from "../entity-management/file-artifacts";
import { ResourceKind } from "../entity-management/resource-selectors";
import { createResourceFile } from "../entity-management/add/new-files.ts";
import { navigateToExplore, navigateToFile } from "../workspaces/edit-routing";

export async function createAndPreviewExplore(
  client: RuntimeClient,
  queryClient: QueryClient,
  instanceId: string,
  metricsViewResource: V1Resource,
) {
  // Create the Explore file
  const filePath = await createResourceFile(
    client,
    ResourceKind.Explore,
    metricsViewResource,
  );

  // Wait until the Explore resource is ready
  const fileArtifact = fileArtifacts.getFileArtifact(filePath);
  const resource = fileArtifact.getResource(queryClient);

  await waitUntil(() => {
    return get(resource).data !== undefined;
  }, 10000);

  const name = get(resource).data?.meta?.name?.name;
  if (!name) throw new Error("Failed to create an Explore resource");

  const isPreview = get(previewModeStore);
  await (isPreview ? navigateToExplore(name) : navigateToFile(filePath));
}
