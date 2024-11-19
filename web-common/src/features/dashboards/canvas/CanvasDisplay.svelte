<script lang="ts">
  import VegaLiteRenderer from "@rilldata/web-common/features/canvas-components/render/VegaLiteRenderer.svelte";
  import CanvasComponentSidebar from "@rilldata/web-common/features/dashboards/canvas/CanvasComponentSidebar.svelte";
  import CanvasEmpty from "@rilldata/web-common/features/dashboards/canvas/CanvasEmpty.svelte";
  import { generateVLBarChartSpec } from "@rilldata/web-common/features/dashboards/canvas/chart/bar/bar";
  import { getChartData } from "@rilldata/web-common/features/dashboards/canvas/chart/chartQuery";
  import { chartConfig } from "@rilldata/web-common/features/dashboards/canvas/chart/configStore";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import type { View } from "svelte-vega";

  const stateManagers = getStateManagers();

  let viewVL: View;

  $: data = getChartData(stateManagers, $chartConfig);
</script>

<div class="layout">
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
        <div
          class="size-full border overflow-hidden rounded-[2px] bg-background flex flex-col items-center justify-center"
        >
          <CanvasEmpty />
        </div>
      {/if}
    </div>
  </div>
  <CanvasComponentSidebar />
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
