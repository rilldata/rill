<script lang="ts">
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import PublicURLsTable from "@rilldata/web-admin/features/public-urls/PublicURLsTable.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  // TODO: create a runtime service to get all public urls
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

  $: console.log("$resources.data: ", $resources.data);
</script>

<ContentContainer>
  <div class="flex flex-col w-full">
    <!-- TODO: what is the token for radix/h3? -->
    <!-- TODO: font color -->
    <h3 class="text-lg font-medium">Settings</h3>

    <!-- TODO: placeholder, to put this to the left sidebar -->
    <div class="mt-6">
      <h3 class="text-md font-medium">Public URLs</h3>
    </div>

    <div class="mt-6">
      {#if $resources.isLoading}
        <Spinner status={EntityStatus.Running} size={"16px"} />
      {:else if $resources.error}
        <div class="text-red-500">
          Error loading resources: {$resources.error?.message}
        </div>
      {:else if $resources.data}
        <PublicURLsTable resources={$resources.data} />
      {/if}
    </div>
  </div>
</ContentContainer>
