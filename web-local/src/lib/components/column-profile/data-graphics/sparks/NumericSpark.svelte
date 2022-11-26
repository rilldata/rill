<script lang="ts">
  import { useRuntimeServiceGetNumericHistogram } from "@rilldata/web-common/runtime-client";
  import { COLUMN_PROFILE_CONFIG } from "@rilldata/web-local/lib/application-config";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { DATA_TYPE_COLORS } from "@rilldata/web-local/lib/duckdb-data-types";
  import Tooltip from "../../../tooltip/Tooltip.svelte";
  import TooltipContent from "../../../tooltip/TooltipContent.svelte";
  import Histogram from "../../../viz/histogram/SmallHistogram.svelte";

  // FIXME: figure out how to remove this requirement!
  export let objectName: string;
  export let columnName: string;
  export let compact = false;

  $: summaryWidthSize =
    COLUMN_PROFILE_CONFIG.summaryVizWidth[compact ? "small" : "medium"];

  let histogram;
  $: histogramQuery = useRuntimeServiceGetNumericHistogram(
    $runtimeStore?.instanceId,
    objectName,
    columnName
  );
  $: histogram =
    $histogramQuery?.data?.numericSummary?.numericHistogramBins?.bins;
</script>

<div>
  {#if histogram}
    <Tooltip location="right" alignment="center" distance={8}>
      <Histogram
        data={histogram}
        width={summaryWidthSize}
        height={18}
        fillColor={DATA_TYPE_COLORS["DOUBLE"].vizFillClass}
        baselineStrokeColor={DATA_TYPE_COLORS["DOUBLE"].vizStrokeClass}
      />
      <TooltipContent slot="tooltip-content">
        the distribution of the values of this column
      </TooltipContent>
    </Tooltip>
  {/if}
</div>
