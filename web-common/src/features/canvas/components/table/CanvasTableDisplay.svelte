<script lang="ts">
  import CanvasPivotRenderer from "@rilldata/web-common/features/canvas/components/pivot/CanvasPivotRenderer.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import {
    type PivotDataStore,
    type PivotDataStoreConfig,
    type PivotState,
  } from "@rilldata/web-common/features/dashboards/pivot/types";
  import { onDestroy } from "svelte";
  import { writable, type Readable } from "svelte/store";
  import type { TableCanvasComponent, TableSpec } from ".";
  import {
    clearTableCache,
    tableFieldMapper,
    usePivotForCanvas,
  } from "../pivot/util";
  import { validateTableSchema } from "./selector";
  import { useTableConfig } from "./util";

  export let component: TableCanvasComponent;

  $: ({
    specStore,
    timeAndFilterStore,
    id: componentName,
    parent: { name: canvasName },
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
    tableMode: "flat",
    activeCell: null,
  });

  $: ({ getMetricsViewFromName } = ctx.canvasEntity.spec);
  let pivotDataStore: PivotDataStore | undefined;
  let tableConfig: Readable<PivotDataStoreConfig> | undefined;

  $: tableSpec = rendererProperties as TableSpec;
  $: tableSpecStore.set(tableSpec);

  $: columns = tableSpec?.columns || [];

  $: metricViewSpec = getMetricsViewFromName(tableSpec.metrics_view);

  $: schema = validateTableSchema(ctx, tableSpec);

  $: if (tableSpec && $schema.isValid) {
    pivotState.update((state) => ({
      ...state,
      sorting: [],
      expanded: {},
      columns: tableFieldMapper(columns, $metricViewSpec),
    }));
  }

  $: if ($schema.isValid && tableSpec.metrics_view) {
    tableConfig = useTableConfig(
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
      tableConfig,
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
  pivotConfig={tableConfig}
  {pivotState}
/>
