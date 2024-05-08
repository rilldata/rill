<script>
  /** provides the formatting for data types */
  import {
    INTERVALS,
    NUMERICS,
    TIMESTAMPS,
  } from "@rilldata/web-common/lib/duckdb-data-types";
  import Interval from "./Interval.svelte";
  import Number from "./Number.svelte";
  import Timestamp from "./Timestamp.svelte";
  import PercentageChange from "./PercentageChange.svelte";
  import MeasureChange from "./MeasureChange.svelte";
  import Varchar from "./Varchar.svelte";

  export let type = "VARCHAR";
  export let isNull = false;
  export let inTable = false;
  export let dark = false;
  export let value = undefined;
  export let customStyle = "";
  export let truncate = false;

  let dataType;
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

<!--
NOTE:
PercentageChange  and MeasureChange don't take a `type` prop,
so instantiating these directly clears a ton of warnings
about unknown props.
-->
{#if type === "RILL_PERCENTAGE_CHANGE"}
  <PercentageChange {value} {isNull} {inTable} {customStyle} {dark} />
{:else if type === "RILL_CHANGE"}
  <MeasureChange {value} {inTable} {customStyle} {dark} />
{:else}
  <svelte:component
    this={dataType}
    isNull={isNull || value === null}
    {inTable}
    {customStyle}
    {dark}
    {type}
    {value}
    {truncate}
  />
{/if}
