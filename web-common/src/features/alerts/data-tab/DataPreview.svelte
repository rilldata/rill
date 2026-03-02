<script lang="ts">
  import { getAlertPreviewData } from "@rilldata/web-common/features/alerts/alert-preview-data";
  import AlertPreviewTable from "@rilldata/web-common/features/alerts/AlertPreviewTable.svelte";
  import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import type { Filters } from "@rilldata/web-common/features/dashboards/stores/Filters.ts";
  import type { TimeControls } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import PreviewEmpty from "../PreviewEmpty.svelte";
  import type { DimensionTableRow } from "../../dashboards/dimension-table/dimension-table-types";

  export let formValues: AlertFormValues;
  export let filters: Filters;
  export let timeControls: TimeControls;

  const runtimeClient = useRuntimeClient();

  $: alertPreviewQuery = getAlertPreviewData(
    runtimeClient,
    queryClient,
    {
      ...formValues,
      criteria: [],
    },
    filters,
    timeControls,
  );

  $: queryResult = $alertPreviewQuery;

  $: rows = (queryResult.data?.rows as DimensionTableRow[] | undefined) ?? [];
  $: columns = queryResult.data?.schema ?? [];
</script>

{#if queryResult.isFetching}
  <div class="p-2 flex flex-col justify-center">
    <Spinner status={EntityStatus.Running} />
  </div>
{:else if !queryResult.data}
  <PreviewEmpty
    topLine="No data to preview"
    bottomLine="To see a preview, select measures above."
  />
{:else}
  <div class="max-h-64 overflow-auto">
    <AlertPreviewTable {rows} {columns} />
  </div>
{/if}
