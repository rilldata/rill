<script lang="ts">
  import CanvasComponentSidebar from "@rilldata/web-common/features/dashboards/canvas/CanvasComponentSidebar.svelte";
  import PivotEmpty from "@rilldata/web-common/features/dashboards/pivot/PivotEmpty.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";

  const stateManagers = getStateManagers();

  let showPanels = true;
</script>

<div class="layout">
  {#if showPanels}
    <CanvasComponentSidebar />
  {/if}
  <div class="flex flex-col size-full overflow-hidden">
    <div class="content" role="presentation">
      <PivotEmpty assembled isFetching={false} />
    </div>
  </div>
</div>

<style lang="postcss">
  .layout {
    @apply flex box-border h-full overflow-hidden;
  }

  .content {
    @apply flex w-full flex-col bg-slate-100 overflow-hidden size-full;
    @apply p-2 gap-y-2;
  }
</style>
