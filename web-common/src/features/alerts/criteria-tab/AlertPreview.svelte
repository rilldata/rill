<script lang="ts">
  import { getAlertPreviewData } from "@rilldata/web-common/features/alerts/alert-preview-data";
  import AlertPreviewTable from "@rilldata/web-common/features/alerts/AlertPreviewTable.svelte";
  import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import { mapMeasureFilterToExpr } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { useQueryClient } from "@tanstack/svelte-query";
  import PreviewEmpty from "../PreviewEmpty.svelte";

  export let formValues: AlertFormValues;

  const queryClient = useQueryClient();

  $: alertPreviewQuery = getAlertPreviewData(queryClient, formValues);

  $: isCriteriaEmpty =
    formValues.criteria.map(mapMeasureFilterToExpr).length === 0;
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
    <AlertPreviewTable
      rows={$alertPreviewQuery.data.rows}
      columns={$alertPreviewQuery.data.schema}
    />
  </div>
{:else}
  <div>
    Given the above criteria, this alert will not trigger for the current time
    range.
  </div>
{/if}
