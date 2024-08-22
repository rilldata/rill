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
  import { createEventDispatcher, onMount } from "svelte";
  import { View } from "svelte-vega";
  import { compile } from "vega-lite";
  import { TDDAlternateCharts } from "../types";
  import { patchSpecForTDD } from "./patch-vega-spec";
  import { tddTooltipFormatter } from "./tdd-tooltip-formatter";
  import {
    getVegaSpecForTDD,
    reduceDimensionData,
    updateVegaOnTableHover,
  } from "./utils";

  export let totalsData: TimeSeriesDatum[];
  export let dimensionData: DimensionDataItem[];
  export let expandedMeasureName: string;
  export let chartType: TDDAlternateCharts;
  export let xMin: Date;
  export let xMax: Date;
  export let timeGrain: V1TimeGrain | undefined;
  export let isTimeComparison: boolean;

  let viewVL: View;
  let vegaSpec: any;

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

  // TODO: rename
  // vega lite spec
  $: vegaSpecForTDD = getVegaSpecForTDD(
    chartType,
    expandedMeasureName,
    expandedMeasureLabel,
    isTimeComparison,
    hasDimensionData,
    comparedDimensionLabel,
    selectedValues,
  );

  // TODO: rename
  // vega lite spec
  $: sanitizedVegaSpec = patchSpecForTDD(
    vegaSpecForTDD,
    chartType,
    timeGrain || V1TimeGrain.TIME_GRAIN_DAY,
    xMin,
    xMax,
    isTimeComparison,
    expandedMeasureName,
    selectedValues,
  );

  // TODO: check if sanitized vega spec already has brush params

  $: {
    if (sanitizedVegaSpec) {
      // Compile vega lite spec to vega spec
      // See: https://github.com/vega/vega-lite/issues/5341
      // See: https://github.com/vega/vega-lite/issues/3338
      const compiledSpec = compile(sanitizedVegaSpec).spec;

      // Add custom signals for brushstart and brushend
      // See: https://vega.github.io/vega/docs/signals/
      // See: https://vega.github.io/vega-lite/docs/parameter.html#using-parameters
      vegaSpec = {
        ...compiledSpec,
        // TODO: check if sanitized vega spec already has brush params
        signals: [
          ...(compiledSpec.signals || []),
          {
            name: "brush_start",
            value: {},
            on: [{ events: "brush:start", update: "{time: x()}" }],
          },
          {
            name: "brush_end",
            value: {},
            on: [{ events: "brush:end", update: "{time: x()}" }],
          },
        ],
      };
    }
  }

  $: console.log("vegaSpec: ", vegaSpec);

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
    // Debouncing is a short term solution to prevent the scrubbing from firing
    // on every pixel dragged. The ideal solution is to listen to the completion
    // of the drag and then fire the brush event.
    // See: https://github.com/vega/vega-lite/issues/5341
    // brush: (_name: string, value) => {
    //   const interval = resolveSignalIntervalField(value);

    //   dispatch("chart-brush", { interval });
    // },
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

  onMount(() => {
    window.addEventListener("brushCleared", () => {
      dispatch("chart-brush", { interval: null, isScrubbing: false });
    });

    return () => {
      window.removeEventListener("brushCleared", () => {
        dispatch("chart-brush", { interval: null, isScrubbing: false });
      });
    };
  });
</script>

{#if vegaSpec && data}
  <VegaRenderer
    bind:viewVL
    data={{ table: data }}
    spec={vegaSpec}
    {signalListeners}
    {expressionFunctions}
  />
{:else if sanitizedVegaSpec && data}
  <VegaLiteRenderer
    bind:viewVL
    data={{ table: data }}
    spec={sanitizedVegaSpec}
    {signalListeners}
    {expressionFunctions}
    {tooltipFormatter}
  />
{/if}
