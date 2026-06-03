<script lang="ts">
  import type { PivotCanvasComponent } from "@rilldata/web-common/features/canvas/components/pivot";
  import ComponentHeader from "../../ComponentHeader.svelte";
  import CanvasPivotRenderer from "./CanvasPivotRenderer.svelte";
  import { validateTableSchema } from "./selector";
  import { tableFieldMapper } from "./util";

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

  $: schema = validateTableSchema(metricsViewSpec, tableSpec);
  $: widthScopeKey = `canvas:${component.parent.name}:${component.id}`;

  $: if ("columns" in tableSpec && schema.isValid) {
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
    }));
  } else if ("col_dimensions" in tableSpec && schema.isValid) {
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
