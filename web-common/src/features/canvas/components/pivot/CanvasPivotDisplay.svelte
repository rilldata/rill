<script lang="ts">
  import {
    conditionalFormatSpecToMeasureFormatting,
    type PivotCanvasComponent,
  } from "@rilldata/web-common/features/canvas/components/pivot";
  import ComponentHeader from "../../ComponentHeader.svelte";
  import CanvasPivotRenderer from "./CanvasPivotRenderer.svelte";
  import { validateTableSchema } from "./selector";
  import { normalizeRowLimit, tableFieldMapper } from "./util";

  export let component: PivotCanvasComponent;

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

  // Seed the shared pivot state with per-measure formatting from the YAML spec.
  $: measureFormatting = conditionalFormatSpecToMeasureFormatting(
    tableSpec.conditional_format,
  );

  $: if ("columns" in tableSpec && schema.isValid && !schema.isLoading) {
    const columns = tableSpec?.columns || [];
    pivotState.update((state) => ({
      ...state,
      sorting: [],
      expanded: {},
      activeCell: null,
      columnPage: 1,
      rowPage: 1,
      columns: tableFieldMapper(columns, metricsViewSpec),
      showTotalsColumn: tableSpec.hide_totals_col !== true,
      showTotalsRow: tableSpec.hide_totals_row !== true,
      measureFormatting,
    }));
  } else if (!("columns" in tableSpec) && schema.isValid && !schema.isLoading) {
    const measures = tableSpec.measures || [];
    const colDimensions = tableSpec.col_dimensions || [];
    const rowDimensions = tableSpec.row_dimensions || [];
    pivotState.update((state) => ({
      ...state,
      sorting: [],
      expanded: {},
      activeCell: null,
      columnPage: 1,
      rowPage: 1,
      columns: [
        ...tableFieldMapper(colDimensions, metricsViewSpec),
        ...tableFieldMapper(measures, metricsViewSpec),
      ],
      rows: tableFieldMapper(rowDimensions, metricsViewSpec),
      showTotalsColumn: tableSpec.hide_totals_col !== true,
      showTotalsRow: tableSpec.hide_totals_row !== true,
      measureFormatting,
      rowLimit: normalizeRowLimit(tableSpec.row_limit),
      outermostRowLimit: undefined,
    }));
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
/>
