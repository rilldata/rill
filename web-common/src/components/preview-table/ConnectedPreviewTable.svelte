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

  export let connector: string;
  export let database: string = ""; // The backend interprets an empty string as the default database
  export let databaseSchema: string = ""; // The backend interprets an empty string as the default schema
  export let table: string;
  export let limit = 150;
  export let loading = false;

  let columns: VirtualizedTableColumns[] | undefined;
  let rows: V1TableRowsResponseDataItem[] | undefined;

  $: profileColumnsQuery = createQueryServiceTableColumns(
    $runtime?.instanceId,
    table,
    {
      connector,
      database,
      databaseSchema,
    },
  );

  $: tableQuery = createQueryServiceTableRows($runtime?.instanceId, table, {
    connector,
    database,
    databaseSchema,
    limit,
  });

  $: columns =
    ($profileColumnsQuery?.data?.profileColumns as VirtualizedTableColumns[]) ??
    columns; // Retain old profileColumns

  $: rows = $tableQuery?.data?.data ?? rows;
</script>

{#if loading || $tableQuery.isLoading || $profileColumnsQuery.isLoading}
  <ReconcilingSpinner />
{:else if rows && columns}
  <PreviewTable {rows} columnNames={columns} name={table} />
{/if}
