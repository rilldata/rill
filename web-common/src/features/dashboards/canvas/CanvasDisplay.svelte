<script lang="ts">
  import VegaLiteRenderer from "@rilldata/web-common/features/canvas-components/render/VegaLiteRenderer.svelte";
  import CanvasComponentSidebar from "@rilldata/web-common/features/dashboards/canvas/CanvasComponentSidebar.svelte";
  import { generateVLBarChartSpec } from "@rilldata/web-common/features/dashboards/canvas/chart/bar/bar";
  import { getChartData } from "@rilldata/web-common/features/dashboards/canvas/chart/chartQuery";
  import { chartConfig } from "@rilldata/web-common/features/dashboards/canvas/chart/configStore";
  import PivotEmpty from "@rilldata/web-common/features/dashboards/pivot/PivotEmpty.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";

  const stateManagers = getStateManagers();

  let showPanels = true;
  let viewVL;

  $: data = getChartData(stateManagers, $chartConfig);
</script>

<div class="layout">
  {#if showPanels}
    <CanvasComponentSidebar />
  {/if}
  <div class="flex flex-col size-full overflow-hidden">
    <div class="content" role="presentation">
      {#if $chartConfig.data?.x}
        <VegaLiteRenderer
          bind:viewVL
          canvasDashboard={true}
          data={{ "metrics-view": $data }}
          spec={generateVLBarChartSpec($chartConfig)}
        />
      {:else}
        <PivotEmpty assembled isFetching={false} />
      {/if}
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
