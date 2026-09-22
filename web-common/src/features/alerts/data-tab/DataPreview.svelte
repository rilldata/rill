<script lang="ts">
  import { getAlertPreviewData } from "@rilldata/web-common/features/alerts/alert-preview-data";
  import AlertPreviewTable from "@rilldata/web-common/features/alerts/AlertPreviewTable.svelte";
  import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import PreviewEmpty from "../PreviewEmpty.svelte";
  import type { DimensionTableRow } from "../../dashboards/dimension-table/dimension-table-types";
  import type { ExpressionFilterManager } from "../../dashboards/filters/ExpressionFilterManager.svelte.ts";
  import type { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";

  let {
    formValues,
    expressionFilterManager,
    timeFilterManager,
  }: {
    formValues: AlertFormValues;
    expressionFilterManager: ExpressionFilterManager;
    timeFilterManager: TimeFilterManager;
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
      expressionFilterManager.topLevelJoiner.expr[formValues.metricsViewName],
      timeFilterManager,
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
