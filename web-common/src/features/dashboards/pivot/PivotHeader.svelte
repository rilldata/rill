<script lang="ts">
  import Column from "@rilldata/web-common/components/icons/Column.svelte";
  import Row from "@rilldata/web-common/components/icons/Row.svelte";
  import { splitPivotChips } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils.ts";
  import { slide } from "svelte/transition";
  import DragList from "./DragList.svelte";
  import PivotAutoArrangeZone from "./PivotAutoArrangeZone.svelte";
  import { lastNestState } from "./PivotToolbar.svelte";
  import { PivotChipType, type PivotChipData, type PivotState } from "./types";
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  export let pivotState: PivotState;
  export let setRows: (items: PivotChipData[]) => void;
  export let setColumns: (items: PivotChipData[]) => void;

  $: ({ rows, columns, tableMode } = pivotState);
  $: splitColumns = splitPivotChips(columns);
  $: fullColumns = splitColumns.dimension.concat(splitColumns.measure);
  $: isFlat = tableMode === "flat";
  $: columnsForList = isFlat ? columns : fullColumns;

  function updateColumn(items: PivotChipData[]) {
    // Reset lastNestState when columns are updated
    lastNestState.set(null);
    setColumns(items);
  }

  function updateRows(items: PivotChipData[]) {
    const filtered = items.filter(
      (item) => item.type !== PivotChipType.Measure,
    );
    setRows(filtered);
  }
</script>

<div class="header" transition:slide>
  {#if !isFlat}
    <div class="header-row" transition:slide={{ duration: 200, axis: "y" }}>
      <span class="row-label">
        <Row size="16px" />
        {m.dashboard_rows()}
      </span>
      <DragList
        zone="rows"
        placeholder={m.dashboard_drag_dimensions()}
        items={rows}
        onUpdate={updateRows}
      />
    </div>
    <PivotAutoArrangeZone
      {rows}
      columns={columnsForList}
      setRows={updateRows}
      setColumns={updateColumn}
    />
  {/if}
  <div class="header-row">
    <div class="row-label">
      <Column size="16px" />
      {m.dashboard_columns()}
    </div>

    <DragList
      zone="columns"
      {tableMode}
      items={columnsForList}
      placeholder={m.dashboard_drag_dimensions_or_measures()}
      onUpdate={updateColumn}
    />
  </div>
</div>

<style lang="postcss">
  .header {
    @apply flex flex-col border-b select-none;
    @apply bg-surface-background justify-center py-2 gap-y-2;
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
    @apply w-20 flex items-center gap-x-1 flex-shrink-0 text-fg-secondary;
  }
</style>
