<script lang="ts">
  import FormattedDataType from "$lib/components/data-types/FormattedDataType.svelte";
  import { INTERVALS, TIMESTAMPS } from "$lib/duckdb-data-types";
  import { formatDataType } from "$lib/util/formatters";
  import { createShiftClickAction } from "$lib/util/shift-click-action";
  import { fade } from "svelte/transition";
  import { createNotificationStore as notificationStore } from "$lib/components/notifications/index";
  import type { ColumnConfig } from "$lib/components/table-editor/ColumnConfig";

  const { shiftClickAction } = createShiftClickAction();

  export let value;
  export let index;
  export let column: ColumnConfig;
  export let isNull = false;
</script>

<button
  class="text-left w-full text-ellipsis overflow-hidden whitespace-nowrap"
  use:shiftClickAction
  on:shift-click={async () => {
    let exportedValue = value;
    if (INTERVALS.has(column.type)) {
      exportedValue = formatDataType(value, column.type);
    } else if (TIMESTAMPS.has(column.type)) {
      exportedValue = `TIMESTAMP '${value}'`;
    }
    await navigator.clipboard.writeText(exportedValue);
    notificationStore.send({ message: `copied value to clipboard` });
    // update this to set the active animation in the tooltip text
  }}
>
  {#if value !== undefined}
    <span transition:fade|local={{ duration: 75 }}>
      <FormattedDataType {value} type={column?.type} {isNull} inTable />
    </span>
  {/if}
</button>
