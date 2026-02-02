<script lang="ts">
  /** provides the formatting for data types */
  import type { PERC_DIFF } from "@rilldata/web-common/components/data-types/type-utils";
  import {
    INTERVALS,
    NUMERICS,
    TIMESTAMPS,
  } from "@rilldata/web-common/lib/duckdb-data-types";
  import type { NumberParts } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import Interval from "./Interval.svelte";
  import MeasureChange from "./MeasureChange.svelte";
  import Number from "./Number.svelte";
  import PercentageChange from "./PercentageChange.svelte";
  import Timestamp from "./Timestamp.svelte";
  import Varchar from "./Varchar.svelte";

  export let type = "VARCHAR";
  export let isNull = false;
  export let inTable = false;
  export let value:
    | string
    | boolean
    | number
    | null
    | undefined
    | NumberParts
    | PERC_DIFF = undefined;
  export let customStyle = "";
  export let truncate = false;
  export let color = "";

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
{#if type === "RILL_PERCENTAGE_CHANGE" && typeof value !== "boolean"}
  <PercentageChange {value} {isNull} {inTable} {customStyle} {color} />
{:else if type === "RILL_CHANGE" && typeof value !== "boolean"}
  <MeasureChange {value} {inTable} {customStyle} {color} />
{:else}
  <svelte:component
    this={dataType}
    isNull={isNull || value === null}
    {inTable}
    {customStyle}
    {type}
    {value}
    {truncate}
    {color}
  />
{/if}
