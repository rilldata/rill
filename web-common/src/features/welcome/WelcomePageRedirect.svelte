<script lang="ts">
  import { goto } from "$app/navigation";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useIsProjectInitialized } from "./is-project-initialized";

  // Check the directory for a rill.yaml file
  $: isProjectInitialized = useIsProjectInitialized($runtime.instanceId);

  // If the project file does not exist, redirect to the Welcome page
  $: if (
    $isProjectInitialized.isSuccess &&
    $isProjectInitialized.data === false
  ) {
    goto("/welcome");
  }
</script>

{#if $isProjectInitialized.isSuccess}
  <slot />
{/if}
