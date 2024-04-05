<script lang="ts">
  import PreviewTable from "@rilldata/web-common/components/preview-table/PreviewTable.svelte";
  import ReconcilingSpinner from "@rilldata/web-common/features/entity-management/ReconcilingSpinner.svelte";
  import {
    V1TableRowsResponseDataItem,
    createQueryServiceTableColumns,
    createQueryServiceTableRows,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import type { VirtualizedTableColumns } from "../virtualized-table/types";

  export let objectName: string;
  export let limit = 150;
  export let loading = false;

  let profileColumns: VirtualizedTableColumns[] | undefined;
  let rows: V1TableRowsResponseDataItem[] | undefined;

  $: profileColumnsQuery = createQueryServiceTableColumns(
    $runtime?.instanceId,
    objectName,
    {},
  );

  $: tableQuery = createQueryServiceTableRows(
    $runtime?.instanceId,
    objectName,
    {
      limit,
    },
  );

  $: profileColumns =
    ($profileColumnsQuery?.data?.profileColumns as VirtualizedTableColumns[]) ??
    profileColumns; // Retain old profileColumns

  $: rows = $tableQuery?.data?.data ?? rows;
</script>

{#if loading || $tableQuery.isLoading || $profileColumnsQuery.isLoading}
  <ReconcilingSpinner />
{:else if rows && profileColumns}
  <PreviewTable {rows} columnNames={profileColumns} rowOverscanAmount={10} />
{/if}
