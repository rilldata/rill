<script lang="ts">
  import type { PivotCanvasComponent } from "@rilldata/web-common/features/canvas/components/pivot";
  import ComponentHeader from "../../ComponentHeader.svelte";
  import CanvasPivotRenderer from "./CanvasPivotRenderer.svelte";
  import { validateTableSchema } from "./selector";
  import { tableFieldMapper } from "./util";

  export let component: PivotCanvasComponent;

  $: ({
    parent: {
      spec: { getMetricsViewFromName },
    },
    specStore,
    config,
    pivotState,
    pivotDataStore,
  } = component);

  $: tableSpec = $specStore;

  $: ({ title, description, dimension_filters, time_filters } = tableSpec);

  $: hasHeader = !!title || !!description;

  $: filters = {
    time_filters,
    dimension_filters,
  };

  $: _metricViewSpec = getMetricsViewFromName(tableSpec.metrics_view);
  $: metricsViewSpec = $_metricViewSpec.metricsView;

  $: schema = validateTableSchema(metricsViewSpec, tableSpec);

  $: if ("columns" in tableSpec && schema.isValid) {
    const columns = tableSpec?.columns || [];
    pivotState.update((state) => ({
      ...state,
      sorting: [],
      expanded: {},
      columns: tableFieldMapper(columns, metricsViewSpec),
    }));
  } else if ("col_dimensions" in tableSpec && schema.isValid) {
    const measures = tableSpec.measures || [];
    const colDimensions = tableSpec.col_dimensions || [];
    const rowDimensions = tableSpec.row_dimensions || [];
    pivotState.update((state) => ({
      ...state,
      sorting: [],
      expanded: {},
      columns: [
        ...tableFieldMapper(colDimensions, metricsViewSpec),
        ...tableFieldMapper(measures, metricsViewSpec),
      ],
      rows: tableFieldMapper(rowDimensions, metricsViewSpec),
    }));
  }
</script>

<ComponentHeader {component} {title} {description} {filters} />

<CanvasPivotRenderer
  {hasHeader}
  {schema}
  {pivotDataStore}
  pivotConfig={config}
  {pivotState}
/>
