<script lang="ts">
  import { GraphicContext } from "@rilldata/web-common/components/data-graphic/elements";
  import {
    NicelyFormattedTypes,
    humanizeDataType,
    nicelyFormattedTypesToNumberKind,
  } from "@rilldata/web-common/features/dashboards/humanize-numbers";
  import MeasureChart from "@rilldata/web-common/features/dashboards/time-series/MeasureChart.svelte";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";

  export let start;
  export let end;
  export let label: string;
  export let value: number;
  export let format: string;
  export let trend;

  $: formattedNumber = humanizeDataType(value, format as NicelyFormattedTypes);

  export let mouseoverValue;
</script>

<div class="grid gap-x-2 w-full" style:grid-template-columns="1fr 160px">
  <div class="pt-3">
    <div style:font-size="12px" class="flex font-regular truncate">
      {label}
    </div>
    <!-- <div style:font-size="14px" class="font-regular">
      {formattedNumber}
    </div> -->
  </div>
  <slot>
    {#if trend}
      <div>
        <GraphicContext
          xMin={start}
          xMax={end}
          xType="date"
          bottom={1}
          yType="number"
        >
          <MeasureChart
            width={160}
            height={54}
            data={trend}
            xAccessor="ts"
            yAccessor="value"
            showYAxis={false}
            xMin={trend[0].ts}
            xMax={trend[trend.length - 1].ts}
            numberKind={nicelyFormattedTypesToNumberKind(format)}
            mouseoverFormat={(d) => humanizeDataType(d, format)}
            showTimeMouseover={false}
            showGrid={false}
            timeGrain={V1TimeGrain.TIME_GRAIN_DAY}
            bind:mouseoverValue
          />
        </GraphicContext>
      </div>
    {/if}
  </slot>
</div>
