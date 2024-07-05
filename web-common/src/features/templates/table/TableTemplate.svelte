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

  $: instanceId = $runtime.instanceId;
  const queryClient = useQueryClient();

  $: tableProperties = rendererProperties as TableProperties;

  $: pivotState = writable<PivotState>({
    active: true,
    columns: {
      measure: tableProperties.measures.map((measure) => ({
        id: measure,
        title: measure,
        type: PivotChipType.Measure,
      })),
      dimension: tableProperties.col_dimensions.map((dimension) => ({
        id: dimension,
        title: dimension,
        type: PivotChipType.Dimension,
      })),
    },
    rows: {
      dimension: tableProperties.row_dimensions.map((dimension) => ({
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

  $: stateManagerContext = createStateManagers({
    queryClient,
    metricsViewName: tableProperties.metric_view,
    extraKeyPrefix: "_custom-table",
  });

  $: pivotDataStore = createPivotDataStore(stateManagerContext, pivotConfig);
</script>

<div>
  {#if $pivotDataStore}
    <TableRenderer
      metricsViewName={tableProperties.metric_view + "_custom-table"}
      {pivotDataStore}
      config={$pivotConfig}
      pivotDashboardStore={pivotState}
    />
  {/if}
</div>
