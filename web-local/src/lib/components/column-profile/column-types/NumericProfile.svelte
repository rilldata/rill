<script lang="ts">
  import { useRuntimeServiceGetDescriptiveStatistics } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { copyToClipboard } from "@rilldata/web-local/lib/util/shift-click-action";
  import { derived } from "svelte/store";
  import { DataTypeIcon } from "../../data-types";
  import ProfileContainer from "../ProfileContainer.svelte";
  import {
    getNullPercentage,
    getNumericHistogram,
    getRugPlotData,
    getTopK,
  } from "../queries";
  import NumericPlot from "./details/NumericPlot.svelte";
  import NullPercentageSpark from "./sparks/NullPercentageSpark.svelte";
  import NumericSpark from "./sparks/NumericSpark.svelte";
  export let columnName: string;
  export let objectName: string;
  export let type: string;
  export let mode = "summaries";
  export let example: any;

  export let hideRight = false;
  export let compact = false;
  export let hideNullPercentage = false;

  let active = false;

  $: nulls = getNullPercentage(
    $runtimeStore?.instanceId,
    objectName,
    columnName
  );

  $: numericHistogram = getNumericHistogram(
    $runtimeStore?.instanceId,
    objectName,
    columnName
  );
  $: rug = getRugPlotData($runtimeStore?.instanceId, objectName, columnName);

  $: topK = getTopK($runtimeStore?.instanceId, objectName, columnName);

  $: summary = derived(
    useRuntimeServiceGetDescriptiveStatistics(
      $runtimeStore?.instanceId,
      objectName,
      { columnName: columnName }
    ),
    ($query) => {
      return $query?.data?.numericSummary?.numericStatistics;
    }
  );
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
  {compact}
  {mode}
  {example}
  {type}
>
  <DataTypeIcon {type} slot="icon" />
  <svelte:fragment slot="left">{columnName}</svelte:fragment>
  <NumericSpark slot="summary" data={$numericHistogram} {compact} />
  <NullPercentageSpark
    slot="nullity"
    nullCount={$nulls?.nullCount}
    totalRows={$nulls?.totalRows}
    {type}
  />
  <div slot="details" class="pl-10 pr-4 py-4">
    <NumericPlot
      data={$numericHistogram}
      rug={$rug}
      summary={$summary}
      topK={$topK}
      totalRows={$nulls?.totalRows}
    />
  </div>
</ProfileContainer>
