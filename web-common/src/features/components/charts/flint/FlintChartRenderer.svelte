<script lang="ts">
  import type { PartialMessage, Struct } from "@bufbuild/protobuf";
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import PreviewTable from "@rilldata/web-common/components/preview-table/PreviewTable.svelte";
  import { getRillTheme } from "@rilldata/web-common/components/vega/vega-config";
  import { sanitizeFieldName } from "@rilldata/web-common/components/vega/util";
  import VegaLiteRenderer from "@rilldata/web-common/components/vega/VegaLiteRenderer.svelte";
  import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
  import ComponentError from "@rilldata/web-common/features/components/ComponentError.svelte";
  import ReconcilingSpinner from "@rilldata/web-common/features/entity-management/ReconcilingSpinner.svelte";
  import { useMetricsViewValidSpec } from "@rilldata/web-common/features/dashboards/selectors";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import {
    createRuntimeServiceQueryResolver,
    V1TimeGrain,
    type V1Expression,
    type V1TimeRange,
  } from "@rilldata/web-common/runtime-client";
  import { V1TimeGrainToDateTimeUnit } from "@rilldata/web-common/lib/time/new-grains";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { View, VisualizationSpec } from "svelte-vega";
  import { convertV1ExpressionToMapstructure } from "../custom/expression-utils";
  import {
    compileFlintSpec,
    type FlintChartSpec,
  } from "@rilldata/web-common/features/custom-viz/flint/compile";
  import { EJECTED_DATA_NAME } from "@rilldata/web-common/features/custom-viz/flint/eject";
  import { deriveFlintFields } from "@rilldata/web-common/features/custom-viz/flint/semantic-types";
  import { debounce } from "@rilldata/web-common/lib/create-debouncer";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { onDestroy, tick } from "svelte";

  export let spec: FlintChartSpec | undefined = undefined;
  /**
   * A Vega-Lite spec to render instead of compiling `spec`, set by a component that has been
   * ejected. It renders through the same query, measure formatters and theme as a chart spec, so
   * ejecting changes how the chart is described rather than how it is drawn; what it gives up is
   * the compiler's data-driven layout, which is why the two are mutually exclusive.
   */
  export let vegaSpec: string | undefined = undefined;
  export let metricsSQL: string | undefined = undefined;
  export let metricsViewName: string | undefined = undefined;
  export let timeGrain: V1TimeGrain | undefined = undefined;
  export let renderer: "canvas" | "svg" = "svg";
  export let whereFilter: V1Expression | undefined = undefined;
  export let timeRange: V1TimeRange | undefined = undefined;
  export let showDataTable = false;
  export let name: string = "Custom Chart";
  export let themeMode: "light" | "dark" = "light";
  /**
   * Full theme object with all CSS variables (primary, secondary, background, etc.)
   * If provided, the chart uses these directly. If not, colors resolve from CSS variables.
   */
  export let theme: Record<string, string> | undefined = undefined;

  $: viewOptions = [m.canvas_chart(), m.component_preview_data()];

  let selectedView = 0; // 0 = Chart, 1 = Data
  let chartWidth = 0;
  let chartHeight = 0;
  let compileWidth = 0;
  let compileHeight = 0;
  let plotWidth = 0;
  let plotHeight = 0;
  let chartContainer: HTMLDivElement;
  let viewVL: View;
  let fitRun = 0;

  // Vega resizes the live view immediately. Flint only needs the settled size to recompute
  // layout choices, which avoids recreating the full spec on every drag-resize frame.
  const updateCompileSize = debounce((width: number, height: number) => {
    compileWidth = width;
    compileHeight = height;
  }, 150);
  $: updateCompileSize(chartWidth, chartHeight);
  onDestroy(() => {
    updateCompileSize.cancel();
    fitRun++;
  });

  const runtimeClient = useRuntimeClient();

  $: filterKey = JSON.stringify({ whereFilter, timeRange, timeGrain });

  // Only enable queries when the time range has resolved. When no timeRange is provided at all
  // (e.g. standalone component preview, outside a canvas), queries run unfiltered instead of waiting.
  $: hasValidTimeRange =
    timeRange === undefined || (!!timeRange?.start && !!timeRange?.end);

  $: dataQuery = createRuntimeServiceQueryResolver(
    runtimeClient,
    {
      resolver: "metrics_sql",
      resolverProperties: {
        sql: metricsSQL,
        ...(whereFilter?.cond?.exprs?.length
          ? { additional_where: convertV1ExpressionToMapstructure(whereFilter) }
          : {}),
        ...(timeRange ? { additional_time_range: timeRange } : {}),
        // The grain buckets the query's time dimensions, so the chart shows one mark per bucket of
        // the dashboard's time controls rather than one per raw timestamp. A component that writes
        // its own date_trunc keeps that bucket instead.
        ...(timeGrain && timeGrain !== V1TimeGrain.TIME_GRAIN_UNSPECIFIED
          ? { additional_time_grain: V1TimeGrainToDateTimeUnit[timeGrain] }
          : {}),
        // The time zone rides on the time range, which the resolver reads as a filter only; the
        // grain needs it as well, or the buckets are cut on UTC boundaries.
        ...(timeRange?.timeZone ? { time_zone: timeRange.timeZone } : {}),
      } as unknown as PartialMessage<Struct>,
    },
    {
      query: {
        enabled: !!metricsSQL && hasValidTimeRange,
        queryKey: [
          `metrics_sql`,
          runtimeClient.instanceId,
          name,
          metricsSQL,
          filterKey,
        ],
      },
    },
  );

  $: metricsViewQuery = useMetricsViewValidSpec(
    runtimeClient,
    metricsViewName ?? "",
  );
  $: metricsViewSpec = metricsViewName ? $metricsViewQuery.data : undefined;

  $: rows = ($dataQuery.data?.data ?? []) as Record<string, unknown>[];
  $: schemaFields = $dataQuery.data?.schema?.fields ?? [];
  $: columns = schemaFields.map((field) => field.name as string);

  $: tableColumns = schemaFields.map((field) => ({
    name: field.name,
    type: field.type?.code,
  })) as VirtualizedTableColumns[];

  $: fields = deriveFlintFields(columns, metricsViewSpec, timeGrain);

  // Measures present in the result, so their Rill formatters can be registered and referenced.
  $: measures = (metricsViewSpec?.measures ?? []).filter(
    (measure) => measure.name && columns.includes(measure.name),
  );

  $: expressionFunctions = measures.reduce<
    Record<string, { fn: (val: number) => string }>
  >((acc, measure) => {
    const fieldName = sanitizeFieldName(measure.name as string);
    const formatter = createMeasureValueFormatter<null | undefined>(measure);
    return {
      ...acc,
      // The formatter returns null/undefined for values it cannot handle; fall back to the raw value.
      [fieldName]: { fn: (val: number) => formatter(val) ?? String(val) },
    };
  }, {});

  $: ejected = parseVegaSpec(vegaSpec);

  // Flint reads the data to make layout decisions, so compilation waits for the rows.
  $: compiled =
    spec && !vegaSpec && $dataQuery.isSuccess
      ? compileFlintSpec({
          spec,
          rows,
          fields,
          size:
            compileWidth > 0 && compileHeight > 0
              ? { width: compileWidth, height: compileHeight }
              : undefined,
          measureFields: measures.map((measure) => measure.name as string),
        })
      : null;

  $: queryError = $dataQuery.error;
  $: isLoading = $dataQuery.isLoading;

  // Flint silently drops rows when a discrete axis overflows its layout budget, so surface it.
  $: overflowWarnings =
    compiled?.warnings.filter((warning) => warning.severity !== "info") ?? [];

  $: renderedSpec = ejected?.spec ?? compiled?.spec;
  $: specError = ejected?.error ?? compiled?.error;

  $: if (viewVL && chartContainer && plotWidth > 0 && plotHeight > 0) {
    void fitFlintGuidesToContainer(viewVL);
  }

  /**
   * Vega-Lite's `fit` autosize handles ordinary axes and legends, but an arc chart can still place
   * a wrapped legend beyond its requested height. Keep this correction local to Flint charts and
   * feed the measured overflow back into the view. The second pass covers a legend that wraps
   * differently after the first shrink.
   */
  async function fitFlintGuidesToContainer(view: View) {
    const currentRun = ++fitRun;
    await tick();
    if (currentRun !== fitRun || view !== viewVL) return;

    await view.runAsync();

    const rendererContainer = chartContainer.querySelector<HTMLElement>(
      ".rill-vega-container",
    );
    if (!rendererContainer || currentRun !== fitRun) return;

    let fittedHeight = view.height();
    for (let pass = 0; pass < 2; pass++) {
      const overflow =
        rendererContainer.scrollHeight - rendererContainer.clientHeight;
      if (overflow <= 0 || currentRun !== fitRun) break;
      fittedHeight = Math.max(0, fittedHeight - overflow);
      view.height(fittedHeight);
      await view.runAsync();
    }
  }

  // A compiled chart spec inlines the rows it read, so only an ejected spec is handed a named
  // dataset: passing Vega a dataset the spec never declares makes it throw.
  $: namedData = ejected ? { [EJECTED_DATA_NAME]: rows } : {};

  function parseVegaSpec(
    value: string | undefined,
  ): { spec: VisualizationSpec | null; error: string | null } | null {
    if (!value?.trim()) return null;
    try {
      return { spec: JSON.parse(value) as VisualizationSpec, error: null };
    } catch (e) {
      return {
        spec: null,
        error: `Invalid Vega-Lite spec: ${e instanceof Error ? e.message : String(e)}`,
      };
    }
  }
