<script lang="ts">
  import { sanitizeFieldName } from "@rilldata/web-common/components/vega/util";
  import { getRillTheme } from "@rilldata/web-common/components/vega/vega-config";
  import VegaLiteRenderer from "@rilldata/web-common/components/vega/VegaLiteRenderer.svelte";
  import type { CanvasChartSpec } from "@rilldata/web-common/features/canvas/components/charts";
  import ComponentError from "@rilldata/web-common/features/components/ComponentError.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import {
    createMeasureValueFormatter,
    humanizeDataType,
  } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
  import type { Readable } from "svelte/store";
  import type { View } from "vega-typings";
  import type { ChartDataResult, ChartType } from "./types";
  import { generateSpec, getColorMappingForChart } from "./util";
  import { getChroma } from "../../themes/theme-utils";

  export let chartType: ChartType;
  export let chartSpec: CanvasChartSpec;
  export let chartData: Readable<ChartDataResult>;
  export let measures: MetricsViewSpecMeasure[];
  export let themeMode: "light" | "dark" = "light";
  /**
   * Full theme object with all CSS variables (primary, secondary, background, etc.)
   * If provided, chart uses these directly. If not, falls back to defaults.
   */
  export let theme: Record<string, string> | undefined = undefined;
  export let isCanvas: boolean;

  let viewVL: View;

  $: ({ data, domainValues, isFetching, error } = $chartData);

  $: hasNoData = !isFetching && data.length === 0;

  // Override chartData theme with mode-aware colors if theme prop is provided
  $: chartDataWithTheme = theme
    ? {
        ...$chartData,
        theme: {
          primary: theme.primary
            ? getChroma(theme.primary)
            : $chartData.theme.primary,
          secondary: theme.secondary
            ? getChroma(theme.secondary)
            : $chartData.theme.secondary,
        },
      }
    : $chartData;

  $: spec = generateSpec(chartType, chartSpec, chartDataWithTheme);

  // TODO: Move this to a central cached store
  $: measureFormatters = measures.reduce(
    (acc, measure) => ({
      ...acc,
      [sanitizeFieldName(measure.name || "measure")]:
        createMeasureValueFormatter<null | undefined>(measure),
    }),
    {} as Record<string, (value: number | null | undefined) => string>,
  );

  $: expressionFunctions = {
    humanize: {
      fn: (val: number) =>
        humanizeDataType(val, FormatPreset.HUMANIZE, "table"),
    },
    ...measures.reduce((acc, measure) => {
      const fieldName = sanitizeFieldName(measure.name || "measure");
      const formatter = measureFormatters[fieldName];
      return {
        ...acc,
        [fieldName]: {
          fn: (val: number) => (formatter ? formatter(val) : String(val)),
        },
      };
    }, {}),
  };

  // Color mapping needs to be reactive to theme mode changes (light/dark)
  // because colors are resolved differently for each mode
  $: isThemeModeDark = themeMode === "dark";
  $: colorMapping = getColorMappingForChart(
    chartSpec,
    domainValues,
    isThemeModeDark,
  );
</script>

{#if isFetching || measures.length === 0}
  <div class="flex items-center justify-center h-full w-full">
    <Spinner status={EntityStatus.Running} size="20px" />
  </div>
{:else if error}
  <ComponentError error={error.message} />
{:else if hasNoData}
  <div
    class="flex w-full h-full p-2 text-xl ui-copy-disabled items-center justify-center"
  >
    No Data to Display
  </div>
{:else}
  <VegaLiteRenderer
    bind:viewVL
    canvasDashboard={isCanvas}
    data={{ "metrics-view": data }}
    {themeMode}
    {spec}
    {colorMapping}
    renderer="canvas"
    {expressionFunctions}
    config={getRillTheme(true, isThemeModeDark, theme)}
  />
{/if}
