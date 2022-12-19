<script lang="ts">
  import { httpRequestQueue } from "@rilldata/web-common/runtime-client/http-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { copyToClipboard } from "@rilldata/web-local/lib/util/shift-click-action";
  import { DataTypeIcon } from "../../data-types";
  import ProfileContainer from "../ProfileContainer.svelte";
  import { getCountDistinct, getNullPercentage, getTopK } from "../queries";
  import TopK from "./details/TopK.svelte";
  import ColumnCardinalitySpark from "./sparks/ColumnCardinalitySpark.svelte";
  import NullPercentageSpark from "./sparks/NullPercentageSpark.svelte";

  export let columnName: string;
  export let objectName: string;
  export let example;
  export let type;

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
</script>

<ProfileContainer
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
  <DataTypeIcon slot="icon" type="VARCHAR" />
  <svelte:fragment slot="left">{columnName}</svelte:fragment>

  <ColumnCardinalitySpark
    cardinality={$columnCardinality?.cardinality}
    {compact}
    slot="summary"
    totalRows={$columnCardinality?.totalRows}
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
