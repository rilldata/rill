<script lang="ts">
  import { sanitizeFieldName } from "@rilldata/web-common/components/vega/util";
  import { getRillTheme } from "@rilldata/web-common/components/vega/vega-config";
  import VegaLiteRenderer from "@rilldata/web-common/components/vega/VegaLiteRenderer.svelte";
  import type { ChartSpec } from "@rilldata/web-common/features/canvas/components/charts";
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

  export let chartType: ChartType;
  export let chartSpec: ChartSpec;
  export let chartData: Readable<ChartDataResult>;
  export let measures: MetricsViewSpecMeasure[];
  export let theme: "light" | "dark" = "light";
  export let isCanvas: boolean;

  let viewVL: View;

  $: ({ data, domainValues, isFetching, error } = $chartData);

  $: hasNoData = !isFetching && data.length === 0;

  $: spec = generateSpec(chartType, chartSpec, $chartData);

  // TODO: Move this to a central cached store
  $: measureFormatters = measures.reduce(
    (acc, measure) => ({
      ...acc,
      [sanitizeFieldName(measure.name || "measure")]:
        createMeasureValueFormatter<null | undefined>(measure),
    }),
    {},
  );

  $: expressionFunctions = {
    humanize: {
      fn: (val) => humanizeDataType(val, FormatPreset.HUMANIZE, "table"),
    },
    ...measures.reduce(
      (acc, measure) => {
        const fieldName = sanitizeFieldName(measure.name || "measure");
        return {
          ...acc,
          [fieldName]: { fn: (val) => measureFormatters[fieldName](val) },
        };
      },
      {} as Record<string, { fn: (val: any) => string }>,
    ),
  };

  $: colorMapping = getColorMappingForChart(chartSpec, domainValues);
</script>

{#if isFetching}
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
    {theme}
    {spec}
    {colorMapping}
    renderer="canvas"
    {expressionFunctions}
    config={getRillTheme(true, theme === "dark")}
  />
{/if}
