<script lang="ts">
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/shift-click-action";
  import {
    FLOATS,
    INTERVALS,
    isFloat,
  } from "@rilldata/web-common/lib/duckdb-data-types";
  import {
    QueryServiceColumnNumericHistogramHistogramMethod,
    useQueryServiceColumnDescriptiveStatistics,
    useQueryServiceColumnRugHistogram,
  } from "@rilldata/web-common/runtime-client";
  import { getPriorityForColumn } from "@rilldata/web-common/runtime-client/http-request-queue/priorities";
  import { derived } from "svelte/store";
  import { httpRequestQueue } from "../../../runtime-client/http-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import ColumnProfileIcon from "../ColumnProfileIcon.svelte";
  import ProfileContainer from "../ProfileContainer.svelte";
  import {
    getNullPercentage,
    getNumericHistogram,
    getTopK,
    isFetching,
  } from "../queries";
  import { chooseBetweenDiagnosticAndStatistical } from "../utils";
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

  $: diagnosticHistogram = getNumericHistogram(
    $runtime?.instanceId,
    objectName,
    columnName,
    QueryServiceColumnNumericHistogramHistogramMethod.HISTOGRAM_METHOD_DIAGNOSTIC,
    active
  );
  let fdHistogram;
  $: if (isFloat(type)) {
    fdHistogram = getNumericHistogram(
      $runtime?.instanceId,
      objectName,
      columnName,
      QueryServiceColumnNumericHistogramHistogramMethod.HISTOGRAM_METHOD_FD,
      active
    );
  }

  /**
   * We have two choices of histogram method: diagnostic and freedman-diaconis.
   * For integers, we go with diagnostic. For floating points, let's choose between
   * the most viable of diagnostic and freedman-diaconis. We'll remove
   * this once we've refactored floating-point columns toward a KDE plot.
   */
  $: histogramData = isFloat(type)
    ? chooseBetweenDiagnosticAndStatistical(
        $diagnosticHistogram?.data,
        $fdHistogram?.data
      )
    : $diagnosticHistogram?.data;

  $: rug = useQueryServiceColumnRugHistogram(
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
    useQueryServiceColumnDescriptiveStatistics(
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
    httpRequestQueue.prioritiseColumn(objectName, columnName, active);
  }

  $: fetchingSummaries = FLOATS.has(type)
    ? isFetching($nulls, $diagnosticHistogram, $fdHistogram)
    : isFetching($nulls, $diagnosticHistogram);

  /** if we have a singleton where all summary information is the same, let's construct a single bin. */
  $: if (
    $summary?.min !== undefined &&
    $summary?.min === $summary?.max &&
    $nulls?.totalRows !== undefined
  ) {
    const boundaries = 10;
    histogramData = [
      // add 4 more empty bins
      ...Array.from({ length: boundaries }).map((_, i) => {
        return {
          bucket: -boundaries + i,
          count: 0,
          high: $summary?.min - (boundaries - i - 1),
          low: $summary?.min - (boundaries - i),
        };
      }),
      {
        bucket: boundaries,
        count: $nulls?.totalRows,
        low: $summary?.min,
        high: $summary?.min + 1,
      },
      // add more empty bins
      ...Array.from({ length: boundaries }).map((_, i) => {
        return {
          bucket: boundaries + i + 1,
          count: 0,
          low: $summary?.min + i,
          high: $summary?.min + i + 1,
        };
      }),
    ];
  }
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
  <NumericSpark {type} {compact} data={histogramData} slot="summary" />
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
      data={histogramData}
      rug={$rug?.data}
      summary={$summary}
      topK={$topK}
      totalRows={$nulls?.totalRows}
      {type}
    />
  </div>
</ProfileContainer>
