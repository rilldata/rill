<script lang="ts">
  import { getAlertPreviewData } from "@rilldata/web-common/features/alerts/alert-preview-data";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { useQueryClient } from "@tanstack/svelte-query";
  import PreviewTable from "../../../components/preview-table/PreviewTable.svelte";
  import type { V1Expression, V1TimeRange } from "../../../runtime-client";
  import { isExpressionIncomplete } from "../../dashboards/stores/filter-utils";
  import PreviewEmpty from "../PreviewEmpty.svelte";

  export let metricsViewName: string;
  export let measure: string;
  export let splitByDimension: string;
  export let splitByTimeGrain: string;
  export let whereFilter: V1Expression;
  export let criteria: V1Expression;
  export let timeRange: V1TimeRange;

  const queryClient = useQueryClient();

  $: alertPreviewQuery = getAlertPreviewData(queryClient, {
    metricsViewName,
    measure,
    splitByDimension,
    splitByTimeGrain,
    whereFilter,
    criteria,
    timeRange,
  });

  $: isCriteriaEmpty = isExpressionIncomplete(criteria);
</script>

{#if $alertPreviewQuery.isFetching}
  <div class="p-2 flex flex-col justify-center">
    <Spinner status={EntityStatus.Running} />
  </div>
{:else if isCriteriaEmpty || !$alertPreviewQuery.data}
  <PreviewEmpty
    topLine="No criteria selected"
    bottomLine="Select criteria to see a preview"
  />
{:else if $alertPreviewQuery.data.rows.length > 0}
  <div class="max-h-64 overflow-auto">
    <PreviewTable
      rows={$alertPreviewQuery.data.rows}
      columnNames={$alertPreviewQuery.data.schema}
    />
  </div>
{:else}
  <div>
    Given the above criteria, this alert will not trigger for the current time
    range.
  </div>
{/if}
