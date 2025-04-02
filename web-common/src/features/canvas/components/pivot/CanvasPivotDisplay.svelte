<script lang="ts">
  import type {
    PivotCanvasComponent,
    PivotSpec,
  } from "@rilldata/web-common/features/canvas/components/pivot";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import {
    type PivotDataStore,
    type PivotDataStoreConfig,
    type PivotState,
  } from "@rilldata/web-common/features/dashboards/pivot/types";
  import { onDestroy } from "svelte";
  import { writable, type Readable } from "svelte/store";
  import CanvasPivotRenderer from "./CanvasPivotRenderer.svelte";
  import { validateTableSchema } from "./selector";
  import {
    clearTableCache,
    tableFieldMapper,
    usePivotConfig,
    usePivotForCanvas,
  } from "./util";

  export let component: PivotCanvasComponent;

  $: ({
    parent: { name: canvasName },
    specStore,
    id: componentName,
    timeAndFilterStore,
  } = component);

  $: rendererProperties = $specStore;

  $: hasHeader =
    !!rendererProperties?.title || !!rendererProperties?.description;

  $: ctx = getCanvasStore(canvasName);
  const tableSpecStore = writable(rendererProperties);
  const pivotState = writable<PivotState>({
    active: true,
    columns: [],
    rows: [],
    expanded: {},
    sorting: [],
    columnPage: 1,
    rowPage: 1,
    enableComparison: false,
    tableMode: "nest",
    activeCell: null,
  });

  $: ({ getMetricsViewFromName } = ctx.canvasEntity.spec);
  let pivotDataStore: PivotDataStore | undefined;
  let pivotConfig: Readable<PivotDataStoreConfig> | undefined;

  $: tableSpec = rendererProperties as PivotSpec;
  $: tableSpecStore.set(tableSpec);

  $: measures = tableSpec.measures || [];
  $: colDimensions = tableSpec.col_dimensions || [];
  $: rowDimensions = tableSpec.row_dimensions || [];

  $: metricViewSpec = getMetricsViewFromName(tableSpec.metrics_view);

  $: schema = validateTableSchema(ctx, tableSpec);

  $: if (tableSpec && $schema.isValid) {
    pivotState.update((state) => ({
      ...state,
      sorting: [],
      expanded: {},
      columns: [
        ...tableFieldMapper(colDimensions, $metricViewSpec),
        ...tableFieldMapper(measures, $metricViewSpec),
      ],
      rows: tableFieldMapper(rowDimensions, $metricViewSpec),
    }));
  }

  $: if ($schema.isValid && tableSpec.metrics_view) {
    pivotConfig = usePivotConfig(
      ctx,
      tableSpec.metrics_view,
      tableSpecStore,
      pivotState,
      timeAndFilterStore,
    );

    pivotDataStore = usePivotForCanvas(
      ctx,
      componentName,
      tableSpec.metrics_view,
      pivotConfig,
    );
  } else {
    pivotDataStore = undefined;
    clearTableCache(componentName);
  }

  onDestroy(() => {
    clearTableCache();
  });
</script>

<CanvasPivotRenderer
  {hasHeader}
  {schema}
  {pivotDataStore}
  {pivotConfig}
  {pivotState}
/>
