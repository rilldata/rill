<script lang="ts">
  import { getAlertPreviewData } from "@rilldata/web-common/features/alerts/alert-preview-data";
  import { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { useQueryClient } from "@tanstack/svelte-query";
  import PreviewTable from "../../../components/preview-table/PreviewTable.svelte";
  import PreviewEmpty from "../PreviewEmpty.svelte";

  export let formValues: AlertFormValues;

  const queryClient = useQueryClient();

  $: alertPreviewQuery = getAlertPreviewData(queryClient, {
    ...formValues,
    criteria: [],
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
      name="Data Preview"
      rows={$alertPreviewQuery.data.rows}
      columnNames={$alertPreviewQuery.data.schema}
    />
  </div>
{/if}
