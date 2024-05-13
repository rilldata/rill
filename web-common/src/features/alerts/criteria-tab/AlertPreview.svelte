<script lang="ts">
  import { getAlertPreviewData } from "@rilldata/web-common/features/alerts/alert-preview-data";
  import { mapAlertCriteriaToExpression } from "@rilldata/web-common/features/alerts/criteria-tab/map-alert-criteria";
  import { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { useQueryClient } from "@tanstack/svelte-query";
  import PreviewTable from "../../../components/preview-table/PreviewTable.svelte";
  import PreviewEmpty from "../PreviewEmpty.svelte";

  export let formValues: AlertFormValues;

  const queryClient = useQueryClient();

  $: alertPreviewQuery = getAlertPreviewData(queryClient, formValues);

  $: isCriteriaEmpty =
    formValues.criteria.map(mapAlertCriteriaToExpression).length === 0;
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
      name="Alert Preview"
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
