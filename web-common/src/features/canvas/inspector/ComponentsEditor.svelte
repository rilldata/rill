<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import BarChart from "@rilldata/web-common/components/icons/BarChart.svelte";
  import LineChart from "@rilldata/web-common/components/icons/LineChart.svelte";
  import StackedBar from "@rilldata/web-common/components/icons/StackedBar.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import ChartOptions from "@rilldata/web-common/features/canvas/inspector/chart/ChartOptions.svelte";
  import ComponentInputs from "@rilldata/web-common/features/canvas/inspector/ComponentInputs.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import {
    ResourceKind,
    useResource,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import SidebarWrapper from "@rilldata/web-common/features/visual-editing/SidebarWrapper.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { ArrowUp01, List, Table, Text } from "lucide-svelte";
  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();

  const { canvasStore, validSpecStore } = getCanvasStateManagers();

  const chartTypes = [
    { id: "bar", title: "Bar", icon: BarChart },
    { id: "stacked-bar", title: "Stacked Bar", icon: StackedBar },
    { id: "line", title: "Line", icon: LineChart },
  ];

  const coreComponents = [
    { id: "kpi", title: "KPI", icon: ArrowUp01 },
    { id: "table", title: "Table", icon: Table },
    { id: "text", title: "Text", icon: Text },
    { id: "leaderboard", title: "Leaderboard", icon: List },
  ];

  // TODO: fix accessor
  $: selectedComponent =
    $validSpecStore?.items?.[$canvasStore.selectedComponentIndex || 0];
  let selectedChartType;

  $: resourceQuery = useResource(
    $runtime.instanceId,
    selectedComponent?.component,
    ResourceKind.Component,
  );

  $: ({ data: componentResource } = $resourceQuery);

  $: ({ renderer, rendererProperties } =
    componentResource?.component?.spec ?? {});

  function selectChartType(chartType) {
    selectedChartType = chartType.id;
    dispatch("select", chartType);
  }
</script>

<SidebarWrapper title="Edit {renderer || 'component'} ">
  <p class="text-slate-500 text-sm">Changes below will be auto-saved.</p>
  {#if !renderer}
    <div class="section">
      <InputLabel
        label="Charts"
        id="chart-components"
        hint="Chose a chart component to add to your canvas"
      />
      <div class="chart-icons">
        {#each chartTypes as chart}
          <Tooltip distance={8} location="right">
            <Button
              square
              small
              type="secondary"
              selected={selectedChartType === chart.id}
              on:click={() => selectChartType(chart)}
            >
              <svelte:component this={chart.icon} size="20px" />
            </Button>
            <TooltipContent slot="tooltip-content">
              {chart.title}
            </TooltipContent>
          </Tooltip>
        {/each}
      </div>
      {#if selectedChartType}
        <ChartOptions chartType={selectedChartType} />
      {/if}
    </div>

    <div class="section">
      <InputLabel
        label="Core components"
        id="core-components"
        hint="Chose a core component to add to your canvas"
      />
      <div class="core-icons">
        {#each coreComponents as component}
          <Tooltip distance={8} location="right">
            <Button
              square
              small
              type="secondary"
              on:click={() => selectChartType(component)}
            >
              <svelte:component this={component.icon} size="20px" />
            </Button>
            <TooltipContent slot="tooltip-content">
              {component.title}
            </TooltipContent>
          </Tooltip>
        {/each}
      </div>
    </div>
  {:else}
    <ComponentInputs
      componentType={renderer}
      paramValues={rendererProperties}
    />
  {/if}
</SidebarWrapper>

<style lang="postcss">
  .section {
    @apply flex flex-col gap-y-2;
  }

  .chart-icons,
  .core-icons {
    @apply flex gap-x-2;
  }
</style>
