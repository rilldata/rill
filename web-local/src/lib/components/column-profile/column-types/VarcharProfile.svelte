<script lang="ts">
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/shift-click-action";
  import { httpRequestQueue } from "@rilldata/web-common/runtime-client/http-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
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
  export let example: any;
  export let type: string;

  export let hideRight = false;
  export let compact = false;
  export let hideNullPercentage = false;
  export let mode: "example" | "summaries" = "summaries";

  let active = false;

  let topKLimit = 15;

  $: nulls = getNullPercentage(
    $runtimeStore?.instanceId,
    objectName,
    columnName
  );

  $: columnCardinality = getCountDistinct(
    $runtimeStore?.instanceId,
    objectName,
    columnName
  );

  $: topK = getTopK($runtimeStore?.instanceId, objectName, columnName, active);

  function toggleColumnProfile() {
    active = !active;
    httpRequestQueue.prioritiseColumn(objectName, columnName, active);
  }

  $: fetchingSummaries = isFetching($nulls, $columnCardinality);
</script>

<ProfileContainer
  isFetching={fetchingSummaries}
  {active}
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
  <ColumnProfileIcon slot="icon" {type} isFetching={fetchingSummaries} />
  <svelte:fragment slot="left">{columnName}</svelte:fragment>

  <ColumnCardinalitySpark
    cardinality={$columnCardinality?.cardinality}
    {compact}
    slot="summary"
    totalRows={$columnCardinality?.totalRows}
  />
  <NullPercentageSpark
    isFetching={fetchingSummaries}
    nullCount={$nulls?.nullCount}
    slot="nullity"
    totalRows={$nulls?.totalRows}
    {type}
  />

  <div
    class="pl-10 pr-4 py-4"
    slot="details"
    style:min-height="{Math.min(topKLimit, $columnCardinality?.cardinality) *
      18 +
      42 || 250}px"
  >
    <div>
      <TopK
        topK={$topK}
        k={topKLimit}
        totalRows={$columnCardinality?.totalRows}
      />
    </div>
  </div>
</ProfileContainer>
