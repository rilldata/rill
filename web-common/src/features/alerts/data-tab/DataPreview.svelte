<script lang="ts">
  import { getAlertPreviewData } from "@rilldata/web-common/features/alerts/alert-preview-data";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { useQueryClient } from "@tanstack/svelte-query";

  import PreviewTable from "../../../components/preview-table/PreviewTable.svelte";
  import type { V1Expression, V1TimeRange } from "../../../runtime-client";
  import PreviewEmpty from "../PreviewEmpty.svelte";

  export let metricsViewName: string;
  export let measure: string;
  export let splitByDimension: string;
  export let whereFilter: V1Expression;
  export let timeRange: V1TimeRange;

  const queryClient = useQueryClient();

  $: alertPreviewQuery = getAlertPreviewData(queryClient, {
    metricsViewName,
    measure,
    splitByDimension,
    whereFilter,
    criteria: undefined,
    timeRange,
  });
</script>

{#if $alertPreviewQuery.isFetching}
  <div class="p-2 flex flex-col justify-center">
    <Spinner status={EntityStatus.Running} />
  </div>
{:else if !$alertPreviewQuery.data}
  <PreviewEmpty
    topLine="No data to preview"
    bottomLine="To see a preview, select measures above."
  />
{:else}
  <div class="max-h-64 overflow-auto">
    <PreviewTable
      rows={$alertPreviewQuery.data.rows}
      columns={$alertPreviewQuery.data.schema}
    />
  </div>
{/if}
