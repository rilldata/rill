<script lang="ts">
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useMetaQuery, useModelHasTimeSeries } from "../selectors";
  import {
    createQueryServiceMetricsViewRows,
    createQueryServiceTableColumns,
  } from "@rilldata/web-common/runtime-client";
  import { useDashboardStore } from "../dashboard-stores";
  import PreviewTable from "@rilldata/web-common/components/preview-table/PreviewTable.svelte";

  export let metricViewName = "";

  $: dashboardStore = useDashboardStore(metricViewName);

  $: modelName = useMetaQuery<string>(
    $runtime.instanceId,
    metricViewName,
    (data) => data.model
  );

  $: name = $modelName?.data as string | undefined;

  $: metricTimeSeries = useModelHasTimeSeries(
    $runtime.instanceId,
    metricViewName
  );
  $: hasTimeSeries = $metricTimeSeries.data;

  $: timeStart = $dashboardStore.selectedTimeRange?.start?.toISOString();
  $: timeEnd = $dashboardStore.selectedTimeRange?.end?.toISOString();

  $: tableQuery = createQueryServiceMetricsViewRows(
    $runtime?.instanceId,
    metricViewName,
    {
      limit: 10000,
      filter: $dashboardStore.filters,
      timeStart: timeStart,
      timeEnd: timeEnd,
    },
    {
      query: {
        enabled:
          (hasTimeSeries ? !!timeStart && !!timeEnd : true) &&
          !!$dashboardStore?.filters,
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

  const configOverride = {
    indexWidth: 72,
    rowHeight: 32,
  };
</script>

<div class="h-56 overflow-y-auto bg-gray-100 border-t border-gray-200">
  {#if rows}
    <PreviewTable
      {rows}
      columnNames={profileColumns}
      {rowOverscanAmount}
      {columnOverscanAmount}
      {configOverride}
    />
  {/if}
</div>
