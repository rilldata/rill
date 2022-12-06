<script lang="ts">
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

  $: topK = getTopK($runtimeStore?.instanceId, objectName, columnName);
</script>

<ProfileContainer
  on:select={() => {
    active = !active;
  }}
  on:shift-click={() =>
    copyToClipboard(columnName, `copied ${columnName} to clipboard`)}
  {active}
  emphasize={active}
  {hideRight}
  {hideNullPercentage}
  {mode}
  {example}
  {type}
>
  <DataTypeIcon type="VARCHAR" slot="icon" />
  <svelte:fragment slot="left">{columnName}</svelte:fragment>

  <ColumnCardinalitySpark
    slot="summary"
    cardinality={$columnCardinality?.cardinality}
    totalRows={$columnCardinality?.totalRows}
    {compact}
  />
  <NullPercentageSpark
    slot="nullity"
    nullCount={$nulls?.nullCount}
    totalRows={$nulls?.totalRows}
    {type}
  />

  <div slot="details" class="pl-10 pr-4 py-4">
    <div>
      <TopK topK={$topK} totalRows={$columnCardinality?.totalRows} />
    </div>
  </div>
</ProfileContainer>
