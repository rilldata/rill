<script lang="ts">
  import { getAlertPreviewData } from "@rilldata/web-common/features/alerts/alert-preview-data";
  import AlertPreviewTable from "@rilldata/web-common/features/alerts/AlertPreviewTable.svelte";
  import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import { mapMeasureFilterToExpr } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import type { Filters } from "@rilldata/web-common/features/dashboards/stores/Filters.ts";
  import type { TimeControls } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
  import PreviewEmpty from "../PreviewEmpty.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type { DimensionTableRow } from "../../dashboards/dimension-table/dimension-table-types";

  export let formValues: AlertFormValues;
  export let filters: Filters;
  export let timeControls: TimeControls;

  $: alertPreviewQuery = getAlertPreviewData(
    queryClient,
    formValues,
    filters,
    timeControls,
  );

  $: isCriteriaEmpty =
    formValues.criteria.map(mapMeasureFilterToExpr).length === 0;

  $: queryResult = alertPreviewQuery;

  $: rows = (queryResult.data?.rows as DimensionTableRow[] | undefined) ?? [];
  $: columns = queryResult.data?.schema ?? [];
</script>

{#if alertPreviewQuery.isFetching}
  <div class="p-2 flex flex-col justify-center">
    <Spinner status={EntityStatus.Running} />
  </div>
{:else if isCriteriaEmpty || !alertPreviewQuery.data}
  <PreviewEmpty
    topLine="No criteria selected"
    bottomLine="Select criteria to see a preview"
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
