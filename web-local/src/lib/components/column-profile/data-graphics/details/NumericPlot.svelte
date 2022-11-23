<script lang="ts">
  import {
    useRuntimeServiceGetDescriptiveStatistics,
    useRuntimeServiceGetNumericHistogram,
    useRuntimeServiceGetRugHistogram,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { WithParentClientRect } from "../../../data-graphic/functional-components";
  import NumericHistogram from "../../../viz/histogram/NumericHistogram.svelte";
  import OutlierHistogram from "../../../viz/histogram/OutlierHistogram.svelte";

  export let objectName: string;
  export let columnName: string;

  // DELETE THESE
  let containerWidth = 200;
  let indentLevel = 1;

  $: summaryStatisticsQuery = useRuntimeServiceGetDescriptiveStatistics(
    $runtimeStore?.instanceId,
    objectName,
    columnName
  );
  $: summary = $summaryStatisticsQuery?.data?.numericSummary?.numericStatistics;

  $: histogramQuery = useRuntimeServiceGetNumericHistogram(
    $runtimeStore?.instanceId,
    objectName,
    columnName
  );
  $: histogram =
    $histogramQuery?.data?.numericSummary?.numericHistogramBins?.bins;

  $: outliersQuery = useRuntimeServiceGetRugHistogram(
    $runtimeStore?.instanceId,
    objectName,
    columnName
  );
  $: outliers = $outliersQuery?.data?.numericSummary?.numericOutliers?.outliers;
</script>

<WithParentClientRect let:rect>
  {#if histogram && summary}
    <NumericHistogram
      width={rect?.width || 400}
      height={65}
      data={histogram}
      min={summary?.min}
      qlow={summary?.q25}
      median={summary?.q50}
      qhigh={summary?.q75}
      mean={summary?.mean}
      max={summary?.max}
    />
  {/if}
  {#if outliers && outliers?.length}
    <OutlierHistogram
      width={(rect?.width || 400) - (indentLevel === 1 ? 20 + 24 + 44 : 32)}
      height={15}
      data={outliers}
      mean={summary?.mean}
      sd={summary?.sd}
      min={summary?.min}
      max={summary?.max}
    />
  {/if}
</WithParentClientRect>
