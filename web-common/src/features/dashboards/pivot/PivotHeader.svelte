<script context="module" lang="ts">
  export const lastNestState = writable<PivotRows | null>(null);
</script>

<script lang="ts">
  import { IconButton } from "@rilldata/web-common/components/button";
  import Column from "@rilldata/web-common/components/icons/Column.svelte";
  import Row from "@rilldata/web-common/components/icons/Row.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { ArrowUpDownIcon } from "lucide-svelte";
  import { writable } from "svelte/store";
  import { slide } from "svelte/transition";
  import { metricsExplorerStore } from "../stores/dashboard-stores";
  import DragList from "./DragList.svelte";
  import { PivotChipType, type PivotChipData, type PivotRows } from "./types";

  const stateManagers = getStateManagers();
  const {
    selectors: {
      pivot: { rows, columns, isFlat },
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

  /**
   * This method stores the previous nest state and passes it to
   * dashboard store when toggling back from `flat` to `nest`
   */
  function togglePivotType(newJoinState: "flat" | "nest") {
    if (newJoinState === "flat") {
      lastNestState.set($rows);
      metricsExplorerStore.setPivotRowJoinType(
        $exploreName,
        "flat",
        { dimension: [] },
        {
          measure: $columns.measure,
          dimension: [...$columns.dimension, ...$rows.dimension],
        },
      );
      return;
    }

    // Handle nest state
    const updatedRows = $lastNestState ?? { dimension: $columns.dimension };
    const rowDimensionIds = new Set(updatedRows.dimension.map((d) => d.id));

    metricsExplorerStore.setPivotRowJoinType(
      $exploreName,
      "nest",
      updatedRows,
      {
        measure: $columns.measure,
        dimension: $lastNestState
          ? $columns.dimension.filter((d) => !rowDimensionIds.has(d.id))
          : [],
      },
    );
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
        items={rowsDimensions}
        on:update={updateRows}
      />
    </div>
  {/if}
  <div class="header-row">
    <div class="row-label">
      <Column size="16px" /> Columns

      <IconButton
        marginClasses="ml-1"
        rounded
        on:click={() => togglePivotType($isFlat ? "nest" : "flat")}
      >
        <span slot="tooltip-content">{$isFlat ? "Nest" : "Flatten"} table</span>
        <ArrowUpDownIcon size="16px" />
      </IconButton>
    </div>

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
    @apply flex items-center gap-x-1 flex-shrink-0;
    width: 104px;
  }
</style>
