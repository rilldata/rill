<script lang="ts">
  import Column from "@rilldata/web-common/components/icons/Column.svelte";
  import Row from "@rilldata/web-common/components/icons/Row.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "../stores/dashboard-stores";
  import DragList from "./DragList.svelte";
  import { PivotChipType, type PivotChipData } from "./types";
  import { slide } from "svelte/transition";

  const stateManagers = getStateManagers();
  const {
    selectors: {
      pivot: { rows, columns },
    },
    exploreName,
  } = stateManagers;

  $: ({ dimension: columnsDimensions, measure: columnsMeasures } = $columns);
  $: ({ dimension: rowsDimensions } = $rows);

  function updateColumn(e: CustomEvent<PivotChipData[]>) {
    metricsExplorerStore.setPivotColumns($exploreName, e.detail);
  }

  function updateRows(e: CustomEvent<PivotChipData[]>) {
    const filtered = e.detail.filter(
      (item) => item.type !== PivotChipType.Measure,
    );
    metricsExplorerStore.setPivotRows($exploreName, filtered);
  }
</script>

<div class="header" transition:slide>
  <div class="header-row">
    <span class="row-label">
      <Row size="16px" /> Rows
    </span>
    <DragList
      zone="rows"
      placeholder="Drag dimensions here"
      items={rowsDimensions}
      on:update={updateRows}
    />
  </div>
  <div class="header-row">
    <span class="row-label"> <Column size="16px" /> Columns</span>
    <DragList
      zone="columns"
      items={columnsDimensions.concat(columnsMeasures)}
      placeholder="Drag dimensions or measures here"
      on:update={updateColumn}
    />
  </div>
</div>

<style lang="postcss">
  .header {
    @apply flex flex-col border-b select-none;
    @apply bg-white justify-center py-2 gap-y-2;
    @apply flex flex-col flex-none relative overflow-hidden;
    @apply border-r z-0;
    transition-property: height;
    will-change: height;
    @apply select-none;
  }

  .header-row {
    @apply flex items-center gap-x-2 px-2;
  }
  .row-label {
    @apply flex items-center gap-x-1 w-20 flex-shrink-0;
  }
</style>
