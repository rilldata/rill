<script lang="ts">
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/shift-click-action";
  import { INTERVALS } from "@rilldata/web-common/lib/duckdb-data-types";
  import {
    useRuntimeServiceGetDescriptiveStatistics,
    useRuntimeServiceGetRugHistogram,
  } from "@rilldata/web-common/runtime-client";
  import { getPriorityForColumn } from "@rilldata/web-common/runtime-client/http-request-queue/priorities";
  import { derived } from "svelte/store";
  import { getHttpRequestQueueForHost } from "../../../runtime-client/http-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import ColumnProfileIcon from "../ColumnProfileIcon.svelte";
  import ProfileContainer from "../ProfileContainer.svelte";
  import {
    getNullPercentage,
    getNumericHistogram,
    getTopK,
    isFetching,
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

  $: nulls = getNullPercentage($runtime?.instanceId, objectName, columnName);

  $: numericHistogram = getNumericHistogram(
    $runtime?.instanceId,
    objectName,
    columnName,
    active
  );
  $: rug = useRuntimeServiceGetRugHistogram(
    $runtime?.instanceId,
    objectName,
    { columnName, priority: getPriorityForColumn("rug-histogram", active) },
    {
      query: {
        select($query) {
          return $query?.numericSummary?.numericOutliers?.outliers;
        },
      },
    }
  );
  $: topK = getTopK($runtime?.instanceId, objectName, columnName);

  $: summary = derived(
    useRuntimeServiceGetDescriptiveStatistics(
      $runtime?.instanceId,
      objectName,
      {
        columnName: columnName,
        priority: getPriorityForColumn("descriptive-statistics", active),
      }
    ),
    ($query) => {
      return $query?.data?.numericSummary?.numericStatistics;
    }
  );

  function toggleColumnProfile() {
    active = !active;
    const httpRequestQueue = getHttpRequestQueueForHost($runtime.host);
    httpRequestQueue.prioritiseColumn(objectName, columnName, active);
  }

  $: fetchingSummaries = isFetching($nulls, $numericHistogram);
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
  <NumericSpark {compact} data={$numericHistogram?.data} slot="summary" />
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
    class:hidden={INTERVALS.has(type)}
  >
    <NumericPlot
      data={$numericHistogram.data}
      rug={$rug?.data}
      summary={$summary}
      topK={$topK}
      totalRows={$nulls?.totalRows}
      {type}
    />
  </div>
</ProfileContainer>
