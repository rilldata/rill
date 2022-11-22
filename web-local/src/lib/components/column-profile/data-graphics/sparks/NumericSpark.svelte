<script lang="ts">
  import { runtimeServiceGetNumericHistogram } from "@rilldata/web-common/runtime-client";
  import { COLUMN_PROFILE_CONFIG } from "@rilldata/web-local/lib/application-config";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { DATA_TYPE_COLORS } from "@rilldata/web-local/lib/duckdb-data-types";
  import Tooltip from "../../../tooltip/Tooltip.svelte";
  import TooltipContent from "../../../tooltip/TooltipContent.svelte";
  import Histogram from "../../../viz/histogram/SmallHistogram.svelte";

  // FIXME: figure out how to remove this requirement!
  export let objectName: string;
  export let columnName: string;
  export let containerWidth: number = 300;
  export let compactBreakpoint = 300;

  $: summaryWidthSize =
    COLUMN_PROFILE_CONFIG.summaryVizWidth[
      containerWidth < compactBreakpoint ? "small" : "medium"
    ];

  let histogram;
  if ($runtimeStore?.instanceId) {
    runtimeServiceGetNumericHistogram(
      $runtimeStore?.instanceId,
      objectName,
      columnName
    ).then((results) => (histogram = results.numericHistogramBins.bins));
  }
  $: console.log(histogram);
</script>

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
