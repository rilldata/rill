<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    CANVAS_CHART_CONFIG,
    type CanvasChartSpec,
  } from "@rilldata/web-common/features/canvas/components/charts";
  import type { BaseChart } from "@rilldata/web-common/features/canvas/components/charts/BaseChart";
  import { VISIBLE_CHART_TYPES } from "@rilldata/web-common/features/components/charts/config";
  import type { ChartType } from "@rilldata/web-common/features/components/charts/types";

  export let component: BaseChart<CanvasChartSpec>;

  $: ({
    parent: {
      metricsView: { getMetricsViewFromName },
    },
    chartType,
    specStore,
  } = component);

  $: _metricViewSpec = getMetricsViewFromName($specStore.metrics_view);
  $: metricsViewSpec = $_metricViewSpec.metricsView;

  $: type = $chartType;

  function selectChartType(chartType: ChartType) {
    component.updateChartType(chartType, metricsViewSpec);
  }
</script>

<div class="section">
  <InputLabel small label="Chart type" id="chart-components" />
  <div class="chart-icons">
    {#each VISIBLE_CHART_TYPES as chart, i (i)}
      <Tooltip distance={8} location="right">
        <Button
          square
          small
          type="ghost"
          label={CANVAS_CHART_CONFIG[chart].title}
          selected={type === chart}
          onClick={() => selectChartType(chart)}
        >
          <svelte:component
            this={CANVAS_CHART_CONFIG[chart].icon}
            size="20px"
          />
        </Button>
        <TooltipContent slot="tooltip-content">
          {CANVAS_CHART_CONFIG[chart].title}
        </TooltipContent>
      </Tooltip>
    {/each}
  </div>
</div>

<style lang="postcss">
  .section {
    @apply px-5 flex flex-col gap-y-2 p-2;
    @apply border-t;
  }

  .chart-icons {
    @apply flex flex-wrap border px-1 py-1 gap-x-3 gap-y-2 rounded-[2px];
  }
</style>
