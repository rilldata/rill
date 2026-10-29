<script lang="ts">
  import {
    conditionalFormatSpecToMeasureFormatting,
    type PivotCanvasComponent,
  } from "@rilldata/web-common/features/canvas/components/pivot";
  import ComponentHeader from "../../ComponentHeader.svelte";
  import CanvasPivotRenderer from "./CanvasPivotRenderer.svelte";
  import { fieldNames, sortingFromSpec } from "./field-config";
  import { validateTableSchema } from "./selector";
  import { normalizeRowLimit, tableFieldMapper } from "./util";

  export let component: PivotCanvasComponent;
  // Set by the canvas editor; the view and embed surfaces leave it false.
  export let editable = false;

  $: ({
    parent: {
      metricsView: { getMetricsViewFromName },
    },
    specStore,
    config,
    pivotState,
    pivotDataStore,
  } = component);

  $: tableSpec = $specStore;

  $: ({
    title,
    description,
    show_description_as_tooltip,
    dimension_filters,
    time_filters,
  } = tableSpec);

  $: hasHeader = !!title || !!description;

  $: filters = {
    time_filters,
    dimension_filters,
  };

  $: _metricViewSpec = getMetricsViewFromName(tableSpec.metrics_view);
  $: metricsViewSpec = $_metricViewSpec.metricsView;

  $: schema = validateTableSchema($_metricViewSpec, tableSpec);
  $: widthScopeKey = `canvas:${component.parent.name}:${component.id}`;

  $: measureNames = new Set<string>([
    ...(metricsViewSpec?.measures?.map((m) => m.name as string) ?? []),
    ...(tableSpec.adhoc_measures?.map((measure) => measure.name) ?? []),
  ]);

  // Per-measure formatting from the YAML spec.
  $: measureFormatting = conditionalFormatSpecToMeasureFormatting(
    tableSpec.conditional_format,
  );

  // A comparison sort targets columns that only exist while the widget
  // compares, so such a spec reseeds when the comparison is toggled; other
  // specs keep the viewer's sort and expansion across the toggle. The config
  // store value is undefined on the first reactive pass.
  $: comparisonEnabled = $config?.enableComparison === true;

  // The spec keys whose change restarts the pivot (resetting sort, expansion
  // and paging). Presentation keys such as widths, wrapping, labels and number
  // formats are left out on purpose, so editing them keeps the viewer's state.
  $: dataKey = JSON.stringify({
    metricsView: tableSpec.metrics_view,
    columns: "columns" in tableSpec ? fieldNames(tableSpec.columns) : null,
    measures: "measures" in tableSpec ? fieldNames(tableSpec.measures) : null,
    rowDimensions:
      "row_dimensions" in tableSpec
        ? fieldNames(tableSpec.row_dimensions)
        : null,
    colDimensions:
      "col_dimensions" in tableSpec
        ? fieldNames(tableSpec.col_dimensions)
        : null,
    adhocMeasures: tableSpec.adhoc_measures ?? null,
    hideTotalsCol: tableSpec.hide_totals_col ?? null,
    hideTotalsRow: tableSpec.hide_totals_row ?? null,
    totalsRowPosition: tableSpec.totals_row_position ?? null,
    rowLimit: "row_limit" in tableSpec ? (tableSpec.row_limit ?? null) : null,
    sortBy: tableSpec.sort_by ?? null,
    sortDir: tableSpec.sort_dir ?? null,
    sortComparison: tableSpec.sort_comparison ?? null,
    comparisonEnabled: tableSpec.sort_comparison ? comparisonEnabled : null,
    metricsViewMeasures: metricsViewSpec?.measures?.map((m) => m.name) ?? null,
    timeDimension: metricsViewSpec?.timeDimension ?? null,
  });

  // `schema` is a new object on every spec write, so the statement re-runs
  // then too; only a changed data key actually reseeds (and so resets the
  // sort, expansion and paging).
  let seededDataKey: string | undefined;
  $: if (schema.isValid && !schema.isLoading && dataKey !== seededDataKey) {
    seededDataKey = dataKey;
    seedPivotState(dataKey);
  }

  // Formatting changes apply in place without restarting the pivot.
  let appliedFormattingKey = "";
  $: formattingKey = JSON.stringify(measureFormatting);
  $: if (formattingKey !== appliedFormattingKey) {
    appliedFormattingKey = formattingKey;
    pivotState.update((state) => ({ ...state, measureFormatting }));
  }

  // Seeds the shared pivot state from the spec. Reads the spec directly rather
  // than through reactive dependencies so it only runs when `dataKey` changes.
  function seedPivotState(_dataKey: string) {
    const spec = tableSpec;
    const isMeasure = (name: string) => measureNames.has(name);
    const sorting = sortingFromSpec(spec, isMeasure, comparisonEnabled);

    if ("columns" in spec) {
      pivotState.update((state) => ({
        ...state,
        sorting,
        expanded: {},
        activeCell: null,
        columnPage: 1,
        rowPage: 1,
        columns: tableFieldMapper(
          fieldNames(spec.columns),
          metricsViewSpec,
          spec.adhoc_measures,
        ),
        showTotalsColumn: spec.hide_totals_col !== true,
        showTotalsRow: spec.hide_totals_row !== true,
        totalsRowPosition: spec.totals_row_position ?? "top",
        measureFormatting,
      }));
    } else {
      pivotState.update((state) => ({
        ...state,
        sorting,
        expanded: {},
        activeCell: null,
        columnPage: 1,
        rowPage: 1,
        columns: [
          ...tableFieldMapper(fieldNames(spec.col_dimensions), metricsViewSpec),
          ...tableFieldMapper(
            fieldNames(spec.measures),
            metricsViewSpec,
            spec.adhoc_measures,
          ),
        ],
        rows: tableFieldMapper(
          fieldNames(spec.row_dimensions),
          metricsViewSpec,
        ),
        showTotalsColumn: spec.hide_totals_col !== true,
        showTotalsRow: spec.hide_totals_row !== true,
        totalsRowPosition: spec.totals_row_position ?? "top",
        measureFormatting,
        rowLimit: normalizeRowLimit(spec.row_limit),
        outermostRowLimit: undefined,
      }));
    }
  }
</script>

<ComponentHeader
  {component}
  {title}
  {description}
  showDescriptionAsTooltip={show_description_as_tooltip}
  {filters}
/>

<CanvasPivotRenderer
  {hasHeader}
  {schema}
  {pivotDataStore}
  pivotConfig={config}
  {pivotState}
  {component}
  {widthScopeKey}
  {editable}
/>
