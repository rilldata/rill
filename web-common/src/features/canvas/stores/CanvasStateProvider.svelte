<script lang="ts">
  import { createCanvasStateSync } from "@rilldata/web-common/features/canvas/stores/syncCanvasState";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { getCanvasStateManagers } from "../state-managers/state-managers";

  const canvasStoreReady = createCanvasStateSync(getCanvasStateManagers());
</script>

{#if canvasStoreReady.isFetching}
  <div class="grid place-items-center size-full">
    <DelayedSpinner isLoading={canvasStoreReady.isFetching} size="40px" />
  </div>
{:else}
  <slot />
{/if}
