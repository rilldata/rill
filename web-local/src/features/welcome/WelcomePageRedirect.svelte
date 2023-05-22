<script lang="ts">
  import { goto } from "$app/navigation";
  import { createRuntimeServiceListFiles } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  // Check the directory for a rill.yaml file
  $: includesProjectFile = createRuntimeServiceListFiles(
    $runtime.instanceId,
    {
      glob: "rill.yaml",
    },
    {
      query: {
        refetchOnWindowFocus: true,
      },
    }
  );

  // If the project file does not exist, redirect to the Welcome page
  $: if (
    $includesProjectFile.isSuccess &&
    $includesProjectFile.data.paths.length === 0
  ) {
    goto("/welcome");
  }
</script>

{#if $includesProjectFile.isSuccess}
  <slot />
{/if}
