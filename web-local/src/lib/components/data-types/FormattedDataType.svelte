<script>
  /** provides the formatting for data types */
  import { formatDataType } from "@rilldata/web-common/lib/formatters";
  import { INTERVALS, NUMERICS, TIMESTAMPS } from "../../duckdb-data-types.ts";
  import Interval from "./Interval.svelte";
  import Number from "./Number.svelte";
  import Timestamp from "./Timestamp.svelte";
  import Varchar from "./Varchar.svelte";

  export let type = "VARCHAR";
  export let isNull = false;
  export let inTable = false;
  export let dark = false;
  export let value = undefined;
  export let customStyle = "";

  let dataType = Varchar;
  $: {
    if (NUMERICS.has(type)) {
      dataType = Number;
    } else if (TIMESTAMPS.has(type)) {
      dataType = Timestamp;
    } else if (INTERVALS.has(type)) {
      dataType = Interval;
    } else {
      // default to the varchar style
      dataType = Varchar;
    }
  }
</script>

<svelte:component
  this={dataType}
  isNull={isNull || value === null}
  {inTable}
  {customStyle}
  {dark}
>
  {#if value === undefined}
    <slot />
  {:else}
    {formatDataType(value, type)}
  {/if}
</svelte:component>
