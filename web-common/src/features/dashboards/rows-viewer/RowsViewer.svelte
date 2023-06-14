<script lang="ts">
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useMetaQuery } from "../selectors";
  import {
    createQueryServiceMetricsViewRows,
    createQueryServiceTableColumns,
  } from "@rilldata/web-common/runtime-client";
  import { useDashboardStore } from "../dashboard-stores";
  import PreviewTable from "@rilldata/web-common/components/preview-table/PreviewTable.svelte";

  export let metricViewName: string = "";

  $: dashboardStore = useDashboardStore(metricViewName);

  $: modelName = useMetaQuery<string>(
    $runtime.instanceId,
    metricViewName,
    (data) => data.model
  );

  $: name = $modelName?.data as string | undefined;

  $: tableQuery = createQueryServiceMetricsViewRows(
    $runtime?.instanceId,
    metricViewName,
    {
      limit: 10000,
      filter: $dashboardStore.filters,
      timeStart: $dashboardStore?.selectedTimeRange?.start,
      timeEnd: $dashboardStore.selectedTimeRange?.end,
    },
    {
      query: {
        enabled: true, // TODO: add check for filters, etc first...
      },
    }
  );

  let rows;
  $: {
    if ($tableQuery.isSuccess) {
      rows = $tableQuery.data.data;
    }
  }

  $: profileColumnsQuery = createQueryServiceTableColumns(
    $runtime?.instanceId,
    name,
    {}
  );
  $: profileColumns = $profileColumnsQuery?.data?.profileColumns;

  let rowOverscanAmount = 0;
  let columnOverscanAmount = 0;
</script>

<div class="h-80 overflow-y-auto bg-gray-100 border-t border-gray-200">
  {#if rows}
    <PreviewTable
      {rows}
      columnNames={profileColumns}
      {rowOverscanAmount}
      {columnOverscanAmount}
    />
  {/if}
</div>
