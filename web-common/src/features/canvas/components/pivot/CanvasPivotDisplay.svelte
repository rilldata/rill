<script lang="ts">
  import type { PivotCanvasComponent } from "@rilldata/web-common/features/canvas/components/pivot";
  import { isTimeDimension } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
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
  $: metricsViewLoading = $_metricViewSpec.isLoading;

  $: schema = validateTableSchema(metricsViewSpec, tableSpec, metricsViewLoading);
  $: widthScopeKey = `canvas:${component.parent.name}:${component.id}`;

  // Build accessible field lists by filtering out any fields not present in the
  // metrics view spec (e.g. excluded by a security policy).
  $: accessibleColumns =
    "columns" in tableSpec
      ? (tableSpec.columns || []).filter((c) => {
          const allMeasures =
            metricsViewSpec?.measures?.map((m) => m.name as string) || [];
          const allDimensions =
            metricsViewSpec?.dimensions?.map(
              (d) => d.name || (d.column as string),
            ) || [];
          return allMeasures.includes(c) || allDimensions.includes(c);
        })
      : [];

  $: accessibleMeasures =
    !("columns" in tableSpec)
      ? (tableSpec.measures || []).filter((m) =>
          metricsViewSpec?.measures?.some((mv) => mv.name === m),
        )
      : [];

  $: accessibleRowDimensions =
    !("columns" in tableSpec)
      ? (tableSpec.row_dimensions || []).filter(
          (d) =>
            metricsViewSpec?.dimensions?.some(
              (mv) => mv.name === d || mv.column === d,
            ) ||
            (metricsViewSpec?.timeDimension !== undefined &&
              isTimeDimension(d, metricsViewSpec.timeDimension)),
        )
      : [];

  $: accessibleColDimensions =
    !("columns" in tableSpec)
      ? (tableSpec.col_dimensions || []).filter(
          (d) =>
            metricsViewSpec?.dimensions?.some(
              (mv) => mv.name === d || mv.column === d,
            ) ||
            (metricsViewSpec?.timeDimension !== undefined &&
              isTimeDimension(d, metricsViewSpec.timeDimension)),
        )
      : [];

  $: if ("columns" in tableSpec && schema.isValid) {
    pivotState.update((state) => ({
      ...state,
      sorting: [],
      expanded: {},
      activeCell: null,
      columnPage: 1,
      rowPage: 1,
      columns: tableFieldMapper(accessibleColumns, metricsViewSpec),
      showTotalsColumn: tableSpec.hide_totals_col !== true,
      showTotalsRow: tableSpec.hide_totals_row !== true,
    }));
  } else if (!("columns" in tableSpec) && schema.isValid) {
    pivotState.update((state) => ({
      ...state,
      sorting: [],
      expanded: {},
      activeCell: null,
      columnPage: 1,
      rowPage: 1,
      columns: [
        ...tableFieldMapper(accessibleColDimensions, metricsViewSpec),
        ...tableFieldMapper(accessibleMeasures, metricsViewSpec),
      ],
      rows: tableFieldMapper(accessibleRowDimensions, metricsViewSpec),
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