</script>

<div class="flex flex-col gap-2 h-full">
  {#if showDataTable && (spec || vegaSpec) && !specError && !queryError}
    <div class="flex flex-row items-center gap-2 p-1">
      <FieldSwitcher
        fields={viewOptions}
        selected={selectedView}
        onClick={(i) => (selectedView = i)}
        small={true}
      />
    </div>
  {/if}
  <div class="size-full flex flex-col overflow-hidden">
    {#if selectedView === 0}
      <div
        class="size-full flex flex-col"
        bind:clientWidth={chartWidth}
        bind:clientHeight={chartHeight}
      >
        {#if !spec && !vegaSpec}
          <ComponentError error={m.component_preview_no_chart_spec()} />
        {:else if !metricsSQL}
          <ComponentError error={m.component_preview_no_metrics_sql()} />
        {:else if queryError}
          <ComponentError
            error={queryError.message || m.component_preview_query_error()}
          />
        {:else if isLoading}
          <ReconcilingSpinner />
        {:else if specError}
          <ComponentError error={specError} />
        {:else if renderedSpec}
          {#if overflowWarnings.length}
            <div
              class="px-2 py-1 text-[11px] text-yellow-800 bg-yellow-50 border-b border-yellow-200"
            >
              {overflowWarnings.map((warning) => warning.message).join(" ")}
            </div>
          {/if}
          <div
            class="grow min-h-0"
            bind:this={chartContainer}
            bind:clientWidth={plotWidth}
            bind:clientHeight={plotHeight}
          >
            <VegaLiteRenderer
              {renderer}
              spec={renderedSpec}
              data={namedData}
              canvasDashboard
              {themeMode}
              config={getRillTheme(themeMode === "dark", theme)}
              {expressionFunctions}
              bind:viewVL
            />
          </div>
        {/if}
      </div>
    {:else}
      <div class="flex-1 min-h-0 min-w-0">
        {#if $dataQuery.isSuccess}
          <PreviewTable
            {rows}
            columnNames={tableColumns}
            rowHeight={32}
            {name}
          />
        {:else if isLoading}
          <ReconcilingSpinner />
        {:else if queryError}
          <ComponentError
            error={queryError.message || m.component_preview_query_error()}
          />
        {/if}
      </div>
    {/if}
  </div>
</div>
