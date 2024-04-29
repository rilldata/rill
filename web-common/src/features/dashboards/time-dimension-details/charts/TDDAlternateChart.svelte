<script lang="ts">
  import VegaLiteRenderer from "@rilldata/web-common/features/charts/render/VegaLiteRenderer.svelte";
  import {
    resolveSignalField,
    resolveSignalTimeField,
  } from "@rilldata/web-common/features/charts/render/vega-signals";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { DimensionDataItem } from "@rilldata/web-common/features/dashboards/time-series/multiple-dimension-queries";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import { TDDAlternateCharts } from "../types";
  import { tddTooltipFormatter } from "./tdd-tooltip-formatter";
  import {
    getVegaSpecForTDD,
    reduceDimensionData,
    sanitizeSpecForTDD,
  } from "./utils";

  export let totalsData;
  export let dimensionData: DimensionDataItem[];
  export let expandedMeasureName: string;
  export let chartType: TDDAlternateCharts;
  export let xMin: Date;
  export let xMax: Date;
  export let timeGrain: V1TimeGrain | undefined;

  const dispatch = createEventDispatcher();
  const {
    selectors: {
      measures: { measureLabel },
      dimensions: { comparisonDimension },
    },
  } = getStateManagers();

  const signalListeners = {
    hover: (_name: string, value) => {
      const dimension = resolveSignalField(value, "dimension");
      const ts = resolveSignalTimeField(value);

      dispatch("chart-hover", { dimension, ts });
    },
  };

  $: hasDimensionData = !!dimensionData?.length;
  $: data = hasDimensionData ? reduceDimensionData(dimensionData) : totalsData;
  $: selectedValues = hasDimensionData ? dimensionData.map((d) => d.value) : [];
  $: expandedMeasureLabel = $measureLabel(expandedMeasureName);
  $: comparedDimensionLabel =
    $comparisonDimension?.label || $comparisonDimension?.name;

  $: vegaSpec = getVegaSpecForTDD(
    chartType,
    expandedMeasureName,
    expandedMeasureLabel,
    hasDimensionData,
    comparedDimensionLabel,
    selectedValues,
  );

  $: sanitizedVegaSpec = sanitizeSpecForTDD(
    vegaSpec,
    timeGrain || V1TimeGrain.TIME_GRAIN_DAY,
    xMin,
    xMax,
    selectedValues,
  );

  $: tooltipFormatter = tddTooltipFormatter(
    chartType,
    expandedMeasureLabel,
    comparedDimensionLabel,
    selectedValues,
    timeGrain,
  );
</script>

{#if sanitizedVegaSpec && data}
  <VegaLiteRenderer
    data={{ table: data }}
    spec={sanitizedVegaSpec}
    {signalListeners}
    {tooltipFormatter}
  />
{/if}
