<script lang="ts">
  import ReconcilingSpinner from "@rilldata/web-common/features/entity-management/ReconcilingSpinner.svelte";
  import {
    type V1TableRowsResponseDataItem,
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

  $: ({ instanceId } = $runtime);

  $: columnsQuery = createQueryServiceTableColumns(instanceId, table, {
    connector,
    database,
    databaseSchema,
  });
  $: ({
    data: columnsData,
    isLoading: columnsIsLoading,
    error: columnsError,
  } = $columnsQuery);

  $: rowsQuery = createQueryServiceTableRows(instanceId, table, {
    connector,
    database,
    databaseSchema,
    limit,
  });
  $: ({
    data: rowsData,
    isLoading: rowsIsLoading,
    error: rowsError,
  } = $rowsQuery);

  $: columns =
    (columnsData?.profileColumns as VirtualizedTableColumns[]) ?? columns; // Retain old profileColumns
  $: rows = rowsData?.data ?? rows; // Retain old rows
</script>

{#if loading || rowsIsLoading || columnsIsLoading}
  <ReconcilingSpinner />
{:else if rowsError || columnsError}
  <WorkspaceError
    message={`Error loading table: ${rowsError?.response.data?.message || columnsError?.response.data?.message}`}
  />
{:else if rows && columns}
  <PreviewTable {rows} columnNames={columns} name={table} />
{/if}
