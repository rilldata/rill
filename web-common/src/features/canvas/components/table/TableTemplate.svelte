<script lang="ts">
  import { createPivotDataStore } from "@rilldata/web-common/features/dashboards/pivot/pivot-data-store";
  import {
    PivotChipType,
    type PivotDataStore,
    type PivotDataStoreConfig,
    type PivotState,
  } from "@rilldata/web-common/features/dashboards/pivot/types";
  import { createStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import type { TableProperties } from "@rilldata/web-common/features/templates/types";
  import type { V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { type Readable, writable } from "svelte/store";
  import { getTableConfig, hasValidTableSchema } from "./selector";
  import TableRenderer from "./TableRenderer.svelte";

  export let rendererProperties: V1ComponentSpecRendererProperties;

  const queryClient = useQueryClient();
  const TABLE_PREFIX = "_custom-table";

  $: instanceId = $runtime.instanceId;

  $: tableProperties = rendererProperties as TableProperties;

  $: tableSchema = hasValidTableSchema(instanceId, tableProperties);

  $: isValidSchema = $tableSchema.isValid;

  $: colDimensions = tableProperties.col_dimensions || [];
  $: rowDimensions = tableProperties.row_dimensions || [];
  $: whereSql = tableProperties.filter;

  $: pivotState = writable<PivotState>({
    active: true,
    columns: {
      measure: tableProperties.measures.map((measure) => ({
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
    whereSql,
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
  $: if (isValidSchema) {
    const stateManagerContext = createStateManagers({
      queryClient,
      exploreName: "TODO", // Historically, State Managers have only been used for Explore, not Canvas.
      metricsViewName: tableProperties.metrics_view,
      extraKeyPrefix: TABLE_PREFIX,
    });

    pivotConfig = getTableConfig(instanceId, tableProperties, $pivotState);
    pivotDataStore = createPivotDataStore(stateManagerContext, pivotConfig);
  }
</script>

<div class="overflow-y-scroll">
  {#if !isValidSchema}
    <div>{$tableSchema.error}</div>
  {:else if pivotDataStore && pivotConfig && $pivotConfig}
    <TableRenderer
      {pivotDataStore}
      config={$pivotConfig}
      pivotDashboardStore={pivotState}
    />
  {:else}
    <div>Loading...</div>
  {/if}
</div>
