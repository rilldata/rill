<script lang="ts">
  import VegaLiteRenderer from "@rilldata/web-common/features/canvas-components/render/VegaLiteRenderer.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { tableInteractionStore } from "@rilldata/web-common/features/dashboards/time-dimension-details/time-dimension-data-store";
  import type { DimensionDataItem } from "@rilldata/web-common/features/dashboards/time-series/multiple-dimension-queries";
  import type { TimeSeriesDatum } from "@rilldata/web-common/features/dashboards/time-series/timeseries-data-store";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import {
    type MetricsViewSpecMeasureV2,
    V1TimeGrain,
  } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher, onDestroy } from "svelte";
  import type { VegaSpec, View } from "svelte-vega";
  import type { TopLevelSpec } from "vega-lite";
  import type { TDDAlternateCharts } from "../types";
  import { patchSpecForTDD } from "./patch-vega-spec";
  import { tddTooltipFormatter } from "./tdd-tooltip-formatter";
  import {
    getVegaLiteSpecForTDD,
    hasBrushParam,
    reduceDimensionData,
    updateChartOnTableCellHover,
  } from "./utils";
  import { VegaSignalManager } from "./vega-signal-manager";
  import VegaRenderer from "@rilldata/web-common/features/canvas-components/render/VegaRenderer.svelte";
  import {
    resolveSignalField,
    resolveSignalTimeField,
    resolveSignalIntervalField,
  } from "@rilldata/web-common/features/canvas-components/render/vega-signals";

  export let totalsData: TimeSeriesDatum[];
  export let dimensionData: DimensionDataItem[];
  export let expandedMeasureName: string;
  export let chartType: TDDAlternateCharts;
  export let xMin: Date;
  export let xMax: Date;
  export let timeGrain: V1TimeGrain | undefined;
  export let isTimeComparison: boolean;
  export let isScrubbing: boolean;

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
    $comparisonDimension?.displayName || $comparisonDimension?.name;

  $: hoveredTime = $tableInteractionStore.time;
  $: hoveredDimensionValue = $tableInteractionStore.dimensionValue;

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
    updateChartOnTableCellHover(
      viewVL,
      chartType,
      isTimeComparison,
      hasDimensionData,
      hoveredTime,
      hoveredDimensionValue,
    );
  }

  /**
   *
   * Compile vega lite spec to vega spec
   * See: https://github.com/vega/vega-lite/issues/5341
   *
   * Add brush signals to vega spec
   * See: https://github.com/vega/vega-lite/issues/3338
   * See: https://vega.github.io/vega/docs/signals/
   * Related: https://github.com/vega/vega-lite/issues/1830
   */
  $: {
    if (hasBrushParam(sanitizedVegaLiteSpec)) {
      const signalManager = new VegaSignalManager(
        sanitizedVegaLiteSpec as TopLevelSpec,
      );
      vegaSpec = signalManager.updateVegaSpec();
    }
  }

  $: tooltipFormatter = tddTooltipFormatter(
    chartType,
    expandedMeasureLabel,
    comparedDimensionLabel,
    isTimeComparison,
    selectedValues,
    timeGrain,
  );

  function updateAdaptiveScrubRange(interval) {
    let rafId: number | null = null;
    let lastUpdateTime = 0;
    let currentInterval = 1000 / 60; // Start with 60fps

    const MIN_INTERVAL = 1000 / 120; // Max 120fps
    const MAX_INTERVAL = 1000 / 30; // Min 30fps
    const ADJUSTMENT_FACTOR = 1.2; // Adjust interval based on performance

    if (rafId) {
      cancelAnimationFrame(rafId);
    }

    rafId = requestAnimationFrame((timestamp) => {
      const elapsed = timestamp - lastUpdateTime;
      if (elapsed >= currentInterval) {
        dispatch("chart-brush", { interval });
        lastUpdateTime = timestamp;

        // Adjust interval based on performance
        if (elapsed > currentInterval * ADJUSTMENT_FACTOR) {
          currentInterval = Math.min(
            currentInterval * ADJUSTMENT_FACTOR,
            MAX_INTERVAL,
          );
        } else {
          currentInterval = Math.max(
            currentInterval / ADJUSTMENT_FACTOR,
            MIN_INTERVAL,
          );
        }
      }
      rafId = null;
    });

    onDestroy(() => {
      if (rafId) {
        cancelAnimationFrame(rafId);
      }
    });

    return updateAdaptiveScrubRange;
  }

  const signalListeners = {
    hover: (_name: string, value) => {
      const dimension = resolveSignalField(value, "dimension");
      const ts = resolveSignalTimeField(value);

      dispatch("chart-hover", { dimension, ts });
    },
    brush: (_name: string, value) => {
      const interval = resolveSignalIntervalField(value);

      // Update view to prevent race condition
      viewVL.runAsync();

      updateAdaptiveScrubRange(interval);
    },
    brush_end: (_name: string, value: boolean) => {
      const interval = resolveSignalIntervalField(value);

      dispatch("chart-brush-end", { interval });
    },
    brush_clear: (_name: string, value: boolean) => {
      if (value) {
        dispatch("chart-brush-clear", {
          start: undefined,
          end: undefined,
        });
      }
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

{#if hasBrushParam(sanitizedVegaLiteSpec) && data}
  <VegaRenderer
    bind:view={viewVL}
    data={{ table: data }}
    spec={vegaSpec}
    {signalListeners}
    {expressionFunctions}
    {tooltipFormatter}
    {isScrubbing}
  />
{:else}
  <!-- JIC we add a new chart type without brush param -->
  <VegaLiteRenderer
    bind:viewVL
    data={{ table: data }}
    spec={sanitizedVegaLiteSpec}
    {signalListeners}
    {expressionFunctions}
    {tooltipFormatter}
  />
{/if}
