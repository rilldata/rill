<script lang="ts">
  import type { TableSpec } from "@rilldata/web-common/features/canvas/components/table";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { createPivotDataStore } from "@rilldata/web-common/features/dashboards/pivot/pivot-data-store";
  import {
    PivotChipType,
    type PivotDashboardContext,
    type PivotDataStore,
    type PivotDataStoreConfig,
    type PivotState,
  } from "@rilldata/web-common/features/dashboards/pivot/types";
  import type { V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";
  import { readable, type Readable, writable } from "svelte/store";
  import { getTableConfig } from "./selector";
  import TableRenderer from "./TableRenderer.svelte";

  export let rendererProperties: V1ComponentSpecRendererProperties;

  const ctx = getCanvasStateManagers();

  $: tableSpec = rendererProperties as TableSpec;

  $: colDimensions = tableSpec.col_dimensions || [];
  $: rowDimensions = tableSpec.row_dimensions || [];

  // TODO: Should we move this to canvas entity store?
  $: pivotState = writable<PivotState>({
    active: true,
    columns: {
      measure: tableSpec.measures.map((measure) => ({
        id: measure,
        title: measure,
        type: PivotChipType.Measure,
      })),
      dimension: colDimensions.map((dimension) => ({
        id: dimension,
        title: dimension,
        type: PivotChipType.Dimension,
      })),
    },
    rows: {
      dimension: rowDimensions.map((dimension) => ({
        id: dimension,
        title: dimension,
        type: PivotChipType.Dimension,
      })),
    },
    expanded: {},
    sorting: [],
    columnPage: 1,
    rowPage: 1,
    enableComparison: false,
    rowJoinType: "nest",
    activeCell: null,
  });

  let pivotDataStore: PivotDataStore | undefined = undefined;
  let pivotConfig: Readable<PivotDataStoreConfig> | undefined = undefined;
  $: {
    const pivotDashboardContext: PivotDashboardContext = {
      metricsViewName: readable(tableSpec.metrics_view),
      queryClient: ctx.queryClient,
      enabled: !!ctx.canvasEntity.spec.canvasSpec,
    };
    pivotConfig = getTableConfig(ctx, tableSpec, $pivotState);
    pivotDataStore = createPivotDataStore(pivotDashboardContext, pivotConfig);
  }
</script>

<div class="overflow-y-auto">
  {#if pivotDataStore && pivotConfig && $pivotConfig}
    <TableRenderer
      {pivotDataStore}
      config={$pivotConfig}
      pivotDashboardStore={pivotState}
    />
  {:else}
    <div>Loading...</div>
  {/if}
</div>
