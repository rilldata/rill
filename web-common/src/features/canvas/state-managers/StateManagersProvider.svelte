<script lang="ts">
  import { createCanvasStateSync } from "@rilldata/web-common/features/canvas/stores/syncCanvasState";
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

  const canvasStoreReady = createCanvasStateSync(canvasName);

  setContext(DEFAULT_STORE_KEY, stateManagers);
</script>

{#if canvasStoreReady.isFetching}
  <div class="grid place-items-center size-full">
    <DelayedSpinner isLoading={canvasStoreReady.isFetching} size="40px" />
  </div>
{:else}
  <slot />
{/if}
