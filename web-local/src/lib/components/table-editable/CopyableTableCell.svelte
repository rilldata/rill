<script lang="ts">
  import { notifications } from "@rilldata/web-common/components/notifications";
  import {
    INTERVALS,
    TIMESTAMPS,
  } from "@rilldata/web-common/lib/duckdb-data-types";
  import { formatDataType } from "@rilldata/web-common/lib/formatters";
  import { fade } from "svelte/transition";
  import { createShiftClickAction } from "../../util/shift-click-action";
  import FormattedDataType from "../data-types/FormattedDataType.svelte";
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
    notifications.send({ message: `copied value to clipboard` });
    // update this to set the active animation in the tooltip text
  }}
>
  {#if value !== undefined}
    <span transition:fade|local={{ duration: 75 }}>
      <FormattedDataType {value} type={column?.type} {isNull} inTable />
    </span>
  {/if}
</button>
