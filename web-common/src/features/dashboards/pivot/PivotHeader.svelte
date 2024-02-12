<script lang="ts">
  import Column from "@rilldata/web-common/components/icons/Column.svelte";
  import Row from "@rilldata/web-common/components/icons/Row.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "../stores/dashboard-stores";
  import DragList from "./DragList.svelte";

  import { PivotChipType, PivotChipData } from "./types";

  const stateManagers = getStateManagers();
  const {
    selectors: {
      pivot: { rows, columns },
    },
    metricsViewName,
  } = stateManagers;

  function updateColumn(e: CustomEvent<PivotChipData[]>) {
    metricsExplorerStore.setPivotColumns($metricsViewName, e.detail);
  }

  function updateRows(e: CustomEvent<PivotChipData[]>) {
    const filtered = e.detail.filter(
      (item) => item.type !== PivotChipType.Measure,
    );
    metricsExplorerStore.setPivotRows($metricsViewName, filtered);
  }
</script>

<div class="header">
  <div class="header-row">
    <span class="row-label"> <Column size="16px" /> Columns</span>
    <DragList
      type="columns"
      removable
      items={$columns.dimension.concat($columns.measure)}
      style="horizontal"
      placeholder="Drag dimensions here"
      on:update={updateColumn}
    />
  </div>
  <div class="header-row">
    <span class="row-label"> <Row size="16px" /> Rows</span>
    <DragList
      type="rows"
      removable
      placeholder="Drag dimensions or measures here"
      on:update={updateRows}
      items={$rows.dimension}
      style="horizontal"
    />
  </div>
</div>

<style lang="postcss">
  .header {
    @apply flex flex-col;
    border-bottom: 1px solid #ddd;
    @apply bg-white py-2 px-2.5 gap-y-2;
  }
  .header-row {
    @apply flex items-center gap-x-2 px-2;
  }
  .row-label {
    @apply flex items-center gap-x-1 w-20;
  }
</style>
