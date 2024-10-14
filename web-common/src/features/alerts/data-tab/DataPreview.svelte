<script lang="ts">
  import { getAlertPreviewData } from "@rilldata/web-common/features/alerts/alert-preview-data";
  import AlertPreviewTable from "@rilldata/web-common/features/alerts/AlertPreviewTable.svelte";
  import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { useQueryClient } from "@tanstack/svelte-query";
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
    <AlertPreviewTable
      rows={$alertPreviewQuery.data.rows}
      columns={$alertPreviewQuery.data.schema}
    />
  </div>
{/if}
