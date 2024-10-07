<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "../../components/button";
  import { type V1Resource, runtimeServicePutFile } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import { ResourceKind } from "../entity-management/resource-selectors";
  import {
    generateBlobForNewResourceFile,
    getPathForNewResourceFile,
  } from "../file-explorer/new-files";

  export let metricsViewResource: V1Resource | undefined;

  $: ({ instanceId } = $runtime);

  async function createExploreFromMetricsView(baseResource: V1Resource) {
    try {
      const newFilePath = getPathForNewResourceFile(
        ResourceKind.Explore,
        baseResource,
      );

      await runtimeServicePutFile(instanceId, {
        path: newFilePath,
        blob: generateBlobForNewResourceFile(
          ResourceKind.Explore,
          baseResource,
        ),
        create: true,
        createOnly: true,
      });

      await goto(`/files/${newFilePath}`);
    } catch (error) {
      console.error(error);
    }
  }
</script>

<Button
  type="primary"
  disabled={!metricsViewResource}
  on:click={() => {
    if (metricsViewResource) {
      void createExploreFromMetricsView(metricsViewResource);
    }
  }}
>
  Create Explore dashboard
</Button>
