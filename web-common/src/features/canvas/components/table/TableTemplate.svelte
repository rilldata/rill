<script lang="ts">
  import ComponentError from "@rilldata/web-common/features/canvas/components/ComponentError.svelte";
  import type { TableSpec } from "@rilldata/web-common/features/canvas/components/table";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
  import { splitPivotChips } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
  import PivotEmpty from "@rilldata/web-common/features/dashboards/pivot/PivotEmpty.svelte";
  import PivotError from "@rilldata/web-common/features/dashboards/pivot/PivotError.svelte";
  import PivotTable from "@rilldata/web-common/features/dashboards/pivot/PivotTable.svelte";
  import {
    type PivotDataStore,
    type PivotState,
  } from "@rilldata/web-common/features/dashboards/pivot/types";
  import type { V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";
  import { onDestroy } from "svelte";
  import { writable, type Readable } from "svelte/store";
  import { validateTableSchema } from "./selector";
  import {
    clearTableCache,
    getTableConfig,
    tableFieldMapper,
    usePivotForCanvas,
  } from "./util";

  export let rendererProperties: V1ComponentSpecRendererProperties;
  export let timeAndFilterStore: Readable<TimeAndFilterStore>;
  export let componentName: string;

  const ctx = getCanvasStateManagers();
  const tableSpecStore = writable(rendererProperties as TableSpec);
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

  const { getMetricsViewFromName } = ctx.canvasEntity.spec;
  let pivotDataStore: PivotDataStore | undefined;

  $: tableSpec = rendererProperties as TableSpec;
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
        ...tableFieldMapper(colDimensions, $metricViewSpec?.timeDimension),
        ...tableFieldMapper(measures, $metricViewSpec?.timeDimension),
      ],
      rows: tableFieldMapper(rowDimensions, $metricViewSpec?.timeDimension),
    }));
  }

  $: pivotConfig = getTableConfig(
    ctx,
    tableSpec.metrics_view,
    tableSpecStore,
    pivotState,
    timeAndFilterStore,
  );

  $: if ($schema.isValid && tableSpec.metrics_view) {
    pivotDataStore = usePivotForCanvas(
      ctx,
      componentName,
      tableSpec.metrics_view,
      pivotState,
      tableSpecStore,
      timeAndFilterStore,
    );
  } else {
    pivotDataStore = undefined;
    clearTableCache(componentName);
  }

  $: pivotColumns = splitPivotChips($pivotState.columns);

  $: hasColumnAndNoMeasure =
    pivotColumns.dimension.length > 0 && pivotColumns.measure.length === 0;

  onDestroy(() => {
    clearTableCache();
  });
</script>

<div class="size-full overflow-hidden" style:max-height="inherit">
  {#if !$schema.isValid}
    <ComponentError error={$schema.error} />
  {:else if pivotDataStore && $pivotDataStore && pivotConfig && $pivotConfig}
    {#if $pivotDataStore?.error?.length}
      <PivotError errors={$pivotDataStore.error} />
    {:else if !$pivotDataStore?.data || $pivotDataStore?.data?.length === 0}
      <PivotEmpty
        assembled={$pivotDataStore.assembled}
        isFetching={$pivotDataStore.isFetching}
        {hasColumnAndNoMeasure}
      />
    {:else}
      <PivotTable
        border={false}
        {pivotDataStore}
        config={pivotConfig}
        {pivotState}
        setPivotExpanded={(expanded) => {
          pivotState.update((state) => ({
            ...state,
            expanded,
          }));
        }}
        setPivotSort={(sorting) => {
          pivotState.update((state) => ({
            ...state,
            sorting,
            rowPage: 1,
            expanded: {},
          }));
        }}
        setPivotRowPage={(page) => {
          pivotState.update((state) => ({
            ...state,
            rowPage: page,
          }));
        }}
      />
    {/if}
  {/if}
</div>
