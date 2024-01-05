<script lang="ts">
  import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import ProjectResourcesTable from "./ProjectResourcesTable.svelte";

  const resources = createRuntimeServiceListResources(
    $runtime.instanceId,
    // All resource "kinds"
    undefined,
    {
      query: {
        select: (data) => {
          // Filter out the "ProjectParser" resource
          return data.resources.filter(
            (resource) =>
              resource.meta.name.kind !== "rill.runtime.v1.ProjectParser",
          );
        },
      },
    },
  );
</script>

<section class="flex flex-col gap-y-4">
  <h2 class="text-lg font-medium">Resources</h2>
  {#if $resources.data}
    <ProjectResourcesTable resources={$resources.data} />
  {/if}
</section>
