<script lang="ts">
  import VegaLiteRenderer from "@rilldata/web-common/features/charts/render/VegaLiteRenderer.svelte";
  import {
    resolveSignalField,
    resolveSignalTimeField,
  } from "@rilldata/web-common/features/charts/render/vega-signals";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { DimensionDataItem } from "@rilldata/web-common/features/dashboards/time-series/multiple-dimension-queries";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import {
    MetricsViewSpecMeasureV2,
    V1TimeGrain,
  } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import { TDDAlternateCharts } from "../types";
  import { tddTooltipFormatter } from "./tdd-tooltip-formatter";
  import {
    getVegaSpecForTDD,
    patchSpecForTDD,
    reduceDimensionData,
  } from "./utils";

  export let totalsData;
  export let dimensionData: DimensionDataItem[];
  export let expandedMeasureName: string;
  export let chartType: TDDAlternateCharts;
  export let xMin: Date;
  export let xMax: Date;
  export let timeGrain: V1TimeGrain | undefined;
  export let isTimeComparison: boolean;

  const dispatch = createEventDispatcher();
  const {
    selectors: {
      measures: { measureLabel, getMeasureByName },
      dimensions: { comparisonDimension },
    },
  } = getStateManagers();

  $: hasDimensionData = !!dimensionData?.length;
  $: data = hasDimensionData ? reduceDimensionData(dimensionData) : totalsData;
  $: selectedValues = hasDimensionData ? dimensionData.map((d) => d.value) : [];
  $: expandedMeasureLabel = $measureLabel(expandedMeasureName);
  $: measure = $getMeasureByName(expandedMeasureName);
  $: comparedDimensionLabel =
    $comparisonDimension?.label || $comparisonDimension?.name;

  $: vegaSpec = getVegaSpecForTDD(
    chartType,
    expandedMeasureName,
    expandedMeasureLabel,
    isTimeComparison,
    hasDimensionData,
    comparedDimensionLabel,
    selectedValues,
  );

  $: sanitizedVegaSpec = patchSpecForTDD(
    vegaSpec,
    chartType,
    timeGrain || V1TimeGrain.TIME_GRAIN_DAY,
    xMin,
    xMax,
    isTimeComparison,
    expandedMeasureName,
    selectedValues,
  );

  $: tooltipFormatter = tddTooltipFormatter(
    chartType,
    expandedMeasureLabel,
    comparedDimensionLabel,
    isTimeComparison,
    selectedValues,
    timeGrain,
  );

  const signalListeners = {
    hover: (_name: string, value) => {
      const dimension = resolveSignalField(value, "dimension");
      const ts = resolveSignalTimeField(value);

      dispatch("chart-hover", { dimension, ts });
    },
  };

  $: measureFormatter = createMeasureValueFormatter<null | undefined>(
    measure as MetricsViewSpecMeasureV2,
  );

  function vegaCustomFormatter(val) {
    return measureFormatter(val);
  }

  const expressionFunctions = {
    measureFormatter: { fn: vegaCustomFormatter },
  };
</script>

{#if sanitizedVegaSpec && data}
  <VegaLiteRenderer
    data={{ table: data }}
    spec={sanitizedVegaSpec}
    {signalListeners}
    {expressionFunctions}
    {tooltipFormatter}
  />
{/if}
