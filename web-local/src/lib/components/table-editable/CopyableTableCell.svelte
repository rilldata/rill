<script lang="ts">
  import FormattedDataType from "../data-types/FormattedDataType.svelte";
  import { INTERVALS, TIMESTAMPS } from "../../duckdb-data-types";
  import { formatDataType } from "../../util/formatters";
  import { createShiftClickAction } from "../../util/shift-click-action";
  import { fade } from "svelte/transition";
  import { createNotificationStore as notificationStore } from "../notifications";
  import type { ColumnConfig } from "./ColumnConfig";

  const { shiftClickAction } = createShiftClickAction();

  export let value;
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
