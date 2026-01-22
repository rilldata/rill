<script context="module" lang="ts">
  export const lastNestState = writable<PivotChipData[] | null>(null);
</script>

<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import PivotPanel from "@rilldata/web-common/components/icons/PivotPanel.svelte";
  import { PIVOT_ROW_LIMIT_OPTIONS } from "@rilldata/web-common/features/dashboards/pivot/pivot-constants";
  import { splitPivotChips } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils.ts";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { writable } from "svelte/store";
  import Collapse from "../../../components/icons/Collapse.svelte";
  import Pivot from "../../../components/icons/Pivot.svelte";
  import Tooltip from "../../../components/tooltip/Tooltip.svelte";
  import TooltipContent from "../../../components/tooltip/TooltipContent.svelte";
  import TableIcon from "../../canvas/icons/TableIcon.svelte";
  import type { PivotChipData, PivotState, PivotTableMode } from "./types";

  export let pivotState: PivotState;
  export let showPanels = true;
  export let isFetching = false;
  export let setTableMode: (
    tableMode: PivotTableMode,
    rows: PivotChipData[],
    columns: PivotChipData[],
  ) => void;
  export let setRowLimit: (limit: number | undefined) => void;
  export let collapseAll: () => void;

  $: ({ rows, columns, tableMode, expanded, rowLimit } = pivotState);
  $: splitColumns = splitPivotChips(columns);
  $: isFlat = tableMode === "flat";

  // Row limit options - uses shared constants to ensure sync with URL validation
  const rowLimitOptions: { value: string; label: string }[] = [
    ...PIVOT_ROW_LIMIT_OPTIONS.map((limit) => ({
      value: limit.toString(),
      label: limit.toString(),
    })),
    { value: "all", label: "All" },
  ];

  // Convert rowLimit to string for Select component binding
  $: rowLimitValue = rowLimit === undefined ? "all" : rowLimit.toString();

  function handleRowLimitChange(value: string) {
    if (value === "all") {
      setRowLimit(undefined);
    } else {
      const limit = parseInt(value, 10);
      setRowLimit(limit);
    }
  }

  /**
   * This method stores the previous nest state and passes it to
   * dashboard store when toggling back from `flat` to `nest`
   */
  function togglePivotType(newJoinState: PivotTableMode) {
    if (newJoinState === "flat") {
      lastNestState.set(rows);
      setTableMode(
        "flat",
        [],
        [...splitColumns.dimension, ...rows, ...splitColumns.measure],
      );
      return;
    }

    // Handle nest state
    const updatedRows = $lastNestState ?? splitColumns.dimension;
    const rowDimensionIds = new Set(updatedRows.map((d) => d.id));

    setTableMode("nest", updatedRows, [
      ...($lastNestState
        ? splitColumns.dimension.filter((d) => !rowDimensionIds.has(d.id))
        : []),
      ...splitColumns.measure,
    ]);
  }

  function blurCurrentTarget(e: MouseEvent) {
    (e.currentTarget as HTMLButtonElement | null)?.blur();
  }
</script>

<div class="flex items-center gap-x-4 select-none pointer-events-none">
  <Tooltip location="bottom" alignment="start" distance={8}>
    <Button
      square
      type="secondary"
      theme
      selected={showPanels}
      onClick={(e) => {
        showPanels = !showPanels;
        blurCurrentTarget(e);
      }}
    >
      <PivotPanel size="18px" open={showPanels} colorClass="fill-theme-800" />
    </Button>
    <TooltipContent slot="tooltip-content">
      {showPanels ? "Hide panels" : "Show panels"}
    </TooltipContent>
  </Tooltip>

  <div class="flex items-center gap-x-1">
    <Tooltip location="bottom" alignment="start" distance={8}>
      <Button
        type="toolbar"
        onClick={() => togglePivotType(isFlat ? "nest" : "flat")}
      >
        {#if isFlat}
          <TableIcon size="16px" />
        {:else}
          <Pivot size="16px" />
        {/if}
        <span>{isFlat ? "Flat table" : "Pivot table"}</span>
      </Button>
      <TooltipContent slot="tooltip-content">
        {isFlat ? "Switch to a pivot table" : "Switch to a flat table"}
      </TooltipContent>
    </Tooltip>

    <Button
      type="toolbar"
      onClick={collapseAll}
      disabled={Object.keys(expanded).length === 0}
    >
      <Collapse size="16px" />
      Collapse All
    </Button>

    {#if !isFlat}
      <div class="flex items-center gap-x-2 pointer-events-auto">
        <Tooltip location="bottom" alignment="start" distance={8}>
          <span class="text-fg-primary pl-2">Row limit</span>
          <TooltipContent slot="tooltip-content">
            Only up to top N child rows are shown under each dimension
          </TooltipContent>
        </Tooltip>
        <Select
          id="pivot-row-limit"
          value={rowLimitValue}
          options={rowLimitOptions}
          onChange={handleRowLimitChange}
          size="sm"
          width={80}
          placeholder="Row limit"
        />
      </div>
    {/if}
    <slot name="export-menu" />

    {#if isFetching}
      <Spinner size="18px" status={EntityStatus.Running} />
    {/if}
  </div>
</div>
