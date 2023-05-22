<script lang="ts">
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/shift-click-action";
  import {
    DATA_TYPE_COLORS,
    INTERVALS,
  } from "@rilldata/web-common/lib/duckdb-data-types";
  import { httpRequestQueue } from "../../../runtime-client/http-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import ColumnProfileIcon from "../ColumnProfileIcon.svelte";
  import ProfileContainer from "../ProfileContainer.svelte";
  import {
    getCountDistinct,
    getNullPercentage,
    getTopK,
    isFetching,
  } from "../queries";
  import TopK from "./details/TopK.svelte";
  import ColumnCardinalitySpark from "./sparks/ColumnCardinalitySpark.svelte";
  import NullPercentageSpark from "./sparks/NullPercentageSpark.svelte";

  export let columnName: string;
  export let objectName: string;
  export let type: string;
  export let mode = "summaries";
  export let example: any;

  export let hideRight = false;
  export let compact = false;
  export let hideNullPercentage = false;

  let topKLimit = 15;

  let active = false;

  $: nulls = getNullPercentage($runtime?.instanceId, objectName, columnName);

  $: columnCardinality = getCountDistinct(
    $runtime?.instanceId,
    objectName,
    columnName
  );

  $: topK = getTopK($runtime?.instanceId, objectName, columnName);

  function toggleColumnProfile() {
    active = !active;
    httpRequestQueue.prioritiseColumn(objectName, columnName, active);
  }

  $: fetchingSummaries = isFetching($nulls);
</script>

<ProfileContainer
  isFetching={fetchingSummaries}
  {active}
  {compact}
  emphasize={active}
  {example}
  {hideNullPercentage}
  {hideRight}
  {mode}
  on:select={toggleColumnProfile}
  on:shift-click={() =>
    copyToClipboard(columnName, `copied ${columnName} to clipboard`)}
  {type}
>
  <ColumnProfileIcon slot="icon" isFetching={fetchingSummaries} {type} />

  <svelte:fragment slot="left">{columnName}</svelte:fragment>

  <ColumnCardinalitySpark
    cardinality={$columnCardinality?.cardinality}
    {compact}
    slot="summary"
    totalRows={$columnCardinality?.totalRows}
    {type}
  />
  <NullPercentageSpark
    nullCount={$nulls?.nullCount}
    slot="nullity"
    totalRows={$nulls?.totalRows}
    {type}
  />
  <div
    class="pl-10 pr-4 py-4"
    slot="details"
    class:hidden={INTERVALS.has(type)}
  >
    <TopK
      {type}
      topK={$topK}
      k={topKLimit}
      totalRows={$columnCardinality?.totalRows}
      colorClass={DATA_TYPE_COLORS["STRUCT"].bgClass}
    />
  </div>
</ProfileContainer>
