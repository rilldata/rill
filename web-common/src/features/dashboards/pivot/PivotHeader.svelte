<script lang="ts">
  import Column from "@rilldata/web-common/components/icons/Column.svelte";
  import Row from "@rilldata/web-common/components/icons/Row.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { slide } from "svelte/transition";
  import { metricsExplorerStore } from "../stores/dashboard-stores";
  import DragList from "./DragList.svelte";
  import { lastNestState } from "./PivotToolbar.svelte";
  import { PivotChipType, type PivotChipData } from "./types";

  const stateManagers = getStateManagers();
  const {
    selectors: {
      pivot: { rows, columns, isFlat, originalColumns },
    },
    exploreName,
  } = stateManagers;

  $: ({ dimension: columnsDimensions, measure: columnsMeasures } = $columns);

  function updateColumn(items: PivotChipData[]) {
    // Reset lastNestState when columns are updated
    lastNestState.set(null);
    metricsExplorerStore.setPivotColumns($exploreName, items);
  }

  function updateRows(items: PivotChipData[]) {
    const filtered = items.filter(
      (item) => item.type !== PivotChipType.Measure,
    );
    metricsExplorerStore.setPivotRows($exploreName, filtered);
  }
</script>

<div class="header" transition:slide>
  {#if !$isFlat}
    <div
      class="header-row"
      transition:slide={{
        duration: 200,
        axis: "y",
      }}
    >
      <span class="row-label">
        <Row size="16px" /> Rows
      </span>
      <DragList
        zone="rows"
        placeholder="Drag dimensions here"
        items={$rows}
        onUpdate={updateRows}
      />
    </div>
  {/if}
  <div class="header-row">
    <div class="row-label">
      <Column size="16px" /> Columns
    </div>

    <DragList
      zone="columns"
      tableMode={$isFlat ? "flat" : "nest"}
      items={$isFlat
        ? $originalColumns
        : columnsDimensions.concat(columnsMeasures)}
      placeholder="Drag dimensions or measures here"
      onUpdate={updateColumn}
    />
  </div>
</div>

<style lang="postcss">
  .header {
    @apply flex flex-col border-b select-none;
    @apply bg-surface justify-center py-2 gap-y-2;
    @apply flex flex-col flex-none relative overflow-hidden;
    @apply z-0;
    transition-property: height;
    will-change: height;
    @apply select-none;
  }

  .header-row {
    @apply flex items-center gap-x-2 px-2;
  }
  .row-label {
    @apply w-20 flex items-center gap-x-1 flex-shrink-0;
  }
</style>
