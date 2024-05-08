<script lang="ts">
  import ReconcilingSpinner from "@rilldata/web-common/features/entity-management/ReconcilingSpinner.svelte";
  import {
    V1TableRowsResponseDataItem,
    createQueryServiceTableColumns,
    createQueryServiceTableRows,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import WorkspaceError from "../WorkspaceError.svelte";
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

  $: columnsQuery = createQueryServiceTableColumns(
    $runtime?.instanceId,
    table,
    {
      connector,
      database,
      databaseSchema,
    },
  );

  $: rowsQuery = createQueryServiceTableRows($runtime?.instanceId, table, {
    connector,
    database,
    databaseSchema,
    limit,
  });

  $: columns =
    ($columnsQuery?.data?.profileColumns as VirtualizedTableColumns[]) ??
    columns; // Retain old profileColumns

  $: rows = $rowsQuery?.data?.data ?? rows;
</script>

{#if loading || $rowsQuery.isLoading || $columnsQuery.isLoading}
  <ReconcilingSpinner />
{:else if $rowsQuery.isError || $columnsQuery.isError}
  <WorkspaceError
    message={`Error loading table: ${$rowsQuery.error?.response.data.message || $columnsQuery.error?.response.data.message}`}
  />
{:else if rows && columns}
  <PreviewTable {rows} columnNames={columns} name={table} />
{/if}
