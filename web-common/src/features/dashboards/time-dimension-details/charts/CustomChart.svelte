<script lang="ts">
  import VegaLiteRenderer from "@rilldata/web-common/features/charts/render/VegaLiteRenderer.svelte";
  import { DimensionDataItem } from "@rilldata/web-common/features/dashboards/time-series/multiple-dimension-queries";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { TDDCustomCharts } from "../types";
  import {
    getVegaSpec,
    reduceDimensionData,
    sanitizeSpecForTDD,
  } from "./utils";

  export let totalsData;
  export let dimensionData: DimensionDataItem[];
  export let expandedMeasureName: string;
  export let chartType: TDDCustomCharts;
  export let xMin: Date;
  export let xMax: Date;
  export let timeGrain: V1TimeGrain | undefined;

  // Reactive statements
  $: hasDimensionData = !!dimensionData?.length;
  $: data = hasDimensionData ? reduceDimensionData(dimensionData) : totalsData;
  $: vegaSpec = getVegaSpec(chartType, expandedMeasureName, hasDimensionData);
  $: sanitizedVegaSpec = sanitizeSpecForTDD(
    vegaSpec,
    timeGrain || V1TimeGrain.TIME_GRAIN_DAY,
    xMin,
    xMax,
    chartType,
  );

  $: console.log(sanitizedVegaSpec);
</script>

{#if sanitizedVegaSpec && data}
  <VegaLiteRenderer data={{ table: data }} spec={sanitizedVegaSpec} />
{/if}
