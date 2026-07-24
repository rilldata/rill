<script lang="ts">
  import { getAlertPreviewData } from "@rilldata/web-common/features/alerts/alert-preview-data";
  import AlertPreviewTable from "@rilldata/web-common/features/alerts/AlertPreviewTable.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import { mapMeasureFilterToExpr } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import type { TimeControls } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
  import PreviewEmpty from "../PreviewEmpty.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { DimensionTableRow } from "../../dashboards/dimension-table/dimension-table-types";
  import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/manager/expression-filter-manager.svelte.ts";

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
      formValues,
      filters.expr,
      timeControls,
    ),
  );

  let isCriteriaEmpty = $derived(
    formValues.criteria.map(mapMeasureFilterToExpr).length === 0,
  );

  let queryResult = $derived($alertPreviewQuery);

  let rows = $derived(
    (queryResult.data?.rows as DimensionTableRow[] | undefined) ?? [],
  );
  let columns = $derived(queryResult.data?.schema ?? []);
</script>

{#if $alertPreviewQuery.isFetching}
  <div class="p-2 flex flex-col justify-center">
    <Spinner status={EntityStatus.Running} />
  </div>
{:else if isCriteriaEmpty || !$alertPreviewQuery.data}
  <PreviewEmpty
    topLine={m.alert_form_no_criteria()}
    bottomLine={m.alert_form_select_criteria()}
  />
{:else if rows.length > 0}
  <div class="max-h-64 overflow-auto">
    <AlertPreviewTable {rows} {columns} />
  </div>
{:else}
  <div>
    Given the above criteria, this alert will not trigger for the current time
    range.
  </div>
{/if}
