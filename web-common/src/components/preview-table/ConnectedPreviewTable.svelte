<script lang="ts">
  import PreviewTable from "@rilldata/web-common/components/preview-table/PreviewTable.svelte";
  import ReconcilingSpinner from "@rilldata/web-common/features/entity-management/ReconcilingSpinner.svelte";
  import {
    createQueryServiceTableColumns,
    createQueryServiceTableRows,
  } from "@rilldata/web-common/runtime-client";
  import { onMount } from "svelte";
  import { runtime } from "../../runtime-client/runtime-store";

  export let connector: string;
  export let database: string | undefined;
  export let databaseSchema: string | undefined;
  export let table: string;
  export let limit = 150;
  export let loading = false;

  $: instanceId = $runtime.instanceId;

  $: profileColumnsQuery =
    table === undefined
      ? undefined
      : createQueryServiceTableColumns(instanceId, table, {
          connector: connector,
          database: database,
          databaseSchema: databaseSchema,
        });

  $: profileColumns =
    profileColumnsQuery === undefined
      ? undefined
      : $profileColumnsQuery?.data?.profileColumns ?? profileColumns; // Retain old profileColumns

  $: tableQuery =
    table === undefined
      ? undefined
      : createQueryServiceTableRows(instanceId, table, {
          connector: connector,
          database: database,
          databaseSchema: databaseSchema,
          limit,
        });

  $: rows = $tableQuery?.data?.data ?? rows; // Retain old rows

  /** We will set the overscan amounts to 0 for initial render;
   * in practice, this will shave off around 200ms from the initial render.
   * Then, after 1 second, we will set the overscan amounts to 40 and 10,
   * which wil then cause the table to render with the overscan amounts.
   */
  let rowOverscanAmount = 0;
  let columnOverscanAmount = 0;
  onMount(() => {
    setTimeout(() => {
      rowOverscanAmount = 40;
      columnOverscanAmount = 10;
    }, 1000);
  });
</script>

{#if loading}
  <ReconcilingSpinner />
{:else if rows && profileColumns}
  <PreviewTable
    {rows}
    columnNames={profileColumns}
    {rowOverscanAmount}
    {columnOverscanAmount}
  />
{/if}
