<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import BarChart from "@rilldata/web-common/components/icons/BarChart.svelte";
  import LineChart from "@rilldata/web-common/components/icons/LineChart.svelte";
  import StackedBar from "@rilldata/web-common/components/icons/StackedBar.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    isCanvasComponentType,
    isChartComponentType,
  } from "@rilldata/web-common/features/canvas/components/util";
  import ComponentInputs from "@rilldata/web-common/features/canvas/inspector/ComponentInputs.svelte";
  import {
    ResourceKind,
    useResource,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import SidebarWrapper from "@rilldata/web-common/features/visual-editing/SidebarWrapper.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createEventDispatcher } from "svelte";

  export let selectedComponentName: string;

  const dispatch = createEventDispatcher();

  const chartTypes = [
    { id: "bar", title: "Bar", icon: BarChart },
    { id: "stacked-bar", title: "Stacked Bar", icon: StackedBar },
    { id: "line", title: "Line", icon: LineChart },
  ];

  // const coreComponents = [
  //   { id: "kpi", title: "KPI", icon: ArrowUp01 },
  //   { id: "table", title: "Table", icon: Table },
  //   { id: "text", title: "Text", icon: Text },
  //   { id: "leaderboard", title: "Leaderboard", icon: List },
  // ];

  let selectedChartType;

  // TODO: Avoid resource query if possible
  $: resourceQuery = useResource(
    $runtime.instanceId,
    selectedComponentName,
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
  {#if isChartComponentType(renderer)}
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
    </div>
  {/if}

  {#if isCanvasComponentType(renderer) && rendererProperties}
    <ComponentInputs
      componentType={renderer}
      paramValues={rendererProperties}
    />
  {:else}
    <div>
      Unknown Component {renderer}
    </div>
  {/if}
</SidebarWrapper>

<style lang="postcss">
  .section {
    @apply flex flex-col gap-y-2;
  }

  .chart-icons {
    @apply flex gap-x-2;
  }
</style>
