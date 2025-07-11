<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    CHART_CONFIG,
    VISIBLE_CHART_TYPES,
    type ChartSpec,
  } from "@rilldata/web-common/features/canvas/components/charts";
  import type { BaseChart } from "@rilldata/web-common/features/canvas/components/charts/BaseChart";
  import type { ChartType } from "@rilldata/web-common/features/canvas/components/charts/types";

  export let component: BaseChart<ChartSpec>;

  $: ({
    parent: {
      spec: { getMetricsViewFromName },
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
          type="secondary"
          label={CHART_CONFIG[chart].title}
          selected={type === chart}
          onClick={() => selectChartType(chart)}
        >
          <svelte:component this={CHART_CONFIG[chart].icon} size="20px" />
        </Button>
        <TooltipContent slot="tooltip-content">
          {CHART_CONFIG[chart].title}
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
    @apply flex border px-2 py-1 gap-x-4 rounded-[2px];
  }
</style>
