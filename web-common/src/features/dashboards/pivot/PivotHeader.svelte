<script lang="ts">
  import Column from "@rilldata/web-common/components/icons/Column.svelte";
  import Row from "@rilldata/web-common/components/icons/Row.svelte";
  import DragList from "./DragList.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";

  const stateManagers = getStateManagers();
  const {
    dashboardStore,
    selectors: {
      activeMeasure: { activeMeasure },
      dimensions: { comparisonDimension },
    },
    metricsViewName,
    runtime,
  } = stateManagers;

  $: startingMeasure = [
    {
      id: $activeMeasure?.name,
      title: $activeMeasure?.label || $activeMeasure?.name,
    },
  ];

  $: startingDimension = $comparisonDimension
    ? [
        {
          id: $comparisonDimension.column || $comparisonDimension.name,
          title:
            $comparisonDimension.label ||
            $comparisonDimension.name ||
            $comparisonDimension.column,
        },
      ]
    : [];
</script>

<div class="header">
  <div class="header-row">
    <Column size="16px" /> Columns
    <DragList items={startingMeasure} style="horizontal" />
  </div>
  <div class="header-row">
    <Row size="16px" /> Rows
    <DragList items={startingDimension} style="horizontal" />
  </div>
</div>

<style lang="postcss">
  .header {
    @apply flex flex-col;
    border-bottom: 1px solid #ddd;
  }
  .header-row {
    @apply flex items-center gap-x-1 px-2 py-1;
  }
</style>
