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

  let dataType = Varchar;
  $: {
    if (type === "RILL_PERCENTAGE_CHANGE") {
      dataType = PercentageChange;
    } else if (type === "RILL_CHANGE") {
      dataType = MeasureChange;
    } else if (NUMERICS.has(type)) {
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
  {type}
  {value}
/>
