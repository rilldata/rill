<script lang="ts">
  import { getAlertPreviewData } from "@rilldata/web-common/features/alerts/alert-preview-data";
  import AlertPreviewTable from "@rilldata/web-common/features/alerts/AlertPreviewTable.svelte";
  import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import type { TimeControls } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import PreviewEmpty from "../PreviewEmpty.svelte";
  import type { DimensionTableRow } from "../../dashboards/dimension-table/dimension-table-types";
  import type { ExpressionFilterManager } from "../../dashboards/filters/manager/ExpressionFilterManager.svelte.ts";

  let {
    formValues,
    filters,
    timeControls,
  }: {
    formValues: AlertFormValues;
    filters: ExpressionFilterManager;
    timeControls: TimeControls;
  } = $props();

  const runtimeClient = useRuntimeClient();

  let alertPreviewQuery = $derived(
    getAlertPreviewData(
      runtimeClient,
      queryClient,
      {
        ...formValues,
        criteria: [],
      },
      filters.exprByMetricsView[formValues.metricsViewName],
      timeControls,
    ),
  );

  let queryResult = $derived($alertPreviewQuery);

  let rows = $derived(
    (queryResult.data?.rows as DimensionTableRow[] | undefined) ?? [],
  );
  let columns = $derived(queryResult.data?.schema ?? []);
</script>

{#if queryResult.isFetching}
  <div class="p-2 flex flex-col justify-center">
    <Spinner status={EntityStatus.Running} />
  </div>
{:else if !queryResult.data}
  <PreviewEmpty
    topLine={m.alert_form_no_data()}
    bottomLine="To see a preview, select measures above."
  />
{:else}
  <div class="max-h-64 overflow-auto">
    <AlertPreviewTable {rows} {columns} />
  </div>
{/if}
