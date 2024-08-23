<script lang="ts">
  import VegaLiteRenderer from "@rilldata/web-common/features/canvas-components/render/VegaLiteRenderer.svelte";
  import VegaRenderer from "@rilldata/web-common/features/charts/render/VegaRenderer.svelte";
  import {
    resolveSignalField,
    resolveSignalTimeField,
    resolveSignalIntervalField,
  } from "@rilldata/web-common/features/charts/render/vega-signals";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { tableInteractionStore } from "@rilldata/web-common/features/dashboards/time-dimension-details/time-dimension-data-store";
  import { DimensionDataItem } from "@rilldata/web-common/features/dashboards/time-series/multiple-dimension-queries";
  import { TimeSeriesDatum } from "@rilldata/web-common/features/dashboards/time-series/timeseries-data-store";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import {
    MetricsViewSpecMeasureV2,
    V1TimeGrain,
  } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import { VegaSpec, View } from "svelte-vega";
  import { compile } from "vega-lite";
  import { TDDAlternateCharts } from "../types";
  import { patchSpecForTDD } from "./patch-vega-spec";
  import { tddTooltipFormatter } from "./tdd-tooltip-formatter";
  import {
    getVegaLiteSpecForTDD,
    hasBrushParam,
    reduceDimensionData,
    updateVegaOnTableHover,
  } from "./utils";
  import { VEGA_BAR_WIDTH_SIZE } from "../../time-series/chart-colors";

  export let totalsData: TimeSeriesDatum[];
  export let dimensionData: DimensionDataItem[];
  export let expandedMeasureName: string;
  export let chartType: TDDAlternateCharts;
  export let xMin: Date;
  export let xMax: Date;
  export let timeGrain: V1TimeGrain | undefined;
  export let isTimeComparison: boolean;

  let viewVL: View;
  let vegaSpec: VegaSpec;

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

  $: hoveredTime = $tableInteractionStore.time;
  $: hoveredDimensionValue = $tableInteractionStore.dimensionValue;

  $: {
    updateVegaOnTableHover(
      viewVL,
      chartType,
      isTimeComparison,
      hasDimensionData,
      hoveredTime,
      hoveredDimensionValue,
    );
  }

  $: specForTDD = getVegaLiteSpecForTDD(
    chartType,
    expandedMeasureName,
    expandedMeasureLabel,
    isTimeComparison,
    hasDimensionData,
    comparedDimensionLabel,
    selectedValues,
  );

  $: sanitizedVegaLiteSpec = patchSpecForTDD(
    specForTDD,
    chartType,
    timeGrain || V1TimeGrain.TIME_GRAIN_DAY,
    xMin,
    xMax,
    isTimeComparison,
    expandedMeasureName,
    selectedValues,
  );

  $: {
    if (hasBrushParam(sanitizedVegaLiteSpec)) {
      // Compile vega lite spec to vega spec
      // See: https://github.com/vega/vega-lite/issues/5341
      const compiledSpec = compile(sanitizedVegaLiteSpec).spec;

      // Add vega signal
      // See: https://github.com/vega/vega-lite/issues/3338
      // See: https://vega.github.io/vega/docs/signals/
      vegaSpec = {
        ...compiledSpec,
        signals: [
          ...(compiledSpec.signals || []),
          {
            name: "brush_end",
            on: [
              {
                events: {
                  source: "scope",
                  type: "pointerup",
                },
                update: { signal: "brush" },
              },
              {
                events: {
                  source: "scope",
                  type: "pointerdown",
                },
                update: { signal: "brush" },
              },
              // Track global pointer events (when user pointerups outside of chart)
              {
                events: {
                  source: "window",
                  type: "pointerup",
                },
                update: { signal: "brush" },
              },
              {
                events: {
                  source: "window",
                  type: "pointerdown",
                },
                update: { signal: "brush" },
              },
            ],
          },
        ],
      };
    }
  }

  // $: console.log("vegaSpec: ", vegaSpec.signals);

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
    brush_end: (_name: string, value: boolean) => {
      const interval = resolveSignalIntervalField(value);
      dispatch("chart-brush-end", { interval, isScrubbing: false });
    },
    brush: (_name: string, value) => {
      const interval = resolveSignalIntervalField(value);
      dispatch("chart-brush", { interval, isScrubbing: true });
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

  // onMount(() => {
  //   window.addEventListener("brushCleared", () => {
  //     dispatch("chart-brush", { interval: null, isScrubbing: false });
  //   });

  //   return () => {
  //     window.removeEventListener("brushCleared", () => {
  //       dispatch("chart-brush", { interval: null, isScrubbing: false });
  //     });
  //   };
  // });
</script>

{#if hasBrushParam(sanitizedVegaLiteSpec) && data}
  <VegaRenderer
    bind:viewVL
    data={{ table: data }}
    spec={vegaSpec}
    {signalListeners}
    {expressionFunctions}
    {tooltipFormatter}
  />
{:else}
  <VegaLiteRenderer
    bind:viewVL
    data={{ table: data }}
    spec={sanitizedVegaLiteSpec}
    {signalListeners}
    {expressionFunctions}
    {tooltipFormatter}
  />
{/if}
