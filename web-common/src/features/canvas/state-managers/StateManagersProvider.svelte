<script lang="ts">
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { setContext } from "svelte";
  import { createStateManagers, DEFAULT_STORE_KEY } from "./state-managers";

  export let canvasName: string;

  const queryClient = useQueryClient();
  const stateManagers = createStateManagers({
    queryClient,
    canvasName,
  });

  $: isLoading = stateManagers.canvasEntity.spec.isLoading;

  setContext(DEFAULT_STORE_KEY, stateManagers);
</script>

{#if $isLoading}
  <div class="grid place-items-center size-full">
    <DelayedSpinner isLoading size="40px" />
  </div>
{:else}
  <slot />
{/if}
