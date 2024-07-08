<script lang="ts">
  import { createPivotDataStore } from "@rilldata/web-common/features/dashboards/pivot/pivot-data-store";
  import {
    PivotChipType,
    PivotState,
  } from "@rilldata/web-common/features/dashboards/pivot/types";
  import { createStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import TableRenderer from "@rilldata/web-common/features/templates/table/TableRenderer.svelte";
  import { TableProperties } from "@rilldata/web-common/features/templates/types";
  import { V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { writable } from "svelte/store";
  import { getTableConfig } from "./selector";

  export let rendererProperties: V1ComponentSpecRendererProperties;

  const queryClient = useQueryClient();
  const TABLE_PREFIX = "_custom-table";

  $: instanceId = $runtime.instanceId;

  $: tableProperties = rendererProperties as TableProperties;

  $: colDimensions = tableProperties.col_dimensions || [];
  $: rowDimensions = tableProperties.row_dimensions || [];

  $: stateManagerContext = createStateManagers({
    queryClient,
    metricsViewName: tableProperties.metric_view,
    extraKeyPrefix: TABLE_PREFIX,
  });

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
    expanded: {},
    sorting: [],
    columnPage: 1,
    rowPage: 1,
    enableComparison: false,
    rowJoinType: "nest",
  });

  $: pivotConfig = getTableConfig(instanceId, tableProperties, $pivotState);
  $: pivotDataStore = createPivotDataStore(stateManagerContext, pivotConfig);
</script>

<div>
  {#if $pivotDataStore}
    <TableRenderer
      metricsViewName={tableProperties.metric_view + TABLE_PREFIX}
      {pivotDataStore}
      config={$pivotConfig}
      pivotDashboardStore={pivotState}
    />
  {/if}
</div>
