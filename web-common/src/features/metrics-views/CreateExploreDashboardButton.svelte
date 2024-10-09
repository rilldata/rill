<script lang="ts">
  import { goto } from "$app/navigation";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { Button } from "../../components/button";
  import { type V1Resource } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import { ResourceKind } from "../entity-management/resource-selectors";
  import { createResourceFile } from "../file-explorer/new-files";

  const queryClient = useQueryClient();

  export let metricsViewResource: V1Resource | undefined;

  $: ({ instanceId } = $runtime);

  async function handleCreateDashboard() {
    // Create the Explore file
    const newExploreFilePath = await createResourceFile(
      ResourceKind.Explore,
      metricsViewResource,
    );

    // Wait until the Explore resource is ready
    const exploreFileArtifact =
      fileArtifacts.getFileArtifact(newExploreFilePath);
    const exploreResource = exploreFileArtifact.getResource(
      queryClient,
      instanceId,
    );
    await waitUntil(() => get(exploreResource).data !== undefined);
    const newExploreName = get(exploreResource).data?.meta?.name?.name;
    if (!newExploreName) {
      throw new Error("Failed to create an Explore resource");
    }

    // Navigate to the Explore Preview
    await goto(`/explore/${newExploreName}`);
  }
</script>

<Button
  type="primary"
  disabled={!metricsViewResource}
  on:click={() => {
    if (metricsViewResource) void handleCreateDashboard();
  }}
>
  Create Explore dashboard
</Button>
