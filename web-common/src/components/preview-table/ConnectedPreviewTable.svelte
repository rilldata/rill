<script lang="ts">
  import ReconcilingSpinner from "@rilldata/web-common/features/entity-management/ReconcilingSpinner.svelte";
  import {
    V1TableRowsResponseDataItem,
    createQueryServiceTableColumns,
    createQueryServiceTableRows,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import type { VirtualizedTableColumns } from "../virtualized-table/types";
  import PreviewTable from "./PreviewTable.svelte";

  export let objectName: string;
  export let limit = 150;
  export let loading = false;

  let columns: VirtualizedTableColumns[] | undefined;
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

  $: columns =
    ($profileColumnsQuery?.data?.profileColumns as VirtualizedTableColumns[]) ??
    columns; // Retain old profileColumns

  $: rows = $tableQuery?.data?.data ?? rows;
</script>

{#if loading || $tableQuery.isLoading || $profileColumnsQuery.isLoading}
  <ReconcilingSpinner />
{:else if rows && columns}
  <PreviewTable {rows} columnNames={columns} name={objectName} />
{/if}
