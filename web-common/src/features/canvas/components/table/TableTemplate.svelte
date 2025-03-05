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
    PivotChipType,
    type PivotDataStore,
  } from "@rilldata/web-common/features/dashboards/pivot/types";
  import type { V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";
  import { writable, type Readable } from "svelte/store";
  import {
    getTableConfig,
    pivotState,
    usePivotForCanvas,
    validateTableSchema,
  } from "./selector";

  export let rendererProperties: V1ComponentSpecRendererProperties;
  export let timeAndFilterStore: Readable<TimeAndFilterStore>;

  const ctx = getCanvasStateManagers();
  const tableSpecStore = writable(rendererProperties as TableSpec);

  let pivotDataStore: PivotDataStore;
  let isFetching = false;
  let assembled = false;

  $: tableSpec = rendererProperties as TableSpec;
  $: tableSpecStore.set(tableSpec);

  $: measures = tableSpec.measures || [];
  $: colDimensions = tableSpec.col_dimensions || [];
  $: rowDimensions = tableSpec.row_dimensions || [];

  $: schema = validateTableSchema(ctx, tableSpec);

  $: if (tableSpec && $schema.isValid) {
    pivotState.update((state) => ({
      ...state,
      columns: [
        ...colDimensions.map((dimension) => ({
          id: dimension,
          title: dimension,
          type: PivotChipType.Dimension,
        })),
        ...measures.map((measure) => ({
          id: measure,
          title: measure,
          type: PivotChipType.Measure,
        })),
      ],
      rows: rowDimensions.map((dimension) => ({
        id: dimension,
        title: dimension,
        type: PivotChipType.Dimension,
      })),
    }));
  }

  $: pivotConfig = getTableConfig(
    ctx,
    tableSpec.metrics_view,
    tableSpecStore,
    pivotState,
    timeAndFilterStore,
  );

  $: if ($schema.isValid && tableSpec.metrics_view && !pivotDataStore) {
    pivotDataStore = usePivotForCanvas(
      ctx,
      tableSpec.metrics_view,
      tableSpecStore,
      timeAndFilterStore,
    );
    ({ isFetching, assembled } = $pivotDataStore);
  }

  $: pivotColumns = splitPivotChips($pivotState.columns);

  $: hasColumnAndNoMeasure =
    pivotColumns.dimension.length > 0 && pivotColumns.measure.length === 0;
</script>

<div class="size-full overflow-hidden" style:max-height="inherit">
  {#if !$schema.isValid}
    <ComponentError error={$schema.error} />
  {:else if pivotDataStore && $pivotDataStore && pivotConfig && $pivotConfig}
    {#if $pivotDataStore?.error?.length}
      <PivotError errors={$pivotDataStore.error} />
    {:else if !$pivotDataStore?.data || $pivotDataStore?.data?.length === 0}
      <PivotEmpty {assembled} {isFetching} {hasColumnAndNoMeasure} />
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
</div>
